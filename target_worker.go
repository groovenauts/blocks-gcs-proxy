package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/cenkalti/backoff"
	log "github.com/sirupsen/logrus"
)

type WorkerConfig struct {
	Workers  int `json:"workers,omitempty"`
	MaxTries int `json:"max_tries,omitempty"`
}

func (c *WorkerConfig) setup() *ConfigError {
	if c.Workers < 1 {
		c.Workers = 1
	}
	return nil
}

type Target struct {
	Bucket    string
	Object    string
	LocalPath string
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

		flds["target"] = t
		log.WithFields(flds).Debugf("Start to %v\n", w.name)

		f := func() error {
			return w.impl(t.Bucket, t.Object, t.LocalPath)
		}

		eb := backoff.NewExponentialBackOff()
		eb.InitialInterval = 30 * time.Second
		b := backoff.WithMaxTries(eb, uint64(w.maxTries))
		err := backoff.Retry(f, b)
		flds["error"] = err
		if err != nil {
			log.WithFields(flds).Errorf("Failed to %v\n", w.name)
			w.done = true
			w.error = err
			break
		}
		log.WithFields(flds).Debugf("Finished to %v\n", w.name)
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
