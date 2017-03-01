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
	CANCELLING     = 2
	CANCELL_OK     = 3
	CANCELL_ERROR  = 4
	PREPARING      = 5
	PREPARE_OK     = 6
	PREPARE_ERROR  = 7
	DOWNLOADING    = 8
	DOWNLOAD_OK    = 9
	DOWNLOAD_ERROR = 10
	EXECUTING      = 11
	EXECUTE_OK     = 12
	EXECUTE_ERROR  = 13
	UPLOADING      = 14
	UPLOAD_OK      = 15
	UPLOAD_ERROR   = 16
	ACKSENDING     = 17
	ACKSEND_OK     = 18
	ACKSEND_ERROR  = 19
	CLEANUP        = 20

	COMPLETED = ACKSEND_OK
	TOTAL     = CLEANUP
)

var PROGRESS_MESSAFGES = map[int]string{
	PROCESSING:     "PROCESSING",
	CANCELLING:     "CANCELLING",
	CANCELL_OK:     "CANCELL_OK",
	CANCELL_ERROR:  "CANCELL_ERROR",
	PREPARING:      "PREPARING",
	PREPARE_OK:     "PREPARE_OK",
	PREPARE_ERROR:  "PREPARE_ERROR",
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
