package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli"
	"golang.org/x/net/context"
)

func main() {
	app := cli.NewApp()
	app.Name = "blocks-gcs-proxy"
	app.Usage = "github.com/groovenauts/blocks-gcs-proxy"
	app.Version = VERSION

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config, c",
			Usage: "Load configuration from `FILE`",
		},
	}

	app.Action = run

	app.Run(os.Args)
}

func run(c *cli.Context) error {
	config, err := LoadProcessConfig(configPath(c))
	if err != nil {
		fmt.Printf("Error to load %v cause of %v\n", configPath, err)
		os.Exit(1)
	}
	config.setup(os.Args[1:])

	ctx := context.Background()

	p := &Process{config: config}
	err = p.setup(ctx)
	if err != nil {
		fmt.Printf("Error to setup Process cause of %v\n", err)
		os.Exit(1)
	}

	err = p.run()
	if err != nil {
		fmt.Printf("Error to run cause of %v\n", err)
		os.Exit(1)
	}
	return nil
}

func configPath(c *cli.Context) string {
	r := c.String("config")
	if r == "" {
		r = "./config.json"
	}
	return r
}
