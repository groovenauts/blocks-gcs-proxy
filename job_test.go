package main

import (
	"testing"

	"github.com/stretchr/testify/assert"

	pubsub "google.golang.org/api/pubsub/v1"
)

const (
	workspace = "/tmp/workspace"
	downloads_dir = workspace + "/downloads"
	uploads_dir = workspace + "/uploads"
)

func NewBasicJob() *Job {
	return &Job{
		config: &CommandConfig{
			Template: []string{"./app.sh", "%{uploads_dir}", "%{download_files.0}"},
		},
		workspace: workspace,
		downloads_dir: downloads_dir,
		uploads_dir: uploads_dir,
		localDownloadFiles: []string{downloads_dir + "/bucket1/foo"},
		remoteDownloadFiles: []string{"gs://bucket1/foo"},
		message: &JobMessage{
			raw: &pubsub.ReceivedMessage{
				Message: &pubsub.PubsubMessage{
					Attributes: map[string]string{
						"array": "[100,200,300]",
						"map": `{"foo":"A"}`,
					},
				},
			},
		},
	}
}

func TestJobBuildNormal(t *testing.T) {
	job := NewBasicJob()
	err := job.build()
	assert.NoError(t, err)
}

// Invalid index for the array "download_files"
func TestJobBuildWithInvalidIndexForArray(t *testing.T) {
	job := NewBasicJob()
	job.config.Template = []string{"./app.sh", "%{uploads_dir}", "%{download_files.1}"}
	err := job.build()
	assert.IsType(t, &InvalidJobError{}, err)
}

// Key string is given for the array "download_files"
func TestJobBuildWithStringKeyForArray(t *testing.T) {
	job := NewBasicJob()
	job.config.Template = []string{"./app.sh", "%{uploads_dir}", "%{download_files.foo}"}
	err := job.build()
	assert.IsType(t, &InvalidJobError{}, err)
}

// Invalid key given for the map "download_files"
func TestJobBuildWithInvalidKeyForMap(t *testing.T) {
	job := NewBasicJob()
	job.config.Template = []string{"./app.sh", "%{uploads_dir}", "%{download_files.baz}"}
	job.localDownloadFiles = map[string]interface{}{
		"foo": downloads_dir + "/bucket1/foo",
	}
	err := job.build()
	assert.IsType(t, &InvalidJobError{}, err)
}
