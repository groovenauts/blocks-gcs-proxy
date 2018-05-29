package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
	"text/template"
)

type (
	ProcessConfig struct {
		Command  *CommandConfig              `json:"command,omitempty"`
		Job      *JobSubscriptionConfig      `json:"job,omitempty"`
		JobCheck *JobCheckConfig             `json:"job_check,omitempty"`
		Progress *ProgressNotificationConfig `json:"progress,omitempty"`
		Log      *LogConfig                  `json:"log,omitempty"`
		Download *DownloadConfig             `json:"download"`
		Upload   *UploadConfig               `json:"upload"`
	}
)

func (c *ProcessConfig) setup(args []string) error {
	setups := map[string]ConfigSetup{
		"command": func() *ConfigError {
			return c.setupCommand(args)
		},
		"job":       c.setupJob,
		"job_check": c.setupJobCheck,
		"progress":  c.setupProgress,
		"log":       c.setupLog,
		"download":  c.setupDownload,
		"upload":    c.setupUpload,
	}
	for key, setup := range setups {
		err := setup()
		if err != nil {
			err.Add(key)
			return err
		}
	}
	return nil
}

func (c *ProcessConfig) setupCommand(args []string) *ConfigError {
	if c.Command == nil {
		c.Command = &CommandConfig{}
	}
	c.Command.Template = args
	return c.Command.setup()
}

func (c *ProcessConfig) setupJob() *ConfigError {
	if c.Job == nil {
		c.Job = &JobSubscriptionConfig{}
	}
	return c.Job.setup()
}

func (c *ProcessConfig) setupJobCheck() *ConfigError {
	if c.JobCheck == nil {
		c.JobCheck = &JobCheckConfig{}
	}
	return c.JobCheck.setup()
}

func (c *ProcessConfig) setupProgress() *ConfigError {
	if c.Progress == nil {
		c.Progress = &ProgressNotificationConfig{}
	}
	return c.Progress.setup()
}

func (c *ProcessConfig) setupLog() *ConfigError {
	if c.Log == nil {
		c.Log = &LogConfig{}
	}
	return c.Log.setup()
}

func (c *ProcessConfig) setupDownload() *ConfigError {
	if c.Download == nil {
		c.Download = &DownloadConfig{}
	}
	return c.Download.setup()
}

func (c *ProcessConfig) setupUpload() *ConfigError {
	if c.Upload == nil {
		c.Upload = &UploadConfig{}
	}
	return c.Upload.setup()
}

func LoadProcessConfig(path string) (*ProcessConfig, error) {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	funcMap := template.FuncMap{"env": os.Getenv}
	t, err := template.New("config").Funcs(funcMap).Parse(string(raw))
	if err != nil {
		return nil, err
	}

	env := map[string]string{}
	for _, s := range os.Environ() {
		parts := strings.SplitN(s, "=", 2)
		env[parts[0]] = parts[1]
	}

	buf := new(bytes.Buffer)
	t.Execute(buf, env)

	var res ProcessConfig
	err = json.Unmarshal(buf.Bytes(), &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}
