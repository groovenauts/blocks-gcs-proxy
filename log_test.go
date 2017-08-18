package main

import (
	"testing"

	"github.com/stretchr/testify/assert"

	logrus "github.com/sirupsen/logrus"
)

func init() {
	logger.SetLevel(logrus.PanicLevel)
}

func TestLogConfig(t *testing.T) {
	backup := logrus.GetLevel()
	defer func() {
		logger.SetLevel(backup)
	}()

	c1 := &LogConfig{Level: "debug"}
	c1.setup()
	assert.Equal(t, logrus.DebugLevel, logger.Level)

	c2 := &LogConfig{Level: "warn"}
	c2.setup()
	assert.Equal(t, logrus.WarnLevel, logger.Level)
}
