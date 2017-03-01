package main

import (
	"encoding/base64"
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

const (
	PROCESSING     = 1
	DOWNLOADING    = 2
	DOWNLOAD_OK    = 3
	DOWNLOAD_ERROR = 4
	EXECUTING      = 5
	EXECUTE_OK     = 6
	EXECUTE_ERROR  = 7
	UPLOADING      = 8
	UPLOAD_OK      = 9
	UPLOAD_ERROR   = 10
	ACKSENDING     = 11
	ACKSEND_OK     = 12
	ACKSEND_ERROR  = 13
	CLEANUP        = 14

	COMPLETED = ACKSEND_OK
	TOTAL     = CLEANUP
)

var PROGRESS_MESSAFGES = map[int]string{
	PROCESSING:     "PROCESSING",
	DOWNLOADING:    "DOWNLOADING",
	DOWNLOAD_OK:    "DOWNLOAD_OK",
	DOWNLOAD_ERROR: "DOWNLOAD_ERROR",
	EXECUTING:      "EXECUTING",
	EXECUTE_OK:     "EXECUTE_OK",
	EXECUTE_ERROR:  "EXECUTE_ERROR",
	UPLOADING:      "UPLOADING",
	UPLOAD_OK:      "UPLOAD_OK",
	UPLOAD_ERROR:   "UPLOAD_ERROR",
	ACKSENDING:     "ACKSENDING",
	ACKSEND_OK:     "ACKSEND_OK",
	ACKSEND_ERROR:  "ACKSEND_ERROR",
	CLEANUP:        "CLEANUP",
}

func (pn *ProgressNotification) notify(progress int, job_msg_id, level string) error {
	msg := PROGRESS_MESSAFGES[progress]
	log.Printf("Notify %v/%v %v\n", progress, TOTAL, msg)
	opts := map[string]string{
		"progress":       strconv.Itoa(progress),
		"total":          strconv.Itoa(TOTAL),
		"completed":      strconv.FormatBool(progress == COMPLETED),
		"job_message_id": job_msg_id,
		"level":          level,
	}
	m := &pubsub.PubsubMessage{Data: base64.StdEncoding.EncodeToString([]byte(msg)), Attributes: opts}
	_, err := pn.publisher.Publish(pn.config.Topic, m)
	if err != nil {
		log.Printf("Error to publish notification to %v msg: %v cause of %v\n", pn.config.Topic, m, err)
		return err
	}
	return nil
}
