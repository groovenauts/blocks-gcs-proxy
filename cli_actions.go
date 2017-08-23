package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	pubsub "google.golang.org/api/pubsub/v1"

	"github.com/urfave/cli"
)

const (
	flag_config = "config"
	flag_log_config = "log-config"
	flag_workers = "workers"
	flag_max_tries = "max_tries"
	flag_wait = "wait"
)

var flagAliases = map[string]string{
	flag_config: "c",
	flag_log_config: "l",
	flag_workers: "n",
	flag_max_tries: "m",
	flag_wait: "w",
}

type CliActions struct {
}

func (act *CliActions) flagName(flag string) string {
	return fmt.Sprintf("%s, %s", flag, flagAliases[flag])
}

func (act *CliActions) flagConfig() cli.StringFlag {
	return cli.StringFlag{
		Name:  act.flagName(flag_config),
		Usage: "`FILE` to load configuration",
		Value: "./config.json",
	}
}

func (act *CliActions) flagLogConfig() cli.BoolFlag {
	return cli.BoolFlag{
		Name:  act.flagName(flag_log_config),
		Usage: "Set to log your configuration loaded",
	}
}

func (act *CliActions) MainFlags() []cli.Flag {
	return []cli.Flag{
		act.flagConfig(),
		act.flagLogConfig(),
	}
}

func (act *CliActions) Main(c *cli.Context) error {
	config := act.LoadAndSetupProcessConfig(c)
	p := act.newProcess(config)

	err := p.run()
	if err != nil {
		fmt.Printf("Error to run cause of %v\n", err)
		os.Exit(1)
	}
	return nil
}


func (act *CliActions) CheckCommand() cli.Command {
	return cli.Command{
		Name:  "check",
		Usage: "Check config file is valid",
		Action: act.Check,
		Flags: []cli.Flag{
			act.flagConfig(),
			act.flagLogConfig(),
		},
	}
}

func (act *CliActions) Check(c *cli.Context) error {
	act.LoadAndSetupProcessConfig(c)
	fmt.Println("OK")
	return nil
}

func (act *CliActions) DownloadCommand() cli.Command {
	return cli.Command{
		Name:  "download",
		Usage: "Download the files from GCS to downloads directory",
		Action: act.Download,
		Flags: []cli.Flag{
			act.flagConfig(),
			act.flagLogConfig(),
			cli.StringFlag{
				Name:  "downloads_dir, d",
				Usage: "`PATH` to the directory which has bucket_name/path/to/file",
			},
			cli.IntFlag{
				Name:  act.flagName(flag_workers),
				Usage: "`NUMBER` of workers",
				Value: 5,
			},
			cli.IntFlag{
				Name:  act.flagName(flag_max_tries),
				Usage: "`NUMBER` of max tries",
				Value: 3,
			},
			cli.IntFlag{
				Name:  act.flagName(flag_wait),
				Usage: "`NUMBER` of seconds to wait",
				Value: 0,
			},
		},
	}
}

func (act *CliActions) Download(c *cli.Context) error {
	config := act.LoadAndSetupProcessConfigWith(c, func(cfg *ProcessConfig) error {
		cfg.Download.Workers = c.Int(flag_workers)
		cfg.Download.MaxTries = c.Int(flag_max_tries)
		cfg.Job.Sustainer = &JobSustainerConfig{
			Disabled: true,
		}
		return nil
	})
	p := act.newProcess(config)
	files := []interface{}{}
	for _, arg := range c.Args() {
		files = append(files, arg)
	}
	job := &Job{
		config:              config.Command,
		downloads_dir:       c.String("downloads_dir"),
		remoteDownloadFiles: files,
		storage:             p.storage,
		downloadConfig:      config.Download,
	}
	err := job.setupDownloadFiles()
	if err != nil {
		return err
	}
	err = job.downloadFiles()

	w := c.Int(flag_wait)
	if w > 0 {
		time.Sleep(time.Duration(w) * time.Second)
	}

	return nil
}


func (act *CliActions) UploadCommand() cli.Command {
	return cli.Command{
		Name:  "upload",
		Usage: "Upload the files under uploads directory",
		Action: act.Upload,
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
	}
}

func (act *CliActions) Upload(c *cli.Context) error {
	fmt.Printf("Uploading files\n")
	config := &ProcessConfig{}
	config.Log = &LogConfig{Level: "debug"}
	config.setup([]string{})
	config.Upload.Workers = c.Int("uploaders")
	config.Job.Sustainer = &JobSustainerConfig{
		Disabled: true,
	}
	p := act.newProcess(config)
	p.setup()
	job := &Job{
		config:      config.Command,
		uploads_dir: c.String("uploads_dir"),
		storage:     p.storage,
	}
	fmt.Printf("Uploading files under %v\n", job.uploads_dir)
	err := job.uploadFiles()
	return err
}


func (act *CliActions) ExecCommand() cli.Command {
	return cli.Command{
		Name:  "exec",
		Usage: "Execute job without download nor upload",
		Action: act.Exec,
		Flags: []cli.Flag{
			act.flagConfig(),
			act.flagLogConfig(),
			cli.StringFlag{
				Name:  "message, m",
				Usage: "Path to the message json file which has attributes and data",
			},
			cli.StringFlag{
				Name:  "workspace, w",
				Usage: "Path to workspace directory which has downloads and uploads",
			},
		},
	}
}

func (act *CliActions) Exec(c *cli.Context) error {
	config := act.LoadAndSetupProcessConfig(c)

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
	fmt.Printf("Preparing job\n")
	err = job.prepare()
	if err != nil {
		return err
	}
	fmt.Printf("Executing job\n")
	err = job.execute()
	return err
}


func (act *CliActions) LoadAndSetupProcessConfig(c *cli.Context) *ProcessConfig {
	return act.LoadAndSetupProcessConfigWith(c, func(_ *ProcessConfig) error{ return nil})
}

func (act *CliActions) LoadAndSetupProcessConfigWith(c *cli.Context, callback func(*ProcessConfig) error) *ProcessConfig {
	path := c.String(flag_config)
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
	err = callback(config)
	if err != nil {
		fmt.Printf("Error to callback on setup %v cause of %v\n", path, err)
		os.Exit(1)
	}

	if c.Bool(flag_log_config) {
		err = act.LogConfig(config)
		if err != nil {
			fmt.Printf("Error to log config %v cause of %v\n", path, err)
			os.Exit(1)
		}
	}

	return config
}

func (act *CliActions) newProcess(config *ProcessConfig) *Process {
	p := &Process{config: config}
	err := p.setup()
	if err != nil {
		fmt.Printf("Error to setup Process cause of %v\n", err)
		os.Exit(1)
	}
	return p
}

func (act *CliActions) LogConfig(config *ProcessConfig) error {
	text, err := json.MarshalIndent(config, "", "  ")
  if err != nil {
		log.Errorf("Failed to json.MarshalIndent because of %v\n", err)
		return err
	}
	log.Infof("config:\n%v\n", string(text))
	return nil
}
