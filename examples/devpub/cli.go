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
		cli.IntFlag{
			Name: "number, n",
			Usage: "Number of go routine to publish",
			Value: 10,
		},
		cli.StringFlag{
			Name:  "loglevel, l",
			Usage: "Log level: debug info warn error fatal panic",
		},
	}

	app.Action = run
	app.Run(os.Args)
}

func run(c *cli.Context) error {
	loglevel := c.String("loglevel")
	if loglevel != "" {
		level, err := log.ParseLevel(loglevel)
		if err != nil {
			log.SetLevel(log.DebugLevel)
			log.Warnf("Invalid log level: %v\n", loglevel)
		} else {
			log.Debugf("Log level: %v\n", loglevel)
			log.SetLevel(level)
		}
	}

	pubsubService, err := NewPubsubService()
	if err != nil {
		return err
	}

	workers := Workers{}
	for i := 0; i < c.Int("number"); i++ {
		worker := &Worker{
			service: pubsubService,
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
