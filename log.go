package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
)

type LogConfig struct {
	Level       string         `json:"level,omitempty"`
	Stackdriver *LoggingConfig `json:"stackdriver,omitempty"`
}

func (c *LogConfig) setup() *ConfigError {
	setups := []ConfigSetup{
		c.setupLevel,
	}
	for _, setup := range setups {
		err := setup()
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *LogConfig) setupLevel() *ConfigError {
	if c.Level == "" {
		c.Level = "info"
	}
	level, err := log.ParseLevel(c.Level)
	if err != nil {
		log.Warnf("Error on log.ParseLevel level: %q because of %v\n", c.Level, err)
		return &ConfigError{Name: "level", Message: fmt.Sprintf("is invalid because of %v", err)}
	}
	log.SetLevel(level)
	return nil
}
