package main

import (
	"fmt"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/groovenauts/concurrent-go"
	logrus "github.com/sirupsen/logrus"
)

type WorkerConfig struct {
	Workers  int `json:"workers,omitempty"`
	MaxTries int `json:"max_tries,omitempty"`
}

func (c *WorkerConfig) setup() *ConfigError {
	if c.Workers < 1 {
		c.Workers = 1
	}
	if c.MaxTries < 1 {
		c.MaxTries = 5
	}
	return nil
}

type Target struct {
	Bucket    string
	Object    string
	LocalPath string
}

func (t *Target) URL() string {
	return fmt.Sprintf("gs://%s/%s", t.Bucket, t.Object)
}

type RetryableFunc struct {
	name     string
	maxTries int
	interval time.Duration
}

func (w *RetryableFunc) Wrap(orig func(*concurrent.Job) error) func(*concurrent.Job) error {
	return func(job *concurrent.Job) error {
		f := func() error {
			return orig(job)
		}

		eb := backoff.NewExponentialBackOff()
		eb.InitialInterval = w.interval
		b := backoff.WithMaxRetries(eb, uint64(w.maxTries))
		// err := backoff.Retry(f, b)
		err := RetryWithSleep(f, b)

		return err
	}
}

func (w *RetryableFunc) WithLog(orig func(*concurrent.Job) error) func(*concurrent.Job) error {
	return func(job *concurrent.Job) error {
		t, ok := job.Payload.(*Target)
		if !ok {
			return fmt.Errorf("Unknown Payload: %v\n", job.Payload)
		}

		flds := logrus.Fields{"target": t}
		log.WithFields(flds).Debugf("Worker Start to %v\n", w.name)

		err := orig(job)

		flds["error"] = err
		if err != nil {
			log.WithFields(flds).Errorf("Worker Failed to %v %v\n", w.name, t.URL())
			return err
		}
		log.WithFields(flds).Debugf("Worker Finished to %v\n", w.name)
		return nil
	}
}
