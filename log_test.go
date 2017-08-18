package main

import (
	"testing"

	"github.com/stretchr/testify/assert"

	logrus "github.com/sirupsen/logrus"
)

func init() {
	log.SetLevel(log.PanicLevel)
}

func TestLogConfig(t *testing.T) {
	backup := log.GetLevel()
	defer func() {
		log.SetLevel(backup)
	}()

	c1 := &LogConfig{Level: "debug"}
	c1.setup()
	assert.Equal(t, log.DebugLevel, log.GetLevel())

	c2 := &LogConfig{Level: "warn"}
	c2.setup()
	assert.Equal(t, log.WarnLevel, log.GetLevel())
}
