package main

import (
	"os"

	"github.com/urfave/cli"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"

	pubsub "google.golang.org/api/pubsub/v1"
	log "github.com/Sirupsen/logrus"
)

func main() {
	app := cli.NewApp()
	app.Name = "blocks-gcs-proxy"
	app.Usage = "github.com/groovenauts/blocks-gcs-proxy"
	app.Version = VERSION

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "filepath, f",
			Usage: "Messages filepath (.jsonl JSON Lines format)",
		},
		cli.StringFlag{
			Name:  "topic, t",
			Usage: "Topic to publish",
		},
		cli.IntFlag{
			Name: "number, n",
			Usage: "Number of go routine to publish",
		},
	}

	app.Action = run
	app.Run(os.Args)
}

func run(c *cli.Context) error {
	pubsubService, err := NewPubsubService()
	if err != nil {
		return err
	}

	topic := c.String("topic")
	workers := Workers{}
	for i := 0; i < c.Int("number"); i++ {
		worker := &Worker{
			service: pubsubService,
			topic: topic,
		}
		workers = append(workers, worker)
	}
	
	workers.process(c.String("filepath"))

	return nil
}

func NewPubsubService() (*pubsub.Service, error){
	ctx := context.Background()

	// https://github.com/google/google-api-go-client#application-default-credentials-example
	client, err := google.DefaultClient(ctx, pubsub.PubsubScope)

	if err != nil {
		log.Fatalln("Failed to create DefaultClient")
		return nil, err
	}

	// Creates a pubsubService
	pubsubService, err := pubsub.New(client)
	if err != nil {
		logAttrs := log.Fields{"client": client, "error": err}
		log.WithFields(logAttrs).Fatalln("Failed to create pubsub.Service")
		return nil, err
	}

	return pubsubService, nil
}
