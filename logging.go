package main

type LoggingConfig struct {
	ProjectID       string						`json:"project_id"`
	LogName		      string						`json:"log_name"`
	Type            string            `json:"type"`
	Labels          map[string]string `json:"labels"`
}

func (c *LoggingConfig) setup() *ConfigError {
	if c.ProjectID == "" {
		return &ConfigError{Name: "project_id", Message: "is required"}
	}
	if c.LogName == "" {
		return &ConfigError{Name: "log_name", Message: "is required"}
	}
	if c.Type == "" {
		return &ConfigError{Name: "type", Message: "is required"}
	}
	if c.Labels == nil {
		return &ConfigError{Name: "labels", Message: "are required"}
	}
	return nil
}
