package gcsproxy

import (
	// "golang.org/x/net/context"

	pubsub "google.golang.org/api/pubsub/v1"
)

type (
	Publisher interface {
		Publish(topic string, publishrequest *pubsub.PublishRequest) (*pubsub.PublishResponse, error)
	}

	pubsubPublisher struct {
		topicsService *pubsub.ProjectsTopicsService
	}
)

func (pp *pubsubPublisher) Publish(topic string, publishrequest *pubsub.PublishRequest) (*pubsub.PublishResponse, error) {
	return pp.topicsService.Publish(topic, publishrequest).Do()
}

type (
	ProgressNotificationConfig struct {
		Topic string
	}

	ProgressNotification struct {
		config    *ProgressNotificationConfig
		publisher Publisher
	}
)
