package main

type LoggingConfig struct {
	ProjectID       string						`json:"project_id"`
	LogName		      string						`json:"log_name"`
	Type            string            `json:"type"`
	Labels          map[string]string `json:"labels"`
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
