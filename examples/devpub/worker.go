package main

import (
	"encoding/json"

	pubsub "google.golang.org/api/pubsub/v1"

	log "github.com/Sirupsen/logrus"
)

type Message struct {
	Data string
	Attributes map[string]string
}

type Worker struct {
	service *pubsub.Service
	topic string
	lines chan []byte
	done    bool
	error   error
}

func (w *Worker) run() {
	for {
		flds := log.Fields{}
		log.Debugln("Getting a target")
		var line []byte
		select {
		case line = <-w.lines:
		default: // Do nothing to break
		}
		if line == nil {
			log.Debugln("No target found any more")
			w.done = true
			w.error = nil
			break
		}

		flds["line"] = string(line)
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

func (w *Worker) process(line []byte) error {
	var msg Message
	err := json.Unmarshal(line, &msg)
	if err != nil {
		flds := log.Fields{"error": err, "line": string(line)}
		log.WithFields(flds).Errorln("JSON parse error")
		return err
	}

	topic := w.service.Projects.Topics
	call := topic.Publish(w.topic, &pubsub.PublishRequest{
		Messages: []*pubsub.PubsubMessage{
			&pubsub.PubsubMessage{
				Attributes: msg.Attributes,
				Data: msg.Data,
			},
		},
	})
	res, err := call.Do()
	if err != nil {
		flds := log.Fields{"attributes": msg.Attributes, "data": msg.Data}
		log.WithFields(flds).Errorln("Publish error")
		return err
	}
	
	flds := log.Fields{"MessageIds": res.MessageIds}
	log.WithFields(flds).Infoln("Publish successfully")
	
	return nil
}
