package main

type JobSustainerConfig struct {
	Disabled bool    `json:"disabled,omitempty"`
	Delay    float64 `json:"delay,omitempty"`
	Interval float64 `json:"interval,omitempty"`
}
