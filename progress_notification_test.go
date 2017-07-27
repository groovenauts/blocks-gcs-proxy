package main

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"

	pubsub "google.golang.org/api/pubsub/v1"

	log "github.com/Sirupsen/logrus"
)

const (
	DummyTopic = "projects/dummy-proj-999/topics/dummy-topic-001"
	DummyHost  = "dummyhost1"
	DummyJobID = "dummyJob001"
)

type PublishInvocation struct {
	Topic   string
	Message *pubsub.PubsubMessage
}

type DummyPublisher struct {
	Invocations []*PublishInvocation
}

func (dp *DummyPublisher) Publish(topic string, msg *pubsub.PubsubMessage) (*pubsub.PublishResponse, error) {
	dp.Invocations = append(dp.Invocations, &PublishInvocation{
		Topic:   topic,
		Message: msg,
	})
	return nil, nil
}

type ExpectedProgressMessage struct {
	step        string
	step_status string
	progress    string
	completed   string
	level       string
	data        string
}

func TestProgressNotificationNotify(t *testing.T) {
	publisher := DummyPublisher{}

	config := ProgressNotificationConfig{
		Topic:    DummyTopic,
		LogLevel: "info",
		Hostname: DummyHost,
	}

	notification := ProgressNotification{
		config:    &config,
		publisher: &publisher,
		logLevel:  log.InfoLevel,
	}

	baseAttrs := map[string]string{
		"msg_id": "1234",
	}

	// Normal Pattern
	publisher.Invocations = []*PublishInvocation{}

	notification.notify(DummyJobID, INITIALIZING, STARTING, baseAttrs)
	notification.notify(DummyJobID, INITIALIZING, SUCCESS, baseAttrs)
	notification.notify(DummyJobID, DOWNLOADING, STARTING, baseAttrs)
	notification.notify(DummyJobID, DOWNLOADING, SUCCESS, baseAttrs)
	notification.notify(DummyJobID, EXECUTING, STARTING, baseAttrs)
	notification.notify(DummyJobID, EXECUTING, SUCCESS, baseAttrs)
	notification.notify(DummyJobID, UPLOADING, STARTING, baseAttrs)
	notification.notify(DummyJobID, UPLOADING, SUCCESS, baseAttrs)
	notification.notify(DummyJobID, CLEANUP, STARTING, baseAttrs)
	notification.notify(DummyJobID, CLEANUP, SUCCESS, baseAttrs)
	notification.notify(DummyJobID, ACKSENDING, STARTING, baseAttrs)
	notification.notify(DummyJobID, ACKSENDING, SUCCESS, baseAttrs)

	assert.Equal(t, len(publisher.Invocations), 2)

	expecteds := []ExpectedProgressMessage{
		{
			step:        "INITIALIZING",
			step_status: "SUCCESS",
			progress:    "1",
			completed:   "false",
			level:       "info",
			data:        "INITIALIZING SUCCESS",
		},
		{
			step:        "ACKSENDING",
			step_status: "SUCCESS",
			progress:    "5",
			completed:   "true",
			level:       "info",
			data:        "ACKSENDING SUCCESS",
		},
	}
	for idx, invocation := range publisher.Invocations {
		expected := expecteds[idx]
		assert.Equal(t, invocation.Topic, DummyTopic)
		raw_msg, err := base64.StdEncoding.DecodeString(invocation.Message.Data)
		assert.NoError(t, err)
		assert.Equal(t, string(raw_msg), expected.data)
		attrs := invocation.Message.Attributes
		assert.Equal(t, expected.step, attrs["step"])
		assert.Equal(t, expected.step_status, attrs["step_status"])
		assert.Equal(t, expected.progress, attrs["progress"])
		assert.Equal(t, expected.completed, attrs["completed"])
		assert.Equal(t, expected.level, attrs["level"])
		for k, v := range baseAttrs {
			assert.Equal(t, v, attrs[k])
		}
	}

	// Executing failure pattern

	publisher.Invocations = []*PublishInvocation{}

	notification.notify(DummyJobID, INITIALIZING, STARTING, baseAttrs)
	notification.notify(DummyJobID, INITIALIZING, SUCCESS, baseAttrs)
	notification.notify(DummyJobID, DOWNLOADING, STARTING, baseAttrs)
	notification.notify(DummyJobID, DOWNLOADING, SUCCESS, baseAttrs)
	notification.notify(DummyJobID, EXECUTING, STARTING, baseAttrs)
	notification.notify(DummyJobID, EXECUTING, FAILURE, baseAttrs)
	notification.notify(DummyJobID, CLEANUP, STARTING, baseAttrs)
	notification.notify(DummyJobID, CLEANUP, SUCCESS, baseAttrs)
	notification.notify(DummyJobID, NACKSENDING, STARTING, baseAttrs)
	notification.notify(DummyJobID, NACKSENDING, SUCCESS, baseAttrs)

	assert.Equal(t, len(publisher.Invocations), 3)

	expecteds = []ExpectedProgressMessage{
		{
			step:        "INITIALIZING",
			step_status: "SUCCESS",
			progress:    "1",
			completed:   "false",
			level:       "info",
			data:        "INITIALIZING SUCCESS",
		},
		{
			step:        "EXECUTING",
			step_status: "FAILURE",
			progress:    "2",
			completed:   "false",
			level:       "error",
			data:        "EXECUTING FAILURE",
		},
		{
			step:        "NACKSENDING",
			step_status: "SUCCESS",
			progress:    "3",
			completed:   "false",
			level:       "warning",
			data:        "NACKSENDING SUCCESS",
		},
	}
	for idx, invocation := range publisher.Invocations {
		expected := expecteds[idx]
		assert.Equal(t, invocation.Topic, DummyTopic)
		raw_msg, err := base64.StdEncoding.DecodeString(invocation.Message.Data)
		assert.NoError(t, err)
		assert.Equal(t, string(raw_msg), expected.data)
		attrs := invocation.Message.Attributes
		assert.Equal(t, DummyHost, attrs["host"])
		assert.Equal(t, expected.step, attrs["step"])
		assert.Equal(t, expected.step_status, attrs["step_status"])
		assert.Equal(t, expected.progress, attrs["progress"])
		assert.Equal(t, expected.completed, attrs["completed"])
		assert.Equal(t, expected.level, attrs["level"])
		for k, v := range baseAttrs {
			assert.Equal(t, v, attrs[k])
		}
	}
}
