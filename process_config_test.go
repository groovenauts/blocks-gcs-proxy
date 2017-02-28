package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadProcessConfigReal(t *testing.T) {
	_, err := LoadProcessConfig("./test/config1.json")
	assert.NoError(t, err)
}

func TestLoadProcessConfig1(t *testing.T) {
	// template := []string{"./cmd1", "%{uploads_dir}", "%{download_files}"}
	job_sub := "projects/dummy-gcp-proj/subscriptions/test-job-subscription"
	job_pull_interval := 60
	job_sus_delay := float64(600)
	job_sus_interval := float64(540)
	prog_topic := "projects/dummy-gcp-proj/topics/test-progress-topic"

	d := map[string]interface{}{
		// "command": map[string]interface{}{
		// 	"template": template,
		// },
		"job": map[string]interface{}{
			"subscription":  job_sub,
			"pull_interval": job_pull_interval,
			"sustainer": map[string]interface{}{
				"delay":    job_sus_delay,
				"interval": job_sus_interval,
			},
		},
		"progress": map[string]interface{}{
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

	// config.Command is nil
	assert.Nil(t, config.Command)

	if assert.NotNil(t, config.Job) {
		assert.Equal(t, job_sub, config.Job.Subscription)
		assert.Equal(t, job_pull_interval, config.Job.PullInterval)
		if assert.NotNil(t, config.Job.Sustainer) {
			assert.Equal(t, job_sus_delay, config.Job.Sustainer.Delay)
			assert.Equal(t, job_sus_interval, config.Job.Sustainer.Interval)
		}
	}

	if assert.NotNil(t, config.Progress) {
		assert.Equal(t, prog_topic, config.Progress.Topic)
	}
}

func TestLoadProcessConfig2(t *testing.T) {
	// template := []string{"%{attrs.key}"}
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
		"command": map[string]interface{}{
			// "template": template,
			"options":  commands,
		},
		"job": map[string]interface{}{
			"subscription":  job_sub,
			"pull_interval": job_pull_interval,
			"sustainer": map[string]interface{}{
				"delay":    job_sus_delay,
				"interval": job_sus_interval,
			},
		},
		"progress": map[string]interface{}{
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
		// assert.Equal(t, template, config.Command.Template)
		assert.Equal(t, commands, config.Command.Options)
		assert.Equal(t, false, config.Command.Dryrun)
	}

	if assert.NotNil(t, config.Job) {
		assert.Equal(t, job_sub, config.Job.Subscription)
		assert.Equal(t, job_pull_interval, config.Job.PullInterval)
		if assert.NotNil(t, config.Job.Sustainer) {
			assert.Equal(t, job_sus_delay, config.Job.Sustainer.Delay)
			assert.Equal(t, job_sus_interval, config.Job.Sustainer.Interval)
		}
	}

	if assert.NotNil(t, config.Progress) {
		assert.Equal(t, prog_topic, config.Progress.Topic)
	}
}
