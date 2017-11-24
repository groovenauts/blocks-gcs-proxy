package main

import (
	"os"
)

func FindFromEnv(keys []string) string {
	for _, key := range keys {
		v := os.Getenv(key)
		if v != "" {
			return v
		}
	}
	return ""
}

var (
	GcpProjectId = FindFromEnv([]string{"GCP_PROJECT_ID", "GCP_PROJECT", "PROJECT_ID", "PROJECT"})
	Pipeline     = FindFromEnv([]string{"PIPELINE"})
)
