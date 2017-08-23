package main

import (
	"github.com/cenkalti/backoff"

	pubsub "google.golang.org/api/pubsub/v1"
)

type (
	Puller interface {
		Pull(subscription string, pullrequest *pubsub.PullRequest) (*pubsub.PullResponse, error)
		Acknowledge(subscription, ackId string) (*pubsub.Empty, error)
		ModifyAckDeadline(subscription string, ackIds []string, ackDeadlineSeconds int64) (*pubsub.Empty, error)
		Get(subscription string) (*pubsub.Subscription, error)
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

func (pp *pubsubPuller) Get(subscription string) (*pubsub.Subscription, error) {
	return pp.subscriptionsService.Get(subscription).Do()
}


type BackoffPuller struct {
	Impl     Puller
	Backoff  backoff.BackOff
}

func (bp *BackoffPuller) Pull(subscription string, pullrequest *pubsub.PullRequest) (res *pubsub.PullResponse, err error) {
	f := func() error{
		var e error
		res, e = bp.Impl.Pull(subscription, pullrequest)
		return e
	}
	err = backoff.Retry(f, bp.Backoff)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (bp *BackoffPuller) Acknowledge(subscription, ackId string) (empty *pubsub.Empty, err error) {
	f := func() error {
		var e error
		empty, e = bp.Impl.Acknowledge(subscription, ackId)
		return e
	}
	err = backoff.Retry(f, bp.Backoff)
	if err != nil {
		return nil, err
	}
	return empty, nil
}

func (bp *BackoffPuller) ModifyAckDeadline(subscription string, ackIds []string, ackDeadlineSeconds int64) (empty *pubsub.Empty, err error) {
	f := func() error {
		var e error
		empty, e = bp.Impl.ModifyAckDeadline(subscription, ackIds, ackDeadlineSeconds)
		return e
	}
	err = backoff.Retry(f, bp.Backoff)
	if err != nil {
		return nil, err
	}
	return empty, nil
}

func (bp *BackoffPuller) Get(subscription string) (sub *pubsub.Subscription, err error) {
	f := func() error {
		var e error
		sub, e = bp.Impl.Get(subscription)
		return e
	}
	err = backoff.Retry(f, bp.Backoff)
	if err != nil {
		return nil, err
	}
	return sub, nil
}
