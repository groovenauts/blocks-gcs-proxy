package main

import (
	"fmt"
	logrus "github.com/sirupsen/logrus"
)

type LogConfig struct {
	Level       string         `json:"level,omitempty"`
	CommandSeverity string     `json:"command_severity"`
	commandSeverityLevel logrus.Level
	Stackdriver *LoggingConfig `json:"stackdriver,omitempty"`
}

var logger = logrus.New()
var log = logger.WithFields(logrus.Fields{})

func (c *LogConfig) setup() *ConfigError {
	setups := []ConfigSetup{
		c.setupLevel,
		c.setupCommandSeverity,
		c.setupStackdriver,
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
	level, err := logrus.ParseLevel(c.Level)
	if err != nil {
		log.Warnf("Error on log.ParseLevel level: %q because of %v\n", c.Level, err)
		return &ConfigError{Name: "level", Message: fmt.Sprintf("is invalid because of %v", err)}
	}
	logger.SetLevel(level)
	return nil
}

func (c *LogConfig) setupCommandSeverity() *ConfigError {
	if c.CommandSeverity == "" {
		c.CommandSeverity = "info"
	}
	level, err := logrus.ParseLevel(c.CommandSeverity)
	if err != nil {
		log.Warnf("Error on log.ParseLevel command_severity: %q because of %v\n", c.CommandSeverity, err)
		return &ConfigError{Name: "command_severity", Message: fmt.Sprintf("is invalid because of %v", err)}
	}
	c.commandSeverityLevel = level
	return nil
}

func (c *LogConfig) setupStackdriver() *ConfigError {
	if c.Stackdriver != nil {
		err := c.Stackdriver.setup()
		if err != nil {
			err.Add("stackdriver")
			return err
		}
	}
	return nil
}
