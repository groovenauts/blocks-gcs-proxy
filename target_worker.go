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
	error     error
}

func (t *Target) URL() string {
	return fmt.Sprintf("gs://%s/%s", t.Bucket, t.Object)
}

type TargetWorker struct {
	name     string
	targets  chan *Target
	impl     func(bucket, object, srcPath string) error
	done     bool
	error    error
	maxTries int
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
			w.error = nil
			break
		}
		if t.error != nil {
			continue
		}

		flds["target"] = t
		log.WithFields(flds).Debugf("Worker Start to %v\n", w.name)

		f := func() error {
			return w.impl(t.Bucket, t.Object, t.LocalPath)
		}

		eb := backoff.NewExponentialBackOff()
		eb.InitialInterval = 30 * time.Second
		b := backoff.WithMaxTries(eb, uint64(w.maxTries))
		err := backoff.Retry(f, b)
		flds["error"] = err
		if err != nil {
			log.WithFields(flds).Errorf("Worker Failed to %v %v\n", w.name, t.URL())
			w.done = true
			w.error = err
			t.error = err
			continue
		}
		log.WithFields(flds).Debugf("Worker Finished to %v\n", w.name)
	}
}

type TargetWorkers []*TargetWorker

func (ws TargetWorkers) process(targets []*Target) error {
	c := make(chan *Target, len(targets))
	for _, t := range targets {
		c <- t
	}

	for _, w := range ws {
		w.targets = c
		go w.run()
	}

	for {
		time.Sleep(100 * time.Millisecond)
		if ws.done() {
			break
		}
	}

	return ws.error()
}

func (ws TargetWorkers) done() bool {
	for _, w := range ws {
		if !w.done {
			return false
		}
	}
	return true
}

func (ws TargetWorkers) error() error {
	messages := []string{}
	for _, w := range ws {
		if w.error != nil {
			messages = append(messages, w.error.Error())
		}
	}
	if len(messages) == 0 {
		return nil
	}
	return fmt.Errorf(strings.Join(messages, "\n"))
}
