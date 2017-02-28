package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadProcessConfig1(t *testing.T) {
	template := []string{"./cmd1", "%{uploads_dir}", "%{download_files}"}
	job_sub := "projects/dummy-gcp-proj/subscriptions/test-job-subscription"
	job_pull_interval := 60
	job_sus_delay := float64(600)
	job_sus_interval := float64(540)
	prog_topic := "projects/dummy-gcp-proj/topics/test-progress-topic"

	d := map[string]interface{}{
		"job": map[string]interface{}{
			"template": template,
		},
		"job_subscription": map[string]interface{}{
			"subscription":  job_sub,
			"pull_interval": job_pull_interval,
			"sustainer": map[string]interface{}{
				"delay":    job_sus_delay,
				"interval": job_sus_interval,
			},
		},
		"progress_notification": map[string]interface{}{
			"topic": prog_topic,
		},
	}

	dir, err := ioutil.TempDir("", "process_config")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	fpath := filepath.Join(dir, "config.json")
	s, err := json.Marshal(d)
	assert.NoError(t, err)
	err = ioutil.WriteFile(fpath, s, 0666)
	assert.NoError(t, err)

	config, err := LoadProcessConfig(fpath)
	assert.NoError(t, err)

	if assert.NotNil(t, config.Job) {
		assert.Equal(t, template, config.Job.Template)
		assert.Nil(t, config.Job.Commands)
		assert.Equal(t, false, config.Job.Dryrun)
	}

	if assert.NotNil(t, config.JobSubscription) {
		assert.Equal(t, job_sub, config.JobSubscription.Subscription)
		assert.Equal(t, job_pull_interval, config.JobSubscription.PullInterval)
		if assert.NotNil(t, config.JobSubscription.Sustainer) {
			assert.Equal(t, job_sus_delay, config.JobSubscription.Sustainer.Delay)
			assert.Equal(t, job_sus_interval, config.JobSubscription.Sustainer.Interval)
		}
	}
}

func TestLoadProcessConfig2(t *testing.T) {
	template := []string{"%{attrs.key}"}
	commands := map[string][]string{
		"key1": []string{"./cmd1", "%{uploads_dir}", "%{download_files.foo}", "%{download_files.bar}"},
		"key2": []string{"./cmd2", "%{uploads_dir}", "%{download_files}"},
	}
	job_sub := "projects/dummy-gcp-proj/subscriptions/test-job-subscription"
	job_pull_interval := 60
	job_sus_delay := float64(600)
	job_sus_interval := float64(540)
	prog_topic := "projects/dummy-gcp-proj/topics/test-progress-topic"

	d := map[string]interface{}{
		"job": map[string]interface{}{
			"template": template,
			"commands": commands,
		},
		"job_subscription": map[string]interface{}{
			"subscription":  job_sub,
			"pull_interval": job_pull_interval,
			"sustainer": map[string]interface{}{
				"delay":    job_sus_delay,
				"interval": job_sus_interval,
			},
		},
		"progress_notification": map[string]interface{}{
			"topic": prog_topic,
		},
	}

	dir, err := ioutil.TempDir("", "process_config")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	fpath := filepath.Join(dir, "config2.json")
	s, err := json.Marshal(d)
	assert.NoError(t, err)
	err = ioutil.WriteFile(fpath, s, 0666)
	assert.NoError(t, err)

	config, err := LoadProcessConfig(fpath)
	assert.NoError(t, err)

	if assert.NotNil(t, config.Job) {
		assert.Equal(t, template, config.Job.Template)
		assert.Equal(t, commands, config.Job.Commands)
		assert.Equal(t, false, config.Job.Dryrun)
	}

	if assert.NotNil(t, config.JobSubscription) {
		assert.Equal(t, job_sub, config.JobSubscription.Subscription)
		assert.Equal(t, job_pull_interval, config.JobSubscription.PullInterval)
		if assert.NotNil(t, config.JobSubscription.Sustainer) {
			assert.Equal(t, job_sus_delay, config.JobSubscription.Sustainer.Delay)
			assert.Equal(t, job_sus_interval, config.JobSubscription.Sustainer.Interval)
		}
	}
}
