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
	CANCELLING = 1 + iota
	CANCELL_OK
	CANCELL_ERROR
	PREPARING
	PREPARE_OK
	PREPARE_ERROR
	DOWNLOADING
	DOWNLOAD_OK
	DOWNLOAD_ERROR
	EXECUTING
	EXECUTE_OK
	EXECUTE_ERROR
	UPLOADING
	UPLOAD_OK
	UPLOAD_ERROR
	ACKSENDING
	ACKSEND_OK
	ACKSEND_ERROR
	NACKSENDING
	NACKSEND_OK
	NACKSEND_ERROR
	CLEANUP
	CLEANUP_OK
	CLEANUP_ERROR

	COMPLETED = ACKSEND_OK
	TOTAL     = CLEANUP
)

var PROGRESS_MESSAFGES = map[int]string{
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
	NACKSENDING:    "NACKSENDING",
	NACKSEND_OK:		"NACKSEND_OK",
	NACKSEND_ERROR: "NACKSEND_ERROR",
	CLEANUP:        "CLEANUP",
	CLEANUP_OK:     "CLEANUP_OK",
	CLEANUP_ERROR:  "CLEANUP_ERROR",
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
