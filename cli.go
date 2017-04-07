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

	configFlag := cli.StringFlag{
		Name:  "config, c",
		Usage: "Load configuration from `FILE`",
	}
	app.Flags = []cli.Flag{
		configFlag,
	}

	app.Commands = []cli.Command{
		{
			Name:  "check",
			Usage: "Check config file is valid",
			Action: func(c *cli.Context) error {
				LoadAndSetupProcessConfig(c)
				fmt.Println("OK")
				return nil
			},
			Flags: []cli.Flag{
				configFlag,
			},
		},
	}

	app.Action = run

	app.Run(os.Args)
}

func run(c *cli.Context) error {
	config := LoadAndSetupProcessConfig(c)

	ctx := context.Background()

	p := &Process{config: config}
	err := p.setup(ctx)
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

func LoadAndSetupProcessConfig(c *cli.Context) *ProcessConfig {
	path := configPath(c)
	config, err := LoadProcessConfig(path)
	if err != nil {
		fmt.Printf("Error to load %v cause of %v\n", path, err)
		os.Exit(1)
	}
	err = config.setup(os.Args[1:])
	if err != nil {
		fmt.Printf("Error to setup %v cause of %v\n", path, err)
		os.Exit(1)
	}
	return config
}

func configPath(c *cli.Context) string {
	r := c.String("config")
	if r == "" {
		r = "./config.json"
	}
	return r
}
