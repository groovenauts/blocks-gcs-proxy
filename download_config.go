package main

type DownloadConfig struct {
	Worker            *WorkerConfig `json:"worker,omitempty"`
	AllowIrregularUrl bool          `json:"allow_irregular_url,omitempty"`
}

func (c *DownloadConfig) setup() *ConfigError {
	if c.Worker == nil {
		c.Worker = &WorkerConfig{}
	}
	return c.Worker.setup()
}
