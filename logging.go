package main

type LoggingConfig struct {
	Type            string            `json:"type"`
	Labels          map[string]string `json:"labels"`
}

func (c *LoggingConfig) setup() *ConfigError {
	if c.Type == "" {
		return &ConfigError{Name: "type", Message: "is required"}
	}
	if c.Labels == nil {
		return &ConfigError{Name: "labels", Message: "are required"}
	}
	return nil
}
