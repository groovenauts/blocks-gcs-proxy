package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
	"text/template"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"

	pubsub "google.golang.org/api/pubsub/v1"
	storage "google.golang.org/api/storage/v1"

	log "github.com/Sirupsen/logrus"
)

type (
	ProcessConfig struct {
		Command  *CommandConfig  `json:"command,omitempty"`
		Job      *JobConfig      `json:"job,omitempty"`
		Progress *ProgressConfig `json:"progress,omitempty"`
	}
)

func (c *ProcessConfig) setup(args []string) error {
	if c.Command == nil {
		c.Command = &CommandConfig{}
	}
	c.Command.Template = args
	return nil
}

func LoadProcessConfig(path string) (*ProcessConfig, error) {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	funcMap := template.FuncMap{"env": os.Getenv}
	t, err := template.New("config").Funcs(funcMap).Parse(string(raw))
	if err != nil {
		return nil, err
	}

	env := map[string]string{}
	for _, s := range os.Environ() {
		parts := strings.SplitN(s, "=", 2)
		env[parts[0]] = parts[1]
	}

	buf := new(bytes.Buffer)
	t.Execute(buf, env)

	var res ProcessConfig
	err = json.Unmarshal(buf.Bytes(), &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

type (
	Process struct {
		config       *ProcessConfig
		subscription *JobSubscription
		notification *ProgressNotification
		storage      *CloudStorage
	}
)

func (p *Process) setup(ctx context.Context) error {
	// https://github.com/google/google-api-go-client#application-default-credentials-example
	client, err := google.DefaultClient(ctx, pubsub.PubsubScope, storage.DevstorageReadWriteScope)

	if err != nil {
		log.Fatalln("Failed to create DefaultClient")
		return err
	}

	// Create a storageService
	storageService, err := storage.New(client)
	if err != nil {
		logAttrs := log.Fields{"client": client, "error": err}
		log.WithFields(logAttrs).Fatalln("Failed to create storage.Service")
		return err
	}
	p.storage = &CloudStorage{storageService.Objects}

	// Creates a pubsubService
	pubsubService, err := pubsub.New(client)
	if err != nil {
		logAttrs := log.Fields{"client": client, "error": err}
		log.WithFields(logAttrs).Fatalln("Failed to create pubsub.Service")
		return err
	}

	p.subscription = &JobSubscription{
		config: p.config.Job,
		puller: &pubsubPuller{pubsubService.Projects.Subscriptions},
	}
	p.notification = &ProgressNotification{
		config:    p.config.Progress,
		publisher: &pubsubPublisher{pubsubService.Projects.Topics},
	}
	return nil
}

func (p *Process) run() error {
	err := p.subscription.listen(func(msg *JobMessage) error {
		job := &Job{
			config:       p.config.Command,
			message:      msg,
			notification: p.notification,
			storage:      p.storage,
		}
		err := job.run()
		if err != nil {
			logAttrs := log.Fields{"error": err, "msg": msg}
			log.WithFields(logAttrs).Fatalln("Kpbg Error")
			return err
		}
		return nil
	})
	return err
}
