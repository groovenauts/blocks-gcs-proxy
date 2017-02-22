package gcsproxy

import (
	"log"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"

	pubsub "google.golang.org/api/pubsub/v1"
)

type (
	ProcessConfig struct {
		Job *JobConfig
		JobSubscription *JobSubscriptionConfig
		ProgressNotification *ProgressNotificationConfig
	}
)

func (c *ProcessConfig) setup(ctx context.Context, args []string) error {
	if c.Job == nil {
		c.Job = &JobConfig{}
	}
	c.Job.Template = args
	return nil
}

type (
	Process struct {
		config *ProcessConfig
		subscription *JobSubscription
		notification *ProgressNotification
	}
)

func (p *Process) setup(ctx context.Context) error {
	// https://github.com/google/google-api-go-client#application-default-credentials-example
	client, err := google.DefaultClient(ctx, pubsub.PubsubScope)
	if err != nil {
		log.Printf("Failed to create DefaultClient\n")
		return err
	}

	// Creates a pubsubClient
	service, err := pubsub.New(client)
	if err != nil {
		log.Printf("Failed to create pubsub.Service with %v: %v\n", client, err)
		return err
	}

	
	p.subscription = &JobSubscription{
		config: p.config.JobSubscription,
		puller: &pubsubPuller{service.Projects.Subscriptions},
	}
	p.notification = &ProgressNotification{
		config: p.config.ProgressNotification,
		publisher: &pubsubPublisher{service.Projects.Topics},
	}
	return nil
}

func (p *Process) run(ctx context.Context) error{
	err := p.subscription.listen(ctx, func(msg *pubsub.ReceivedMessage) error{
		job := &Job{
			config: p.config.Job,
			message: msg,
			notification: p.notification,
		}
		err := job.execute(ctx)
		if err != nil {
			log.Printf("Job Error %v cause of %v\n", msg, err)
			return err
		}
		return nil
	})
	return err
}
