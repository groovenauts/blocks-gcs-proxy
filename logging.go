package main

import (
	"net/http"

	"github.com/knq/sdhook"
	log "github.com/sirupsen/logrus"
)

type LoggingConfig struct {
	ProjectID       string						`json:"project_id"`
	LogName		      string						`json:"log_name"`
	Type            string            `json:"type"`
	Labels          map[string]string `json:"labels"`
	ErrorReportingService string      `json:"error_reporting_service"`
}

func (c *LoggingConfig) setup() *ConfigError {
	for name, blank := range map[string]bool {
		"project_id": c.ProjectID == "",
		"log_name": c.LogName == "",
		"type": c.Type == "",
		"labels": c.Labels == nil,
	}{
		if blank {
			return &ConfigError{Name: name, Message: "is required"}
		}
	}
	return nil
}

func (c *LoggingConfig) setupSdHook(client *http.Client) error {
	options := []sdhook.Option{
		sdhook.HTTPClient(client),
		sdhook.ProjectID(c.ProjectID),
		sdhook.LogName(c.LogName),
		sdhook.Resource(sdhook.ResType(c.Type), c.Labels),
	}
	if c.ErrorReportingService != "" {
		options = append(options, sdhook.ErrorReportingService(c.ErrorReportingService))
	}
	hook, err := sdhook.New(options...)
	if err != nil {
		return err
	}
	log.AddHook(hook)
	return nil
}
