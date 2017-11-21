package main

import (
	"fmt"
	"os"

	logrus "github.com/sirupsen/logrus"
)

type ProgressNotificationConfig struct {
	Topic      string            `json:"topic"`
	LogLevel   string            `json:"log_level"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

func (c *ProgressNotificationConfig) setup() *ConfigError {
	if c.Topic == "" {
		c.Topic = fmt.Sprintf("projects/%s/topics/%s-progress-topic", os.Getenv("GCP_PROJECT"), os.Getenv("PIPELINE"))
	}
	if c.LogLevel == "" {
		c.LogLevel = logrus.InfoLevel.String()
	}
	return nil
}
