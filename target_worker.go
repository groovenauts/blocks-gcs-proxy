package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/cenkalti/backoff"
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
	Error     error
}

func (t *Target) URL() string {
	return fmt.Sprintf("gs://%s/%s", t.Bucket, t.Object)
}

type Targets []*Target

func (targets Targets) error() error {
	messages := []string{}
	for _, t := range targets {
		if t.Error != nil {
			messages = append(messages, t.Error.Error())
		}
	}
	if len(messages) == 0 {
		return nil
	}
	return fmt.Errorf(strings.Join(messages, "\n"))
}

type TargetWorker struct {
	name     string
	targets  chan *Target
	impl     func(bucket, object, srcPath string) error
	done     bool
	maxTries int
	interval time.Duration
}

func (w *TargetWorker) run() {
	for {
		flds := logrus.Fields{}
		log.Debugln("Getting a target")
		var t *Target
		select {
		case t = <-w.targets:
		default: // Do nothing to break
		}
		if t == nil {
			log.Debugln("No target found any more")
			w.done = true
			break
		}
		if t.Error != nil {
			continue
		}

		flds["target"] = t
		log.WithFields(flds).Debugf("Worker Start to %v\n", w.name)

		f := func() error {
			return w.impl(t.Bucket, t.Object, t.LocalPath)
		}

		eb := backoff.NewExponentialBackOff()
		eb.InitialInterval = w.interval
		b := backoff.WithMaxTries(eb, uint64(w.maxTries))
		// err := backoff.Retry(f, b)
		err := RetryWithSleep(f, b)
		flds["error"] = err
		if err != nil {
			log.WithFields(flds).Errorf("Worker Failed to %v %v\n", w.name, t.URL())
			t.Error = err
			continue
		}
		log.WithFields(flds).Debugf("Worker Finished to %v\n", w.name)
	}
}

type TargetWorkers []*TargetWorker

func (ws TargetWorkers) process(targets Targets) error {
	c := make(chan *Target, len(targets))
	for _, t := range targets {
		c <- t
	}

	for _, w := range ws {
		w.done = false
		w.targets = c
		go w.run()
	}

	for {
		time.Sleep(100 * time.Millisecond)
		if ws.done() {
			break
		}
	}

	return targets.error()
}

func (ws TargetWorkers) done() bool {
	for _, w := range ws {
		if !w.done {
			return false
		}
	}
	return true
}
