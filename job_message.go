package main

import (
	"log"
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

func (s *JobMessage) MessageId() string {
	return s.raw.Message.MessageId
}

func (s *JobMessage) Attribute(key string) string {
	return s.raw.Message.Attributes[key]
}


func (s *JobMessage) Ack() error {
	s.mux.Lock()
	defer s.mux.Unlock()

	_, err := s.puller.Acknowledge(s.sub, s.raw.AckId)
	if err != nil {
		log.Fatalf("Failed to acknowledge for message: %v cause of %v\n", s.raw, err)
		return err
	}

	s.status = acked

	return nil
}

func (s *JobMessage) Done() {
	if s.status == running {
		s.status = done
	}
}

func (s *JobMessage) running() bool {
	return s.status == running
}

func (s *JobMessage) sendMADPeriodically() error {
	for {
		nextLimit := time.Now().Add(time.Duration(s.config.Interval) * time.Second)
		err := s.waitAndSendMAD(nextLimit)
		if err != nil {
			return err
		}
		if !s.running() {
			return nil
		}
	}
	// return nil
}

func (s *JobMessage) waitAndSendMAD(nextLimit time.Time) error {
	ticker := time.NewTicker(100 * time.Millisecond)
	for now := range ticker.C {
		if !s.running() {
			ticker.Stop()
			return nil
		}
		if now.After(nextLimit) {
			ticker.Stop()
		}
	}

	s.mux.Lock()
	defer s.mux.Unlock()

	// Don't send MAD after sending ACK
	if s.status == acked {
		return nil
	}

	_, err := s.puller.ModifyAckDeadline(s.sub, []string{s.raw.AckId}, int64(s.config.Delay))
	if err != nil {
		log.Fatalf("Failed modifyAckDeadline %v, %v, %v cause of %v\n", s.sub, s.raw.AckId, s.config.Delay, err)
	}
	return nil
}
