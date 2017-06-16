package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	pubsub "google.golang.org/api/pubsub/v1"

	"github.com/urfave/cli"
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
		{
			Name:  "download",
			Usage: "Download the files from GCS to downloads directory",
			Action: func(c *cli.Context) error {
				config := &ProcessConfig{}
				config.Log = &LogConfig{Level: "debug"}
				config.setup([]string{})
				config.Command.Downloaders = c.Int("downloaders")
				config.Job.Sustainer = &JobSustainerConfig{
					Disabled: true,
				}
				p := setupProcess(config)
				p.setup()
				files := []interface{}{}
				for _, arg := range c.Args() {
					files = append(files, arg)
				}
				job := &Job{
					config:              config.Command,
					downloads_dir:       c.String("downloads_dir"),
					remoteDownloadFiles: files,
					storage:             p.storage,
				}
				err := job.setupDownloadFiles()
				if err != nil {
					return err
				}
				err = job.downloadFiles()
				return err
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "downloads_dir, d",
					Usage: "Path to the directory which has bucket_name/path/to/file",
				},
				cli.IntFlag{
					Name:  "downloaders, n",
					Usage: "Number of downloaders",
					Value: 6,
				},
			},
		},
		{
			Name:  "upload",
			Usage: "Upload the files under uploads directory",
			Action: func(c *cli.Context) error {
				fmt.Printf("Uploading files\n")
				config := &ProcessConfig{}
				config.Log = &LogConfig{Level: "debug"}
				config.setup([]string{})
				config.Command.Uploaders = c.Int("uploaders")
				config.Job.Sustainer = &JobSustainerConfig{
					Disabled: true,
				}
				p := setupProcess(config)
				p.setup()
				job := &Job{
					config:      config.Command,
					uploads_dir: c.String("uploads_dir"),
					storage:     p.storage,
				}
				fmt.Printf("Uploading files under %v\n", job.uploads_dir)
				err := job.uploadFiles()
				return err
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "uploads_dir, d",
					Usage: "Path to the directory which has bucket_name/path/to/file",
				},
				cli.IntFlag{
					Name:  "uploaders, n",
					Usage: "Number of uploaders",
					Value: 6,
				},
			},
		},

		{
			Name:  "exec",
			Usage: "Execute job without download nor upload",
			Action: func(c *cli.Context) error {
				config := &ProcessConfig{}
				config.Log = &LogConfig{Level: "debug"}
				config.setup([]string{})
				msg_file := c.String("message")
				workspace := c.String("workspace")

				type Msg struct {
					Attributes  map[string]string `json:"attributes"`
					Data        string            `json:"data"`
					MessageId   string            `json:"messageId"`
					PublishTime string            `json:"publishTime"`
					AckId       string            `json:"ackId"`
				}
				var msg Msg

				data, err := ioutil.ReadFile(msg_file)
				if err != nil {
					fmt.Printf("Error to read file %v because of %v\n", msg_file, err)
					os.Exit(1)
				}

				err = json.Unmarshal(data, &msg)
				if err != nil {
					fmt.Printf("Error to parse json file %v because of %v\n", msg_file, err)
					os.Exit(1)
				}

				job := &Job{
					workspace: workspace,
					config:    config.Command,
					message: &JobMessage{
						raw: &pubsub.ReceivedMessage{
							AckId: msg.AckId,
							Message: &pubsub.PubsubMessage{
								Attributes: msg.Attributes,
								Data:       msg.Data,
								MessageId:  msg.MessageId,
								// PublishTime: time.Now().Format(time.RFC3339),
								PublishTime: msg.PublishTime,
							},
						},
					},
				}
				fmt.Printf("Executing job %v\n", job.workspace)
				err = job.uploadFiles()
				return err
			},
			Flags: []cli.Flag{
				configFlag,
				cli.StringFlag{
					Name:  "message, m",
					Usage: "Path to the message json file which has attributes and data",
				},
				cli.StringFlag{
					Name:  "workspace, w",
					Usage: "Path to workspace directory which has downloads and uploads",
				},
			},
		},
	}

	app.Action = run

	app.Run(os.Args)
}

func run(c *cli.Context) error {
	config := LoadAndSetupProcessConfig(c)
	p := setupProcess(config)

	err := p.run()
	if err != nil {
		fmt.Printf("Error to run cause of %v\n", err)
		os.Exit(1)
	}
	return nil
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
