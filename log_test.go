package main

import (
	"bytes"
	"fmt"
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

func TestLogrusWriter(t *testing.T) {
	var buf bytes.Buffer
	logger := logrus.New()
	logger.Out = &buf
	subject := &LogrusWriter{Dest: logger, Severity: logrus.InfoLevel}
	subject.Setup()

	fmt.Fprintln(subject, "Hello world!")
	s := buf.String()
	assert.Contains(t, s, "level=info")
	assert.Contains(t, s, `msg="Hello world!\n"`)
}
