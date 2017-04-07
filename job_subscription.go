package main

import (
	"time"

	pubsub "google.golang.org/api/pubsub/v1"

	log "github.com/Sirupsen/logrus"
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

type JobConfig struct {
	Subscription string              `json:"subscription,omitempty"`
	PullInterval int                 `json:"pull_interval,omitempty"`
	Sustainer    *JobSustainerConfig `json:"sustainer,omitempty"`
}

type JobSubscription struct {
	config *JobConfig
	puller Puller
}

func (s *JobSubscription) listen(f func(*JobMessage) error) error {
	for {
		err := s.process(f)
		if err != nil {
			return err
		}
		time.Sleep(time.Duration(s.config.PullInterval) * time.Second)
	}
}

func (s *JobSubscription) process(f func(*JobMessage) error) error {
	msg, err := s.waitForMessage()
	if err != nil {
		return err
	}
	if msg == nil {
		return nil
	}

	log.WithFields(log.Fields{"job_message_id": msg.Message.MessageId, "message": msg.Message}).Infoln("Message received")

	jobMsg := &JobMessage{
		sub:    s.config.Subscription,
		raw:    msg,
		config: s.config.Sustainer,
		puller: s.puller,
		status: running,
	}

	return f(jobMsg)
}

func (s *JobSubscription) waitForMessage() (*pubsub.ReceivedMessage, error) {
	pullRequest := &pubsub.PullRequest{
		ReturnImmediately: false,
		MaxMessages:       1,
	}
	res, err := s.puller.Pull(s.config.Subscription, pullRequest)
	if err != nil {
		log.WithFields(log.Fields{"subscription": s.config.Subscription, "error": err}).Errorln("Failed to pull")
		return nil, err
	}
	if len(res.ReceivedMessages) == 0 {
		return nil, nil
	}
	return res.ReceivedMessages[0], nil
}
