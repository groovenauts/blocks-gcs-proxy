package main

import (
	"bytes"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"

	pubsub "google.golang.org/api/pubsub/v1"
)

const (
	workspace     = "/tmp/workspace"
	downloads_dir = workspace + "/downloads"
	uploads_dir   = workspace + "/uploads"
)

func NewBasicJob() *Job {
	return &Job{
		config: &CommandConfig{
			Template: []string{"./app.sh", "%{uploads_dir}", "%{download_files.0}"},
		},
		workspace:           workspace,
		downloads_dir:       downloads_dir,
		uploads_dir:         uploads_dir,
		localDownloadFiles:  []string{downloads_dir + "/bucket1/foo"},
		remoteDownloadFiles: []string{"gs://bucket1/foo"},
		message: &JobMessage{
			raw: &pubsub.ReceivedMessage{
				Message: &pubsub.PubsubMessage{
					Attributes: map[string]string{
						"array": "[100,200,300]",
						"map":   `{"foo":"A"}`,
					},
				},
			},
		},
	}
}

func TestJobSetupExecUUIDNormal(t *testing.T) {
	job := NewBasicJob()
	assert.Empty(t, job.execUUID)
	assert.Empty(t, job.message.raw.Message.Attributes[ExecUUIDKey])
	job.setupExecUUID()
	assert.NotEmpty(t, job.execUUID)
	assert.Equal(t, job.execUUID, job.message.raw.Message.Attributes[ExecUUIDKey])
}

func TestJobBuildNormal(t *testing.T) {
	job := NewBasicJob()
	err := job.build()
	assert.NoError(t, err)
}

func AssertCompositeErrorWithInvalidJobError(t *testing.T, err error) bool {
	if assert.IsType(t, (*CompositeError)(nil), err) {
		c := err.(*CompositeError)
		return assert.True(t, c.Any(func(e error) bool {
			switch e.(type) {
			case *InvalidJobError:
				return true
			default:
				return false
			}
		}))
	}
	return false
}

// Invalid index for the array "download_files"
func TestJobBuildWithInvalidIndexForArray(t *testing.T) {
	job := NewBasicJob()
	job.config.Template = []string{"./app.sh", "%{uploads_dir}", "%{download_files.1}"}
	err := job.build()
	AssertCompositeErrorWithInvalidJobError(t, err)
}

// Key string is given for the array "download_files"
func TestJobBuildWithStringKeyForArray(t *testing.T) {
	job := NewBasicJob()
	job.config.Template = []string{"./app.sh", "%{uploads_dir}", "%{download_files.foo}"}
	err := job.build()
	AssertCompositeErrorWithInvalidJobError(t, err)
}

// Invalid key given for the map "download_files"
func TestJobBuildWithInvalidKeyForMap(t *testing.T) {
	job := NewBasicJob()
	job.config.Template = []string{"./app.sh", "%{uploads_dir}", "%{download_files.baz}"}
	job.localDownloadFiles = map[string]interface{}{
		"foo": downloads_dir + "/bucket1/foo",
	}
	err := job.build()
	AssertCompositeErrorWithInvalidJobError(t, err)
}

// Invalid index and invalid key for the array and map in attrs
func TestJobBuildWithInvalidIndexAndKeyInAttrs(t *testing.T) {
	job := NewBasicJob()
	job.config.Template = []string{"echo", "%{attrs.array.3}", "%{attrs.map.bar}"}
	err := job.build()
	AssertCompositeErrorWithInvalidJobError(t, err)
	assert.Regexp(t, "Invalid index 3", err.Error())
	assert.Regexp(t, "Invalid key bar", err.Error())
}

// Invalid reference download_files in spite of no download_files given
func TestJobBuildWithInvalidDownloadFilesReference(t *testing.T) {
	job := NewBasicJob()
	job.config.Template = []string{"./app.sh", "%{uploads_dir}", "%{download_files}"}
	job.localDownloadFiles = nil
	err := job.build()
	if assert.Error(t, err) {
		AssertCompositeErrorWithInvalidJobError(t, err)
		assert.Regexp(t, "No value found", err.Error())
		assert.Regexp(t, "download_files", err.Error())
	}
}

func TestJobExecuteWithDryrun(t *testing.T) {
	patterns := []struct {
		dryrun   bool
		expected string
	}{
		{dryrun: false, expected: "foo\n"},
		{dryrun: true, expected: ""},
	}
	for _, ptn := range patterns {
		b := new(bytes.Buffer)
		cmd := exec.Command("echo", "foo")
		cmd.Stdout = b
		cmd.Stderr = b
		job := &Job{
			cmd: cmd,
			config: &CommandConfig{
				Dryrun: ptn.dryrun,
			},
		}
		err := job.execute()
		assert.NoError(t, err)
		assert.Equal(t, ptn.expected, b.String())
	}
}
