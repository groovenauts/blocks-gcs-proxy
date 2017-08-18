package main

import (
	"time"

	pubsub "google.golang.org/api/pubsub/v1"

	logrus "github.com/sirupsen/logrus"
)

type JobSubscriptionConfig struct {
	Subscription string              `json:"subscription,omitempty"`
	PullInterval int                 `json:"pull_interval,omitempty"`
	Sustainer    *JobSustainerConfig `json:"sustainer,omitempty"`
}

func (c *JobSubscriptionConfig) setup() *ConfigError {
	if c.PullInterval == 0 {
		c.PullInterval = 10
	}
	return nil
}

func (c *JobSubscriptionConfig) setupSustainer(puller Puller) error {
	flds := logrus.Fields{"subscription": c.Subscription}
	if c.Sustainer != nil {
		cs := c.Sustainer
		if cs.Delay > 0 && cs.Interval > 0 {
			flds["delay"] = cs.Delay
			flds["interval"] = cs.Interval
			log.WithFields(flds).Infoln("Sustainer config OK")
			return nil
		}
	} else {
		c.Sustainer = &JobSustainerConfig{}
	}

	subscription, err := puller.Get(c.Subscription)
	if err != nil {
		flds["error"] = err
		log.WithFields(flds).Errorln("Failed to get subscription")
		return err
	}
	deadline := subscription.AckDeadlineSeconds
	flds["AckDeadline"] = deadline
	log.WithFields(flds).Infoln("AckDeadlineSeconds")

	cs := c.Sustainer
	if cs.Delay == 0 {
		cs.Delay = float64(deadline)
	}
	if cs.Interval == 0 {
		cs.Interval = float64(deadline) * 0.8
	}
	flds["delay"] = cs.Delay
	flds["interval"] = cs.Interval
	log.WithFields(flds).Infoln("Sustainer config OK")
	return nil
}

type JobSubscription struct {
	config *JobSubscriptionConfig
	puller Puller
}

func (s *JobSubscription) listen(f func(*JobMessage) error) error {
	for {
		executed, err := s.process(f)
		if err != nil {
			return err
		}
		if !executed {
			time.Sleep(time.Duration(s.config.PullInterval) * time.Second)
		}
	}
}

func (s *JobSubscription) process(f func(*JobMessage) error) (bool, error) {
	msg, err := s.waitForMessage()
	if err != nil {
		return false, err
	}
	if msg == nil {
		return false, nil
	}

	log.WithFields(logrus.Fields{"job_message_id": msg.Message.MessageId, "message": msg.Message}).Infoln("Message received")

	jobMsg := &JobMessage{
		sub:    s.config.Subscription,
		raw:    msg,
		config: s.config.Sustainer,
		puller: s.puller,
		status: running,
	}

	return true, f(jobMsg)
}

func (s *JobSubscription) waitForMessage() (*pubsub.ReceivedMessage, error) {
	pullRequest := &pubsub.PullRequest{
		ReturnImmediately: false,
		MaxMessages:       1,
	}
	res, err := s.puller.Pull(s.config.Subscription, pullRequest)
	if err != nil {
		log.WithFields(logrus.Fields{"subscription": s.config.Subscription, "error": err}).Errorln("Failed to pull")
		return nil, err
	}
	if res == nil {
		return nil, nil
	}
	if len(res.ReceivedMessages) == 0 {
		return nil, nil
	}
	return res.ReceivedMessages[0], nil
}
