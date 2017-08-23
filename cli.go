package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli"
)

func newApp() *cli.App {
	app := cli.NewApp()
	app.Name = "blocks-gcs-proxy"
	app.Usage = "github.com/groovenauts/blocks-gcs-proxy"
	app.Version = VERSION

	act := &CliActions{}
	app.Flags = act.MainFlags()
	app.Action = act.Main

	app.Commands = []cli.Command{
		act.CheckCommand(),
		act.DownloadCommand(),
		act.UploadCommand(),
		act.ExecCommand(),
	}

	return app
}

func main() {
	app := newApp()
	app.Run(os.Args)
}

func setupProcess(config *ProcessConfig) *Process {
	p := &Process{config: config}
	err := p.setup()
	if err != nil {
		fmt.Printf("Error to setup Process cause of %v\n", err)
		os.Exit(1)
	}
	return p
}

func LoadAndSetupProcessConfig(c *cli.Context) *ProcessConfig {
	path := configPath(c)
	config, err := LoadProcessConfig(path)
	if err != nil {
		fmt.Printf("Error to load %v cause of %v\n", path, err)
		os.Exit(1)
	}
	err = config.setup(c.Args())
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
