package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sync"
	"time"

	pubsub "google.golang.org/api/pubsub/v1"

	logrus "github.com/sirupsen/logrus"
)

type JobMessageStatus uint8

const (
	running JobMessageStatus = iota
	done
	acked
)

func (s JobMessageStatus) String() string {
	switch s {
	case running:
		return "running"
	case done:
		return "done"
	case acked:
		return "acked"
	default:
		return fmt.Sprintf("JobMessageStatus[Unknown:%v]", uint8(s))
	}
}

type (
	JobSustainerConfig struct {
		Disabled bool    `json:"disabled,omitempty"`
		Delay    float64 `json:"delay,omitempty"`
		Interval float64 `json:"interval,omitempty"`
	}

	JobMessage struct {
		sub    string
		raw    *pubsub.ReceivedMessage
		config *JobSustainerConfig
		puller Puller
		status JobMessageStatus
		mux    sync.Mutex
	}
)

const ExecUUIDKey = "concurrent-batch.exec-uuid"

func (m *JobMessage) Validate() error {
	if m.MessageId() == "" {
		return &InvalidJobError{msg: "no MessageId is given"}
	}
	return nil
}

func (m *JobMessage) MessageId() string {
	return m.raw.Message.MessageId
}

func (m *JobMessage) InsertExecUUID() {
}

func (m *JobMessage) DownloadFiles() interface{} {
	// See Cloud Pub/Sub Notifications for Google Cloud Storage
	// https://cloud.google.com/storage/docs/pubsub-notifications
	eventType, ok1 := m.raw.Message.Attributes["eventType"]
	bucketId, ok2 := m.raw.Message.Attributes["bucketId"]
	objectId, ok3 := m.raw.Message.Attributes["objectId"]
	if ok1 && ok2 && ok3 {
		if eventType != "OBJECT_FINALIZE" {
			return nil
		}
		url := "gs://" + bucketId + "/" + objectId
		return []interface{}{url}
	}

	str, ok := m.raw.Message.Attributes["download_files"]
	if !ok {
		return nil
	}
	return m.parseJson(str)
}

func (m *JobMessage) parseJson(str string) interface{} {
	matched, err := regexp.MatchString(`\A\[.*\]\z|\A\{.*\}\z|`, str)
	if err != nil {
		return str
	}
	if !matched {
		return str
	}
	var dest interface{}
	err = json.Unmarshal([]byte(str), &dest)
	if err != nil {
		return str
	}
	return dest
}

func (m *JobMessage) Ack() error {
	m.mux.Lock()
	defer m.mux.Unlock()

	logAttrs := logrus.Fields{"job_message_id": m.MessageId(), "ack_id": m.raw.AckId}
	log.WithFields(logAttrs).Debugln("Sending ACK")

	_, err := m.puller.Acknowledge(m.sub, m.raw.AckId)
	if err != nil {
		logAttrs["raw"] = fmt.Sprintf("%v", m.raw)
		logAttrs["error"] = err
		log.WithFields(logAttrs).Errorln("Failed to acknowledge")
		return err
	}

	logAttrs["status"] = m.status
	log.WithFields(logAttrs).Debugln("Updating status to acked")

	m.status = acked
	return nil
}

func (m *JobMessage) Nack() error {
	m.mux.Lock()
	defer m.mux.Unlock()

	logAttrs := logrus.Fields{"job_message_id": m.MessageId(), "ack_id": m.raw.AckId}
	log.WithFields(logAttrs).Debugln("Sending NACK")

	_, err := m.puller.ModifyAckDeadline(m.sub, []string{m.raw.AckId}, 0)
	if err != nil {
		logAttrs["raw"] = fmt.Sprintf("%v", m.raw)
		logAttrs["error"] = err
		log.WithFields(logAttrs).Errorln("Failed to send ModifyAckDeadline as a NACK")
		return err
	}

	logAttrs["status"] = m.status
	log.WithFields(logAttrs).Debugln("Updating status to done")

	m.status = done
	return nil
}

func (m *JobMessage) Done() {
	logAttrs := logrus.Fields{"job_message_id": m.MessageId(), "status": m.status}
	log.WithFields(logAttrs).Debugln("Done()")
	if m.status == running {
		m.status = done
	}
}

func (m *JobMessage) running() bool {
	return m.status == running
}

func (m *JobMessage) sendMADPeriodically(notification *ProgressNotification) error {
	if m.config.Disabled {
		return nil
	}
	log.Printf("sendMADPeriodically start\n")
	for {
		nextLimit := time.Now().Add(time.Duration(m.config.Interval) * time.Second)
		err := m.waitAndSendMAD(notification, nextLimit)
		if err != nil {
			log.WithFields(logrus.Fields{"error": err}).Errorln("Error in sendMADPeriodically")
			return err
		}
		if !m.running() {
			log.Debugln("sendMADPeriodically return")
			return nil
		}
	}
	// return nil
}

func (m *JobMessage) waitAndSendMAD(notification *ProgressNotification, nextLimit time.Time) error {
	log.Debugln("waitAndSendMAD starting")
	ticker := time.NewTicker(100 * time.Millisecond)
	for now := range ticker.C {
		if !m.running() {
			log.Debugln("waitAndSendMAD ticker stopping")
			ticker.Stop()
			log.Debugln("waitAndSendMAD ticker stopped")
			return nil
		}
		if now.After(nextLimit) {
			log.Debugln("waitAndSendMAD nextLimit passed")
			ticker.Stop()
			log.Debugln("waitAndSendMAD ticker stopped")
			break
		}
	}

	log.Debugln("waitAndSendMAD m.mux locking")

	m.mux.Lock()
	defer m.mux.Unlock()

	logAttrs := logrus.Fields{"status": m.status}
	log.WithFields(logAttrs).Debugln("waitAndSendMAD")

	// Don't send MAD after sending ACK
	if m.status == acked {
		log.WithFields(logAttrs).Infoln("waitAndSendMAD already acked")
		return nil
	}

	log.WithFields(logAttrs).Debugln("waitAndSendMAD sending ModifyAckDeadline")
	_, err := m.puller.ModifyAckDeadline(m.sub, []string{m.raw.AckId}, int64(m.config.Delay))
	if err != nil {
		logAttrs["error"] = err
		logAttrs["subscription"] = m.sub
		logAttrs["AckId"] = m.raw.AckId
		logAttrs["Delay"] = m.config.Delay
		log.WithFields(logAttrs).Errorln("waitAndSendMAD ModifyAckDeadline")
		msg := fmt.Sprintf("Failed modifyAckDeadline %v, %v, %v cause of %v\n", m.sub, m.raw.AckId, m.config.Delay, err)
		log.WithFields(logAttrs).Fatalf(msg)
		notification.notifyProgress(m.MessageId(), WORKING, false, log.ErrorLevel, m.raw.Message.Attributes, msg)
	}
	return nil
}
