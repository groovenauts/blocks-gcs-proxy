package main

type LoggingConfig struct {
	Type            string            `json:"type"`
	Labels          map[string]string `json:"labels"`
}
