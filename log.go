package main

import (
	log "github.com/Sirupsen/logrus"
)

type LogConfig struct {
	Level string `json:"level,omitempty"`
}

func (c *LogConfig) setup() error {
	if c.Level == "" {
		c.Level = "info"
	}
	level, err := log.ParseLevel(c.Level)
	if err != nil {
		log.Warnf("Invalid log level: %q\n", c.Level)
		return err
	}
	log.SetLevel(level)
	return nil
}
