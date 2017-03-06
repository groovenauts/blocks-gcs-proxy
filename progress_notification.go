package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"strconv"

	// "golang.org/x/net/context"

	pubsub "google.golang.org/api/pubsub/v1"
)

type (
	Publisher interface {
		Publish(topic string, msg *pubsub.PubsubMessage) (*pubsub.PublishResponse, error)
	}

	pubsubPublisher struct {
		topicsService *pubsub.ProjectsTopicsService
	}
)

func (pp *pubsubPublisher) Publish(topic string, msg *pubsub.PubsubMessage) (*pubsub.PublishResponse, error) {
	req := &pubsub.PublishRequest{
		Messages: []*pubsub.PubsubMessage{msg},
	}
	return pp.topicsService.Publish(topic, req).Do()
}

type JobStepStatus int

const (
	STARTING = 1 + iota
	SUCCESS
	FAILURE
)

func (jss JobStepStatus) String() string {
	switch jss {
	case STARTING: return "STARTING"
	case SUCCESS: return "SUCCESS"
	case FAILURE: return "FAILURE"
	default: return "Unknown"
	}
}

type JobStep int

const (
	PREPARING   = 1 + iota
	DOWNLOADING
	EXECUTING
	UPLOADING
	CANCELLING
	ACKSENDING
	NACKSENDING
	CLEANUP
)

var (
	JOB_STEP_DEFS = map[JobStep][]string {
		PREPARING:    []string{"PREPARING"	, "info" , "error"},
		DOWNLOADING:	[]string{"DOWNLOADING", "info" , "error"},
		EXECUTING:		[]string{"EXECUTING"	,	"info" , "error"},
		UPLOADING:		[]string{"UPLOADING"	,	"info" , "error"},
		CANCELLING:		[]string{"CANCELLING" , "info" , "fatal"},
		ACKSENDING:		[]string{"ACKSENDING" , "info" , "error"},
		NACKSENDING:	[]string{"NACKSENDING", "info" , "warn" },
		CLEANUP:			[]string{"CLEANUP"		, "debug", "warn" },
	}
)

func (js JobStep) String() string {
	return JOB_STEP_DEFS[js][0]
}
func (js JobStep) successLogLevel() string {
	return JOB_STEP_DEFS[js][1]
}
func (js JobStep) failureLogLevel() string {
	return JOB_STEP_DEFS[js][2]
}

func (js JobStep) completed(st JobStepStatus) bool {
	return (js == ACKSENDING) && (st == SUCCESS)
}
func (js JobStep) logLevelFor(st JobStepStatus) string {
	switch st {
	// case STARTING: return "info"
	case SUCCESS: return js.successLogLevel()
	case FAILURE: return js.failureLogLevel()
	default: return "info"
	}
}


type (
	ProgressConfig struct {
		Topic string
	}

	ProgressNotification struct {
		config    *ProgressConfig
		publisher Publisher
	}
)

func (pn *ProgressNotification) wrap(msg_id string, step JobStep, f func() error) func() error {
	return func() error {
		pn.notify(msg_id, step, STARTING)
		err := f()
		if err != nil {
			pn.notifyWithMessage(msg_id, step, FAILURE, err.Error())
			return err
		}
		pn.notify(msg_id, step, SUCCESS)
		return nil
	}
}

func (pn *ProgressNotification) notify(job_msg_id string, step JobStep, st JobStepStatus) error {
	msg := fmt.Sprintf("%v %v", step, st)
	return pn.notifyWithMessage(job_msg_id, step, st, msg)
}

func (pn *ProgressNotification) notifyWithMessage(job_msg_id string, step JobStep, st JobStepStatus, msg string) error {
	log.Printf("Notify %v: %v %v\n", job_msg_id, step, msg)
	opts := map[string]string{
		"progress":       strconv.Itoa(int(step)),
		"completed":      strconv.FormatBool(step.completed(st)),
		"job_message_id": job_msg_id,
		"level":          step.logLevelFor(st),
	}
	m := &pubsub.PubsubMessage{Data: base64.StdEncoding.EncodeToString([]byte(msg)), Attributes: opts}
	_, err := pn.publisher.Publish(pn.config.Topic, m)
	if err != nil {
		log.Printf("Error to publish notification to %v msg: %v cause of %v\n", pn.config.Topic, m, err)
		return err
	}
	return nil
}
