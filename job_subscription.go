package main

import (
	"log"
	"sync"
	"time"

	"golang.org/x/net/context"

	pubsub "google.golang.org/api/pubsub/v1"
)

type (
	Puller interface {
		Pull(subscription string, pullrequest *pubsub.PullRequest) (*pubsub.PullResponse, error)
		Acknowledge(subscription, ackId string) (*pubsub.Empty, error)
		ModifyAckDeadline(subscription string, ackIds []string, ackDeadlineSeconds int64) (*pubsub.Empty, error)
	}

	pubsubPuller struct {
		subscriptionsService *pubsub.ProjectsSubscriptionsService
	}
)

func (pp *pubsubPuller) Pull(subscription string, pullrequest *pubsub.PullRequest) (*pubsub.PullResponse, error) {
	return pp.subscriptionsService.Pull(subscription, pullrequest).Do()
}

func (pp *pubsubPuller) Acknowledge(subscription, ackId string) (*pubsub.Empty, error) {
	ackRequest := &pubsub.AcknowledgeRequest{
		AckIds: []string{ackId},
	}
	return pp.subscriptionsService.Acknowledge(subscription, ackRequest).Do()
}

func (pp *pubsubPuller) ModifyAckDeadline(subscription string, ackIds []string, ackDeadlineSeconds int64) (*pubsub.Empty, error) {
	req := &pubsub.ModifyAckDeadlineRequest{
		AckDeadlineSeconds: ackDeadlineSeconds,
		AckIds:             ackIds,
	}
	return pp.subscriptionsService.ModifyAckDeadline(subscription, req).Do()
}

type (
	JobSustainerConfig struct {
		Delay    float64 `json:"delay"`
		Interval float64 `json:"interval"`
	}

	JobSubscriptionConfig struct {
		Subscription string              `json:"subscription"`
		PullInterval int                 `json:"pull_interval"`
		Sustainer    *JobSustainerConfig `json:"sustainer"`
	}

	JobSubStatus uint8

	JobSubscription struct {
		config *JobSubscriptionConfig
		puller Puller
		status JobSubStatus
		mux    sync.Mutex
	}
)

const (
	initial JobSubStatus = iota
	running
	done
	acked
)

func (s *JobSubscription) listen(ctx context.Context, f func(msg *pubsub.ReceivedMessage) error) error {
	for {
		err := s.process(ctx, f)
		if err != nil {
			return err
		}
		time.Sleep(time.Duration(s.config.PullInterval) * time.Second)
	}
	return nil
}

func (s *JobSubscription) process(ctx context.Context, f func(msg *pubsub.ReceivedMessage) error) error {
	msg, err := s.waitForMessage(ctx)
	if err != nil {
		return err
	}
	if msg == nil {
		return nil
	}

	s.status = running

	go s.sendMADPeriodically(msg.AckId)

	err = f(msg)
	s.status = done

	if err != nil {
		return err
	}

	s.mux.Lock()
	defer s.mux.Unlock()

	_, err = s.puller.Acknowledge(s.config.Subscription, msg.AckId)
	if err != nil {
		log.Fatalf("Failed to acknowledge for message: %v cause of %v\n", msg, err)
		return err
	}

	s.status = acked

	return nil
}

func (s *JobSubscription) running() bool {
	return s.status == running
}

func (s *JobSubscription) sendMADPeriodically(ackId string) error {
	for {
		nextLimit := time.Now().Add(time.Duration(s.config.Sustainer.Interval) * time.Second)
		err := s.waitAndSendMAD(nextLimit, ackId)
		if err != nil {
			return err
		}
		if !s.running() {
			return nil
		}
	}
	// return nil
}

func (s *JobSubscription) waitAndSendMAD(nextLimit time.Time, ackId string) error {
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

	_, err := s.puller.ModifyAckDeadline(s.config.Subscription, []string{ackId}, int64(s.config.Sustainer.Delay))
	if err != nil {
		log.Fatalf("Failed modifyAckDeadline %v, %v, %v cause of %v\n", s.config.Subscription, ackId, s.config.Sustainer.Delay, err)
	}
	return nil
}

func (s *JobSubscription) waitForMessage(ctx context.Context) (*pubsub.ReceivedMessage, error) {
	pullRequest := &pubsub.PullRequest{
		ReturnImmediately: false,
		MaxMessages:       1,
	}
	res, err := s.puller.Pull(s.config.Subscription, pullRequest)
	if err != nil {
		log.Printf("Failed to pull %v cause of %v\n", s.config.Subscription, err)
		return nil, err
	}
	if len(res.ReceivedMessages) == 0 {
		return nil, nil
	}
	return res.ReceivedMessages[0], nil
}
