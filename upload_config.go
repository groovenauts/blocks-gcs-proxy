package main

type UploadConfig struct {
	Worker           *WorkerConfig `json:"worker,omitempty"`
	ContentTypeByExt bool          `json:"content_type_by_ext,omitempty"`
}

func (c *UploadConfig) setup() *ConfigError {
	if c.Worker == nil {
		c.Worker = &WorkerConfig{}
	}
	return c.Worker.setup()
}
