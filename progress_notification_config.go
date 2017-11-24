package main

import (
	"fmt"

	logrus "github.com/sirupsen/logrus"
)

type ProgressNotificationConfig struct {
	Topic      string            `json:"topic"`
	LogLevel   string            `json:"log_level"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

func (c *ProgressNotificationConfig) setup() *ConfigError {
	if c.Topic == "" {
		c.Topic = fmt.Sprintf("projects/%s/topics/%s-progress-topic", GcpProjectId, Pipeline)
	}
	if c.LogLevel == "" {
		c.LogLevel = logrus.InfoLevel.String()
	}
	return nil
}
