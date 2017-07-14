package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"strconv"

	// "golang.org/x/net/context"

	pubsub "google.golang.org/api/pubsub/v1"

	log "github.com/Sirupsen/logrus"
)

type ProgressNotificationConfig struct {
	Topic    string `json:"topic"`
	LogLevel string `json:"log_level"`
	Hostname string `json:"hostname"`
}

func (c *ProgressNotificationConfig) setup() {
	if c.LogLevel == "" {
		c.LogLevel = log.InfoLevel.String()
	}

	if c.Hostname == "" {
		h, err := os.Hostname()
		if err != nil {
			c.Hostname = "Unknown"
		} else {
			c.Hostname = h
		}
	}
}

type ProgressNotification struct {
	config    *ProgressNotificationConfig
	publisher Publisher
	logLevel  log.Level
}

func (pn *ProgressNotification) wrap(msg_id string, step JobStep, f func() error) func() error {
	return func() error {
		pn.notify(msg_id, step, STARTING)
		err := f()
		if err != nil {
			pn.notifyWithMessage(msg_id, step, FAILURE, err.Error())
			return err
		}
		pn.notify(msg_id, step, SUCCESS)
		return nil
	}
}

func (pn *ProgressNotification) notify(job_msg_id string, step JobStep, st JobStepStatus) error {
	msg := fmt.Sprintf("%v %v", step, st)
	return pn.notifyWithMessage(job_msg_id, step, st, msg)
}

func (pn *ProgressNotification) notifyWithMessage(job_msg_id string, step JobStep, st JobStepStatus, msg string) error {
	opts := map[string]string{
		"step":        step.String(),
		"step_status": st.String(),
	}
	return pn.notifyProgressWithOpts(job_msg_id, step.progressFor(st), step.completed(st), step.logLevelFor(st), msg, opts)
}

func (pn *ProgressNotification) notifyProgress(job_msg_id string, progress Progress, completed bool, level log.Level, data string) error {
	return pn.notifyProgressWithOpts(job_msg_id, progress, completed, level, data, map[string]string{})
}

func (pn *ProgressNotification) notifyProgressWithOpts(job_msg_id string, progress Progress, completed bool, level log.Level, data string, opts map[string]string) error {
	// https://godoc.org/github.com/sirupsen/logrus#Level
	// log.InfoLevel < log.DebugLevel => true
	if pn.logLevel < level {
		return nil
	}
	attrs := map[string]string{}
	for k, v := range opts {
		attrs[k] = v
	}
	attrs["progress"] = strconv.Itoa(int(progress))
	attrs["completed"] = strconv.FormatBool(completed)
	attrs["job_message_id"] = job_msg_id
	attrs["level"] = level.String()
	attrs["host"] = pn.config.Hostname
	logAttrs := log.Fields{}
	for k, v := range attrs {
		logAttrs[k] = v
	}
	log.WithFields(logAttrs).Debugln("Publishing notification")
	m := &pubsub.PubsubMessage{Data: base64.StdEncoding.EncodeToString([]byte(data)), Attributes: attrs}
	_, err := pn.publisher.Publish(pn.config.Topic, m)
	if err != nil {
		logAttrs["error"] = err
		log.WithFields(logAttrs).Debugln("Failed to publish notification")
		return err
	}
	return nil
}
