package main

import (
	"encoding/base64"
	"encoding/json"

	pubsub "google.golang.org/api/pubsub/v1"

	log "github.com/Sirupsen/logrus"
)

type Message struct {
	Topic string `json:"topic"`
	Data string `json:"data,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

type Worker struct {
	service *pubsub.Service
	lines chan string
	done    bool
	error   error
}

func (w *Worker) run() {
	for {
		flds := log.Fields{}
		log.Debugln("Getting a target")
		var line string
		select {
		case line = <-w.lines:
		default: // Do nothing to break
		}
		if line == "" {
			log.Debugln("No target found any more")
			w.done = true
			w.error = nil
			break
		}

		flds["line"] = line
		log.WithFields(flds).Debugln("Job Start")

		err := w.process(line)
		flds["error"] = err
		if err != nil {
			w.done = true
			w.error = err
			break
		}
		log.WithFields(flds).Debugln("Job Finished")
	}
}

func (w *Worker) process(line string) error {
	flds := log.Fields{"line": line}
	log.WithFields(flds).Debugln("Processing line")

	var msg Message
	err := json.Unmarshal([]byte(line), &msg)
	if err != nil {
		flds := log.Fields{"error": err, "line": line}
		log.WithFields(flds).Errorln("JSON parse error")
		return err
	}

	flds["message"] = msg
	log.WithFields(flds).Debugln("Publishing message")

	topic := w.service.Projects.Topics
	call := topic.Publish(msg.Topic, &pubsub.PublishRequest{
		Messages: []*pubsub.PubsubMessage{
			&pubsub.PubsubMessage{
				Attributes: msg.Attributes,
				Data: base64.StdEncoding.EncodeToString([]byte(msg.Data)),
			},
		},
	})

	res, err := call.Do()
	if err != nil {
		flds["attributes"] = msg.Attributes
		flds["data"] = msg.Data
		flds["error"] = err
		log.WithFields(flds).Errorln("Publish error")
		return err
	}

	flds["MessageIds"] = res.MessageIds
	log.WithFields(flds).Infoln("Publish successfully")
	
	return nil
}
