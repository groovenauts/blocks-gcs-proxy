package main

import (
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"

	errorReporting "google.golang.org/api/clouderrorreporting/v1beta1"
	logging "google.golang.org/api/logging/v2beta1"
	pubsub "google.golang.org/api/pubsub/v1"
	storage "google.golang.org/api/storage/v1"

	"github.com/cenkalti/backoff"
	logrus "github.com/sirupsen/logrus"
)

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
	client, err := google.DefaultClient(ctx, pubsub.PubsubScope, storage.DevstorageReadWriteScope, logging.LoggingWriteScope, errorReporting.CloudPlatformScope)

	if err != nil {
		log.Fatalln("Failed to create DefaultClient")
		return err
	}

	// Add stackdriver logging
	if p.config.Log.Stackdriver != nil {
		err = p.config.Log.Stackdriver.setupSdHook(client)
		if err != nil {
			logAttrs := logrus.Fields{"client": client, "error": err}
			log.WithFields(logAttrs).Fatalln("Failed to create storage.Service")
			return err
		}
	}

	// Create a storageService
	storageService, err := storage.New(client)
	if err != nil {
		logAttrs := logrus.Fields{"client": client, "error": err}
		log.WithFields(logAttrs).Fatalln("Failed to create storage.Service")
		return err
	}
	p.storage = &CloudStorage{
		service:          storageService.Objects,
		ContentTypeByExt: p.config.Upload.ContentTypeByExt,
	}

	// Creates a pubsubService
	pubsubService, err := pubsub.New(client)
	if err != nil {
		logAttrs := logrus.Fields{"client": client, "error": err}
		log.WithFields(logAttrs).Fatalln("Failed to create pubsub.Service")
		return err
	}

	eb := backoff.NewExponentialBackOff()
	eb.InitialInterval = 10 * time.Second
	b := backoff.WithMaxRetries(eb, 5)
	puller := &BackoffPuller{
		Impl:    &pubsubPuller{pubsubService.Projects.Subscriptions},
		Backoff: b,
	}

	err = p.config.Job.setupSustainer(puller)
	if err != nil {
		logAttrs := logrus.Fields{"client": client, "error": err}
		log.WithFields(logAttrs).Fatalln("Failed to setup sustainer")
		return err
	}

	p.subscription = &JobSubscription{
		config: p.config.Job,
		puller: puller,
	}

	p.config.Progress.setup()
	level, err := logrus.ParseLevel(p.config.Progress.LogLevel)
	if err != nil {
		logAttrs := logrus.Fields{"log_level": p.config.Progress.LogLevel}
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
		logrus.Fields{
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
			config:               p.config.Command,
			commandSeverityLevel: p.config.Log.commandSeverityLevel,
			downloadConfig:       p.config.Download,
			uploadConfig:         p.config.Upload,
			message:              msg,
			notification:         p.notification,
			storage:              p.storage,
			NackInterval:         p.config.Job.NackInterval,
			NackOnError:          p.config.Job.NackOnError,
		}
		job.setupExecUUID()
		jobLog := logger.WithFields(logrus.Fields{"exec-uuid": job.execUUID, "message-id": msg.MessageId(), ConcurrentBatchJobIdKey4Log: msg.ConcurrentBatchJobId()})
		err := p.replaceGlobalLog(jobLog, func() error {
			err := job.run()
			if err != nil {
				logAttrs := logrus.Fields{"error": err, "msg": msg}
				log.WithFields(logAttrs).Fatalln("Job Error")
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}
		return nil
	})
	return err
}

func (p *Process) replaceGlobalLog(newLog *logrus.Entry, f func() error) error {
	// `log` is a global variable
	logBackup := log
	log = newLog
	defer func() {
		log = logBackup
	}()
	return f()
}
