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
		Log      *LogConfig      `json:"log,omitempty"`
	}
)

func (c *ProcessConfig) setup(args []string) error {
	if c.Command == nil {
		c.Command = &CommandConfig{}
	}
	c.Command.setup()

	c.Command.Template = args
	if c.Job == nil {
		c.Job = &JobConfig{}
	}
	c.Job.setup()

	if c.Progress == nil {
		c.Progress = &ProgressConfig{}
	}
	if c.Log == nil {
		c.Log = &LogConfig{}
	}
	err := c.Log.setup()
	if err != nil {
		return err
	}
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

func (p *Process) setup() error {
	ctx := context.Background()

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

	puller := &pubsubPuller{pubsubService.Projects.Subscriptions}

	if !p.config.Job.Sustainer.Disabled {
	err = p.config.Job.setupSustainer(puller)
	if err != nil {
		logAttrs := log.Fields{"client": client, "error": err}
		log.WithFields(logAttrs).Fatalln("Failed to setup sustainer")
		return err
	}
	}

	p.subscription = &JobSubscription{
		config: p.config.Job,
		puller: puller,
	}

	p.config.Progress.setup()
	level, err := log.ParseLevel(p.config.Progress.LogLevel)
	if err != nil {
		logAttrs := log.Fields{"log_level": p.config.Progress.LogLevel}
		log.WithFields(logAttrs).Fatalln("Failed to parse log_level")
		return err
	}
	p.notification = &ProgressNotification{
		config:    p.config.Progress,
		publisher: &pubsubPublisher{pubsubService.Projects.Topics},
		logLevel:  level,
	}
	return nil
}

func (p *Process) run() error {
	logAttrs :=
		log.Fields{
			"VERSION": VERSION,
			"config": map[string]interface{}{
				"command":  p.config.Command,
				"job":      p.config.Job,
				"progress": p.config.Progress,
				"log":      p.config.Log,
			},
		}
	log.WithFields(logAttrs).Infoln("Start listening")
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
			log.WithFields(logAttrs).Fatalln("Job Error")
			return err
		}
		return nil
	})
	return err
}
