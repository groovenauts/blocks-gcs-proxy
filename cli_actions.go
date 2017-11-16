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
	flag_config              = "config"
	flag_log_config          = "log-config"
	flag_workers             = "workers"
	flag_max_tries           = "max_tries"
	flag_wait                = "wait"
	flag_downloads_dir       = "downloads_dir"
	flag_uploads_dir         = "uploads_dir"
	flag_content_type_by_ext = "content_type_by_ext"
	flag_message             = "message"
	flag_workspace           = "workspace"
)

var flagAliases = map[string]string{
	// For all
	flag_config:     "c",
	flag_log_config: "l",
	// For Download and Upload
	flag_workers:   "n",
	flag_max_tries: "M",
	flag_wait:      "W",
	// For Download only
	flag_downloads_dir: "d",
	// For Upload only
	flag_uploads_dir: "d",
	// For Exec only
	flag_message:   "m",
	flag_workspace: "w",
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

func (act *CliActions) flagWorkers() cli.IntFlag {
	return cli.IntFlag{
		Name:  act.flagName(flag_workers),
		Usage: "`NUMBER` of workers",
		Value: 5,
	}
}

func (act *CliActions) flagMaxTries() cli.IntFlag {
	return cli.IntFlag{
		Name:  act.flagName(flag_max_tries),
		Usage: "`NUMBER` of max tries",
		Value: 3,
	}
}

func (act *CliActions) flagWait() cli.IntFlag {
	return cli.IntFlag{
		Name:  act.flagName(flag_wait),
		Usage: "`NUMBER` of seconds to wait",
		Value: 0,
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
		Name:   "check",
		Usage:  "Check config file is valid",
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
		Name:   "download",
		Usage:  "Download the files from GCS to downloads directory",
		Action: act.Download,
		Flags: []cli.Flag{
			act.flagConfig(),
			act.flagLogConfig(),
			cli.StringFlag{
				Name:  act.flagName(flag_downloads_dir),
				Usage: "`PATH` to the directory which has bucket_name/path/to/file",
			},
			act.flagWorkers(),
			act.flagMaxTries(),
			act.flagWait(),
		},
	}
}

func (act *CliActions) Download(c *cli.Context) error {
	config := act.LoadAndSetupProcessConfigWith(c, func(cfg *ProcessConfig) error {
		cfg.Download.Worker.Workers = c.Int(flag_workers)
		cfg.Download.Worker.MaxTries = c.Int(flag_max_tries)
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
		downloads_dir:       c.String(flag_downloads_dir),
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
		Name:   "upload",
		Usage:  "Upload the files under uploads directory",
		Action: act.Upload,
		Flags: []cli.Flag{
			act.flagConfig(),
			act.flagLogConfig(),
			cli.StringFlag{
				Name:  act.flagName(flag_uploads_dir),
				Usage: "Path to the directory which has bucket_name/path/to/file",
			},
			cli.BoolFlag{
				Name:  act.flagName(flag_content_type_by_ext),
				Usage: "Set Content-Type by file extension with /etc/mime.types, /etc/apache2/mime.types or /etc/apache/mime.types",
			},
			act.flagWorkers(),
			act.flagMaxTries(),
			act.flagWait(),
		},
	}
}

func (act *CliActions) Upload(c *cli.Context) error {
	fmt.Printf("Uploading files\n")
	config := act.LoadAndSetupProcessConfigWith(c, func(cfg *ProcessConfig) error {
		cfg.Upload.ContentTypeByExt = c.Bool(flag_content_type_by_ext)
		cfg.Upload.Worker.Workers = c.Int(flag_workers)
		cfg.Upload.Worker.MaxTries = c.Int(flag_max_tries)
		cfg.Job.Sustainer = &JobSustainerConfig{
			Disabled: true,
		}
		return nil
	})
	p := act.newProcess(config)
	job := &Job{
		config:       config.Command,
		uploads_dir:  c.String(flag_uploads_dir),
		storage:      p.storage,
		uploadConfig: config.Upload,
	}
	fmt.Printf("Uploading files under %v\n", job.uploads_dir)
	err := job.uploadFiles()

	w := c.Int(flag_wait)
	if w > 0 {
		time.Sleep(time.Duration(w) * time.Second)
	}

	return err
}

func (act *CliActions) ExecCommand() cli.Command {
	return cli.Command{
		Name:   "exec",
		Usage:  "Execute job without download nor upload",
		Action: act.Exec,
		Flags: []cli.Flag{
			act.flagConfig(),
			act.flagLogConfig(),
			cli.StringFlag{
				Name:  act.flagName(flag_message),
				Usage: "Path to the message json file which has attributes and data",
			},
			cli.StringFlag{
				Name:  act.flagName(flag_workspace),
				Usage: "Path to workspace directory which has downloads and uploads",
			},
		},
	}
}

func (act *CliActions) Exec(c *cli.Context) error {
	config := act.LoadAndSetupProcessConfig(c)

	msg_file := c.String(flag_message)
	workspace := c.String(flag_workspace)

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
	return act.LoadAndSetupProcessConfigWith(c, func(_ *ProcessConfig) error { return nil })
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
