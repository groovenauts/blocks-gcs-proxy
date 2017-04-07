package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

var ConfigFilePattern = regexp.MustCompile(`\.json\z`)

func TestLoadProcessConfigReal(t *testing.T) {
	oldProject := os.Getenv("GCP_PROJECT")
	defer func() {
		err := os.Setenv("GCP_PROJECT", oldProject)
		assert.NoError(t, err)
	}()
	os.Setenv("GCP_PROJECT", "dummy-proj-999")
	os.Setenv("PIPELINE", "pipeline01")
	os.Setenv("PULL_INTERVAL", "60")
	os.Setenv("SUSTAINER_DELAY", "600")
	os.Setenv("SUSTAINER_INTERVAL", "540")
	files, err := ioutil.ReadDir("./test")
	assert.NoError(t, err)
	for _, file := range files {
		if !ConfigFilePattern.MatchString(file.Name()) {
			continue
		}
		path := "./test/" + file.Name()
		config, err := LoadProcessConfig(path)
		if assert.NoError(t, err, "path: "+path) {
			assert.NotNil(t, config)
			err = config.setup([]string{"%{attrs.command}"})
			assert.NoError(t, err)
		}
	}
}

func tempEnv(t *testing.T, env map[string]string, f func()) {
	backup := map[string]string{}
	for key, _ := range env {
		backup[key] = os.Getenv(key)
	}
	cleanup := func() {
		for key, val := range backup {
			err := os.Setenv(key, val)
			if err != nil {
				panic(err)
			}
		}
	}
	defer cleanup()
	for key, val := range env {
		err := os.Setenv(key, val)
		if err != nil {
			assert.NoError(t, err)
		}
	}
	f()
}

func TestLoadProcessConfigWithEnv(t *testing.T) {
	proj := "test-gcp-proj"
	pipeline := "pipeline1"
	pull_interval := 30
	sustainer_delay := float64(300)
	sustainer_interval := float64(270)
	tempEnv(t, map[string]string{
		"GCP_PROJECT":        proj,
		"PIPELINE":           pipeline,
		"PULL_INTERVAL":      strconv.Itoa(pull_interval),
		"SUSTAINER_DELAY":    fmt.Sprintf("%v", sustainer_delay),
		"SUSTAINER_INTERVAL": fmt.Sprintf("%v", sustainer_interval),
	}, func() {

		files := []string{
			"./test/config_with_env1.json",
			"./test/config_with_env2.json",
			"test/config_with_env_and_default3.json",
		}
		for _, path := range files {
			config, err := LoadProcessConfig(path)
			if assert.NoError(t, err) {
				assert.NotNil(t, config.Job)
				assert.NotNil(t, config.Job.Sustainer)
				assert.NotNil(t, config.Progress)
				assert.Equal(t, fmt.Sprintf("projects/%v/subscriptions/%v-job-subscription", proj, pipeline), config.Job.Subscription)
				assert.Equal(t, pull_interval, config.Job.PullInterval)
				assert.Equal(t, sustainer_delay, config.Job.Sustainer.Delay)
				assert.Equal(t, sustainer_interval, config.Job.Sustainer.Interval)
				assert.Equal(t, fmt.Sprintf("projects/%v/topics/%v-progress-topic", proj, pipeline), config.Progress.Topic)
			}
		}
	})
}

func TestLoadProcessConfigWithDefaultValues(t *testing.T) {
	proj := "test-gcp-proj"
	pipeline := "pipeline01"
	tempEnv(t, map[string]string{
		"GCP_PROJECT": proj,
	}, func() {

		config, err := LoadProcessConfig("test/config_with_env_and_default3.json")
		if assert.NoError(t, err) {
			assert.NotNil(t, config.Job)
			assert.NotNil(t, config.Job.Sustainer)
			assert.NotNil(t, config.Progress)
			assert.Equal(t, fmt.Sprintf("projects/%v/subscriptions/%v-job-subscription", proj, pipeline), config.Job.Subscription)
			assert.Equal(t, 60, config.Job.PullInterval)
			assert.Equal(t, float64(600), config.Job.Sustainer.Delay)
			assert.Equal(t, float64(540), config.Job.Sustainer.Interval)
			assert.Equal(t, fmt.Sprintf("projects/%v/topics/%v-progress-topic", proj, pipeline), config.Progress.Topic)
			assert.Equal(t, int(8), config.Command.Uploaders)
		}
	})
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
			"options": commands,
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
