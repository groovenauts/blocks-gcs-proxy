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
