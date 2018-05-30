package main

import (
	"fmt"
)

type JobCheckConfig struct {
	Method   string `json:"method"`
	Database string `json:"database,omitempty"`
	Bucket   string `json:"bucket,omitempty"`
}

func (c *JobCheckConfig) setup() *ConfigError {
	c.Prepare()
	return c.Validate()
}

func (c *JobCheckConfig) Prepare() {
	if c.Method == "" {
		c.Method = "none"
	}
	switch c.Method {
	case JobCheckMethodNone:
	case JobCheckMethodBuntDB:
		if c.Database == "" {
			c.Database = "blocks-gcs-proxy.db"
		}
		if c.Bucket == "" {
			c.Bucket = "jobs:"
		}
	}
}

const (
	JobCheckMethodNone   = "none"
	JobCheckMethodBuntDB = "buntdb"
)

var JobCheckMethods = []string{
	JobCheckMethodNone,
	JobCheckMethodBuntDB,
}

func (c *JobCheckConfig) Validate() *ConfigError {
	switch c.Method {
	case JobCheckMethodNone:
		return nil
	case JobCheckMethodBuntDB:
		if c.Bucket == "" {
			return &ConfigError{Name: "bucket", Message: fmt.Sprintf("bucket is required for method %q", c.Method)}
		}
		return nil
	default:
		return &ConfigError{Name: "method", Message: fmt.Sprintf("%q is invalid. It must be one of %v", c.Method, JobCheckMethods)}
	}
}

func (c *JobCheckConfig) Checker() func(string, func() error, func() error) error {
	switch c.Method {
	case JobCheckMethodNone:
		return func(job_id string, ack, f func() error) error {
			return f()
		}
	case JobCheckMethodBuntDB:
		checker := &JobCheckByBuntDB{
			File:   c.Database,
			Prefix: c.Bucket,
		}
		return checker.Check
	default:
		return func(job_id string, ack, f func() error) error {
			return &ConfigError{Name: "method", Message: fmt.Sprintf("%q is invalid. It must be one of %v", c.Method, JobCheckMethods)}
		}
	}

}
