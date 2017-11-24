package main

type CommandConfig struct {
	Template []string            `json:"-"`
	Options  map[string][]string `json:"options,omitempty"`
	Dryrun   bool                `json:"dryrun,omitempty"`
}

func (c *CommandConfig) setup() *ConfigError {
	return nil
}
