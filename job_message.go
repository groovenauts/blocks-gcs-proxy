package main

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"sync"
	"time"

	pubsub "google.golang.org/api/pubsub/v1"
)

type (
	JobSustainerConfig struct {
		Delay    float64 `json:"delay,omitempty"`
		Interval float64 `json:"interval,omitempty"`
	}

	JobMessageStatus uint8

	JobMessage struct {
		sub    string
		raw    *pubsub.ReceivedMessage
		config *JobSustainerConfig
		puller Puller
		status JobMessageStatus
		mux    sync.Mutex
	}
)

const (
	running JobMessageStatus = iota
	done
	acked
)

func (m *JobMessage) Validate() error {
	if m.MessageId() == "" {
		return &InvalidJobError{msg: "no MessageId is given"}
	}
	return nil
}

func (m *JobMessage) MessageId() string {
	return m.raw.Message.MessageId
}

func (m *JobMessage) DownloadFiles() interface{} {
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

	_, err := m.puller.Acknowledge(m.sub, m.raw.AckId)
	if err != nil {
		log.Fatalf("Failed to acknowledge for message: %v cause of %v\n", m.raw, err)
		return err
	}

	log.Printf("JobMessage.Ack m.status: %v\n", m.status)
	m.status = acked
	log.Printf("JobMessage.Ack m.status: %v\n", m.status)

	return nil
}

func (m *JobMessage) Nack() error {
	m.mux.Lock()
	defer m.mux.Unlock()

	_, err := m.puller.ModifyAckDeadline(m.sub, []string{m.raw.AckId}, 0)
	if err != nil {
		log.Fatalf("Failed to send ModifyAckDeadline as a nack for message: %v cause of %v\n", m.raw, err)
		return err
	}

	log.Printf("JobMessage.Nack m.status: %v\n", m.status)
	m.status = done
	log.Printf("JobMessage.Nack m.status: %v\n", m.status)

	return nil
}

func (m *JobMessage) Done() {
	log.Printf("JobMessage.Done m.status: %v\n", m.status)
	if m.status == running {
		m.status = done
	}
	log.Printf("JobMessage.Done m.status: %v\n", m.status)
}

func (m *JobMessage) running() bool {
	return m.status == running
}

func (m *JobMessage) sendMADPeriodically(notification *ProgressNotification) error {
	log.Printf("sendMADPeriodically start\n")
	for {
		nextLimit := time.Now().Add(time.Duration(m.config.Interval) * time.Second)
		err := m.waitAndSendMAD(notification, nextLimit)
		if err != nil {
			log.Printf("sendMADPeriodically err: %v\n", err)
			return err
		}
		if !m.running() {
			log.Printf("sendMADPeriodically return\n")
			return nil
		}
	}
	// return nil
}

func (m *JobMessage) waitAndSendMAD(notification *ProgressNotification, nextLimit time.Time) error {
	log.Printf("waitAndSendMAD starting\n")
	ticker := time.NewTicker(100 * time.Millisecond)
	for now := range ticker.C {
		if !m.running() {
			log.Printf("waitAndSendMAD ticker stopping\n")
			ticker.Stop()
			log.Printf("waitAndSendMAD ticker stopped\n")
			return nil
		}
		if now.After(nextLimit) {
			log.Printf("waitAndSendMAD nextLimit passed\n")
			ticker.Stop()
			log.Printf("waitAndSendMAD ticker stopped\n")
		}
	}

	log.Printf("waitAndSendMAD m.mux locking m: %v\n", m)

	m.mux.Lock()
	defer m.mux.Unlock()

	log.Printf("waitAndSendMAD m.status: %v\n", m.status)

	// Don't send MAD after sending ACK
	if m.status == acked {
		log.Printf("waitAndSendMAD already acked\n")
		return nil
	}

	log.Printf("waitAndSendMAD sending ModifyAckDeadline\n")
	_, err := m.puller.ModifyAckDeadline(m.sub, []string{m.raw.AckId}, int64(m.config.Delay))
	if err != nil {
		log.Printf("waitAndSendMAD ModifyAckDeadline err: \n", err)
		msg := fmt.Sprintf("Failed modifyAckDeadline %v, %v, %v cause of %v\n", m.sub, m.raw.AckId, m.config.Delay, err)
		log.Fatalf(msg)
		notification.notifyProgress(m.MessageId(), WORKING, false, "error", msg)
	}
	return nil
}
