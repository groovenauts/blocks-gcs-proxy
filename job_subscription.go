package main

import (
	"log"
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
	JobConfig struct {
		Subscription string              `json:"subscription,omitempty"`
		PullInterval int                 `json:"pull_interval,omitempty"`
		Sustainer    *JobSustainerConfig `json:"sustainer,omitempty"`
	}

	JobSubscription struct {
		config *JobConfig
		puller Puller
	}
)

func (s *JobSubscription) listen(ctx context.Context, f func(*JobMessage) error) error {
	for {
		err := s.process(ctx, f)
		if err != nil {
			return err
		}
		time.Sleep(time.Duration(s.config.PullInterval) * time.Second)
	}
	return nil
}

func (s *JobSubscription) process(ctx context.Context, f func(*JobMessage) error) error {
	msg, err := s.waitForMessage(ctx)
	if err != nil {
		return err
	}
	if msg == nil {
		return nil
	}

	log.Printf("Received: AckId: %v, Message: %v\n", msg.AckId, msg.Message)

	jobMsg := &JobMessage{
		sub:    s.config.Subscription,
		raw:    msg,
		config: s.config.Sustainer,
		puller: s.puller,
		status: running,
	}

	return f(jobMsg)
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
