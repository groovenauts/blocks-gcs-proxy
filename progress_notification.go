package main

import (
	"encoding/base64"
	"fmt"
	"strconv"

	// "golang.org/x/net/context"

	pubsub "google.golang.org/api/pubsub/v1"

	logrus "github.com/sirupsen/logrus"
)

type ProgressNotificationConfig struct {
	Topic      string            `json:"topic"`
	LogLevel   string            `json:"log_level"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

func (c *ProgressNotificationConfig) setup() *ConfigError {
	if c.LogLevel == "" {
		c.LogLevel = logrus.InfoLevel.String()
	}
	return nil
}

type ProgressNotification struct {
	config    *ProgressNotificationConfig
	publisher Publisher
	logLevel  logrus.Level
}

func (pn *ProgressNotification) wrap(msg_id string, step JobStep, attrs map[string]string, f func() error) func() error {
	return func() error {
		pn.notify(msg_id, step, STARTING, attrs)
		err := f()
		if err != nil {
			pn.notifyWithMessage(msg_id, step, FAILURE, attrs, err.Error())
			return err
		}
		pn.notify(msg_id, step, SUCCESS, attrs)
		return nil
	}
}

func (pn *ProgressNotification) notify(job_msg_id string, step JobStep, st JobStepStatus, attrs map[string]string) error {
	msg := fmt.Sprintf("%v %v", step, st)
	return pn.notifyWithMessage(job_msg_id, step, st, attrs, msg)
}

func (pn *ProgressNotification) notifyWithMessage(job_msg_id string, step JobStep, st JobStepStatus, opts map[string]string, msg string) error {
	attrs := map[string]string{}
	for k, v := range opts {
		attrs[k] = v
	}
	attrs["step"] = step.String()
	attrs["step_status"] = st.String()
	return pn.notifyProgress(job_msg_id, step.progressFor(st), step.completed(st), step.logLevelFor(st), attrs, msg)
}

func (pn *ProgressNotification) notifyProgress(job_msg_id string, progress Progress, completed bool, level logrus.Level, opts map[string]string, data string) error {
	// https://godoc.org/github.com/sirupsen/logrus#Level
	// log.InfoLevel < log.DebugLevel => true
	if pn.logLevel < level {
		return nil
	}
	attrs := map[string]string{}
	pn.mergeMsgAttrs(attrs, pn.config.Attributes)
	pn.mergeMsgAttrs(attrs, opts)
	pn.mergeMsgAttrs(attrs, map[string]string{
		"progress":       strconv.Itoa(int(progress)),
		"completed":      strconv.FormatBool(completed),
		"job_message_id": job_msg_id,
		"level":          level.String(),
	})
	logAttrs := logrus.Fields{}
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

func (pn *ProgressNotification) mergeMsgAttrs(dest, src map[string]string) {
	for k, v := range src {
		buf := []byte(v)
		if len(buf) > 1024 {
			dest[k] = string(buf[0:1024])
		} else {
			dest[k] = v
		}
	}
}
