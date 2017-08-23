package main

import (
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
