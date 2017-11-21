package main

import (
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
