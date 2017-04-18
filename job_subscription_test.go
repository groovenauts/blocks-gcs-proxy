package main

import (
	"testing"

	pubsub "google.golang.org/api/pubsub/v1"

	"github.com/stretchr/testify/assert"
)

type DummyPullerForJobSubscription struct {
	callCount int
	result    *pubsub.Subscription
}

func (p *DummyPullerForJobSubscription) Pull(subscription string, pullrequest *pubsub.PullRequest) (*pubsub.PullResponse, error) {
	return nil, nil
}
func (p *DummyPullerForJobSubscription) Acknowledge(subscription, ackId string) (*pubsub.Empty, error) {
	return nil, nil
}
func (p *DummyPullerForJobSubscription) ModifyAckDeadline(subscription string, ackIds []string, ackDeadlineSeconds int64) (*pubsub.Empty, error) {
	return nil, nil
}
func (p *DummyPullerForJobSubscription) Get(subscription string) (*pubsub.Subscription, error) {
	p.callCount++
	return p.result, nil
}

func TestJobConfigSetupSustainer(t *testing.T) {
	puller := &DummyPullerForJobSubscription{}

	jc := &JobConfig{
		Subscription: "projects/dummy-proj-999/subscriptions/test01-job-subscription",
		PullInterval: 10,
		Sustainer: &JobSustainerConfig{
			Delay:    600,
			Interval: 480,
		},
	}
	jc.setupSustainer(puller)
	assert.Equal(t, 0, puller.callCount)

	// Delay was 0
	puller = &DummyPullerForJobSubscription{
		result: &pubsub.Subscription{AckDeadlineSeconds: 300},
	}
	jc = &JobConfig{
		Subscription: "projects/dummy-proj-999/subscriptions/test01-job-subscription",
		PullInterval: 10,
		Sustainer: &JobSustainerConfig{
			//Delay: 600,
			Interval: 480,
		},
	}
	jc.setupSustainer(puller)
	assert.Equal(t, 1, puller.callCount)
	assert.Equal(t, float64(300), jc.Sustainer.Delay)
	assert.Equal(t, float64(480), jc.Sustainer.Interval)

	// Interval was 0
	puller = &DummyPullerForJobSubscription{
		result: &pubsub.Subscription{AckDeadlineSeconds: 300},
	}
	jc = &JobConfig{
		Subscription: "projects/dummy-proj-999/subscriptions/test01-job-subscription",
		PullInterval: 10,
		Sustainer: &JobSustainerConfig{
			Delay: 300,
			// Interval: 480,
		},
	}
	jc.setupSustainer(puller)
	assert.Equal(t, 1, puller.callCount)
	assert.Equal(t, float64(300), jc.Sustainer.Delay)
	assert.Equal(t, float64(240), jc.Sustainer.Interval)

	// No sustainer
	puller = &DummyPullerForJobSubscription{
		result: &pubsub.Subscription{AckDeadlineSeconds: 400},
	}
	jc = &JobConfig{
		Subscription: "projects/dummy-proj-999/subscriptions/test01-job-subscription",
		PullInterval: 10,
	}
	jc.setupSustainer(puller)
	assert.Equal(t, 1, puller.callCount)
	assert.Equal(t, float64(400), jc.Sustainer.Delay)
	assert.Equal(t, float64(320), jc.Sustainer.Interval)
}


type DummyPuller struct {
	responses []*pubsub.PullResponse
}

func (p *DummyPuller) Pull(subscription string, pullrequest *pubsub.PullRequest) (*pubsub.PullResponse, error) {
	if len(p.responses) > 0 {
		res := p.responses[0]
		p.responses = p.responses[1:]
		return res, nil
	} else {
		return nil, nil
	}
}
func (p *DummyPuller) Acknowledge(subscription, ackId string) (*pubsub.Empty, error) {
	return nil, nil
}
func (p *DummyPuller) ModifyAckDeadline(subscription string, ackIds []string, ackDeadlineSeconds int64) (*pubsub.Empty, error) {
	return nil, nil
}
func (p *DummyPuller) Get(subscription string) (*pubsub.Subscription, error) {
	return nil, nil
}

func TestJobSubscriptionProcess(t *testing.T) {
	puller := &DummyPuller{
		responses: []*pubsub.PullResponse{
			&pubsub.PullResponse{
				ReceivedMessages: []*pubsub.ReceivedMessage{
					&pubsub.ReceivedMessage{
						AckId: "dummy-ack-id1",
						Message: &pubsub.PubsubMessage{
							Attributes: map[string]string {
								"foo": "A",
							},
							MessageId: "dummy-msg-id1",
						},
					},
				},
			},
			&pubsub.PullResponse{
				ReceivedMessages: []*pubsub.ReceivedMessage{},
			},
		},
	}

	jc := &JobConfig{
		Subscription: "projects/dummy-proj-999/subscriptions/test01-job-subscription",
		PullInterval: 10,
		Sustainer: &JobSustainerConfig{
			Delay:    600,
			Interval: 480,
		},
	}
	jc.setupSustainer(puller)

	s := &JobSubscription{
		config: jc,
		puller: puller,
	}

	f := func(msg *JobMessage) error {
		return nil
	}
	// MessageId: dummy-msg-id1
	executed, error := s.process(f)
	assert.True(t, executed)
	assert.NoError(t, error)
	// PullResponse is returned but any ReceivedMessage doesn't exist
	executed, error = s.process(f)
	assert.False(t, executed)
	assert.NoError(t, error)
	// PullResponse isn't returned
	executed, error = s.process(f)
	assert.False(t, executed)
	assert.NoError(t, error)
}
