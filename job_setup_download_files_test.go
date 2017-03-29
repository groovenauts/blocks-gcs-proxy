package main

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	pubsub "google.golang.org/api/pubsub/v1"
)

func generateJSON(t *testing.T, obj interface{}) string {
	r, err := json.Marshal(obj)
	assert.NoError(t, err)
	return string(r)
}

func TestJobSetupCase1(t *testing.T) {
	workspace := "/tmp/workspace"
	downloads_dir := workspace + "/downloads"
	uploads_dir := workspace + "/uploads"
	bucket := "bucket1"
	path1 := "path/to/file1"
	url1 := "gs://" + bucket + "/" + path1
	local1 := downloads_dir + "/" + bucket + "/" + path1

	job := &Job{
		config: &CommandConfig{
			Template: []string{"cmd1", "%{download_files}", "%{uploads_dir}"},
		},
		message: &JobMessage{
			raw: &pubsub.ReceivedMessage{
				AckId: "test-ack1",
				Message: &pubsub.PubsubMessage{
					Data: "",
					Attributes: map[string]string{
						"download_files": url1,
					},
					MessageId: "test-message1",
				},
			},
		},
		workspace:     workspace,
		downloads_dir: downloads_dir,
		uploads_dir:   uploads_dir,
	}

	err := job.setupDownloadFiles()
	assert.NoError(t, err)

	assert.Equal(t, map[string]string{
		url1: local1,
	}, job.downloadFileMap)

	assert.Equal(t, url1, job.remoteDownloadFiles)
	assert.Equal(t, local1, job.localDownloadFiles)

	err = job.build()
	assert.NoError(t, err)

	assert.Equal(t, "cmd1", job.cmd.Path)
	assert.Equal(t, []string{"cmd1", local1, uploads_dir}, job.cmd.Args)
}

func TestJobSetupCase2(t *testing.T) {
	workspace := "/tmp/workspace"
	downloads_dir := workspace + "/downloads"
	uploads_dir := workspace + "/uploads"
	bucket := "bucket1"
	path1 := "path/to/file1"
	path2 := "path/to/file2"
	path3 := "path/to/file3"
	url1 := "gs://" + bucket + "/" + path1
	url2 := "gs://" + bucket + "/" + path2
	url3 := "gs://" + bucket + "/" + path3
	local1 := downloads_dir + "/" + bucket + "/" + path1
	local2 := downloads_dir + "/" + bucket + "/" + path2
	local3 := downloads_dir + "/" + bucket + "/" + path3

	job := &Job{
		config: &CommandConfig{
			Template: []string{"cmd1", "%{uploads_dir}", "%{download_files.foo}", "%{download_files.bar}"},
		},
		message: &JobMessage{
			raw: &pubsub.ReceivedMessage{
				AckId: "test-ack1",
				Message: &pubsub.PubsubMessage{
					Data: "",
					Attributes: map[string]string{
						"download_files": generateJSON(t, map[string]interface{}{
							"foo": url1,
							"bar": []string{url2, url3},
						}),
					},
					MessageId: "test-message1",
				},
			},
		},
		workspace:     workspace,
		downloads_dir: downloads_dir,
		uploads_dir:   uploads_dir,
	}

	err := job.setupDownloadFiles()
	assert.NoError(t, err)

	assert.Equal(t, map[string]string{
		url1: local1,
		url2: local2,
		url3: local3,
	}, job.downloadFileMap)

	assert.Equal(t, map[string]interface{}{
		"foo": url1,
		"bar": []interface{}{url2, url3},
	}, job.remoteDownloadFiles)
	assert.Equal(t, map[string]interface{}{
		"foo": local1,
		"bar": []interface{}{local2, local3},
	}, job.localDownloadFiles)

	err = job.build()
	assert.NoError(t, err)

	assert.Equal(t, "cmd1", job.cmd.Path)
	assert.Equal(t, []string{"cmd1", uploads_dir, local1, local2, local3}, job.cmd.Args)
}

func TestJobSetupCase3(t *testing.T) {
	workspace := "/tmp/workspace"
	downloads_dir := workspace + "/downloads"
	uploads_dir := workspace + "/uploads"
	bucket := "bucket1"
	path1 := "path/to/file1"
	path2 := "path/to/file2"
	path3 := "path/to/file3"
	url1 := "gs://" + bucket + "/" + path1
	url2 := "gs://" + bucket + "/" + path2
	url3 := "gs://" + bucket + "/" + path3
	local1 := downloads_dir + "/" + bucket + "/" + path1
	local2 := downloads_dir + "/" + bucket + "/" + path2
	local3 := downloads_dir + "/" + bucket + "/" + path3

	attrs := map[string]string{
		"download_files": generateJSON(t, []interface{}{url1, url2, url3}),
		"foo":            "ABC",
		"bar":            "DEFG",
		"baz":            "HIJKL",
	}
	job := &Job{
		config: &CommandConfig{
			Template: []string{"cmd1", "%{uploads_dir}", "%{attrs.foo}/%{attrs.bar}", "%{attrs.baz}", "%{download_files}"},
		},
		message: &JobMessage{
			raw: &pubsub.ReceivedMessage{
				AckId: "test-ack1",
				Message: &pubsub.PubsubMessage{
					Data:       "",
					Attributes: attrs,
					MessageId:  "test-message1",
				},
			},
		},
		workspace:     workspace,
		downloads_dir: downloads_dir,
		uploads_dir:   uploads_dir,
	}

	v := job.buildVariable()
	val, err := v.Dig(v.Data, "attrs", "attrs")
	assert.NoError(t, err)
	assert.Equal(t, attrs, val)

	val, err = v.Dive("attrs.foo")
	assert.NoError(t, err)
	assert.Equal(t, attrs["foo"], val)

	val, err = v.Expand("%{attrs.foo}")
	assert.NoError(t, err)
	assert.Equal(t, attrs["foo"], val)

	err = job.setupDownloadFiles()
	assert.NoError(t, err)

	assert.Equal(t, map[string]string{
		url1: local1,
		url2: local2,
		url3: local3,
	}, job.downloadFileMap)

	assert.Equal(t, []interface{}{url1, url2, url3}, job.remoteDownloadFiles)
	assert.Equal(t, []interface{}{local1, local2, local3}, job.localDownloadFiles)

	err = job.build()
	assert.NoError(t, err)

	assert.Equal(t, "cmd1", job.cmd.Path)
	assert.Equal(t, []string{"cmd1", uploads_dir, "ABC/DEFG", "HIJKL", local1, local2, local3}, job.cmd.Args)
}

func TestJobSetupCaseWithCommandOptions(t *testing.T) {
	workspace := "/tmp/workspace"
	downloads_dir := workspace + "/downloads"
	uploads_dir := workspace + "/uploads"
	bucket := "bucket1"
	path1 := "path/to/file1"
	path2 := "path/to/file2"
	path3 := "path/to/file3"
	url1 := "gs://" + bucket + "/" + path1
	url2 := "gs://" + bucket + "/" + path2
	url3 := "gs://" + bucket + "/" + path3
	local1 := downloads_dir + "/" + bucket + "/" + path1
	local2 := downloads_dir + "/" + bucket + "/" + path2
	local3 := downloads_dir + "/" + bucket + "/" + path3

	config := &CommandConfig{
		Template: []string{"%{attrs.cmd}"},
		Options: map[string][]string{
			"default": []string{"cmd1", "%{uploads_dir}", "%{attrs.foo}/%{attrs.bar}", "%{attrs.baz}", "%{download_files}"},
			"key2":    []string{"cmd2", "%{uploads_dir}", "%{download_files}"},
		},
	}

	// Use "default" option for the message without attrs.cmd
	attrs := map[string]string{
		"download_files": generateJSON(t, []interface{}{url1, url2, url3}),
		"foo":            "ABC",
		"bar":            "DEFG",
		"baz":            "HIJKL",
	}
	job := &Job{
		config: config,
		message: &JobMessage{
			raw: &pubsub.ReceivedMessage{
				AckId: "test-ack1",
				Message: &pubsub.PubsubMessage{
					Data:       "",
					Attributes: attrs,
					MessageId:  "test-message1",
				},
			},
		},
		workspace:     workspace,
		downloads_dir: downloads_dir,
		uploads_dir:   uploads_dir,
	}

	v := job.buildVariable()
	val, err := v.Dig(v.Data, "attrs", "attrs")
	assert.NoError(t, err)
	assert.Equal(t, attrs, val)

	val, err = v.Dive("attrs.foo")
	assert.NoError(t, err)
	assert.Equal(t, attrs["foo"], val)

	val, err = v.Expand("%{attrs.foo}")
	assert.NoError(t, err)
	assert.Equal(t, attrs["foo"], val)

	err = job.setupDownloadFiles()
	assert.NoError(t, err)

	assert.Equal(t, map[string]string{
		url1: local1,
		url2: local2,
		url3: local3,
	}, job.downloadFileMap)

	assert.Equal(t, []interface{}{url1, url2, url3}, job.remoteDownloadFiles)
	assert.Equal(t, []interface{}{local1, local2, local3}, job.localDownloadFiles)

	err = job.build()
	if assert.NoError(t, err) {
		assert.Equal(t, "cmd1", job.cmd.Path)
		assert.Equal(t, []string{"cmd1", uploads_dir, "ABC/DEFG", "HIJKL", local1, local2, local3}, job.cmd.Args)
	}

	// Use "key2" option for the message with attrs.cmd
	attrs = map[string]string{
		"cmd":            "key2",
		"download_files": generateJSON(t, []interface{}{url1, url2, url3}),
		"foo":            "ABC",
		"bar":            "DEFG",
		"baz":            "HIJKL",
	}
	job = &Job{
		config: config,
		message: &JobMessage{
			raw: &pubsub.ReceivedMessage{
				AckId: "test-ack1",
				Message: &pubsub.PubsubMessage{
					Data:       "",
					Attributes: attrs,
					MessageId:  "test-message1",
				},
			},
		},
		workspace:     workspace,
		downloads_dir: downloads_dir,
		uploads_dir:   uploads_dir,
	}
	err = job.setupDownloadFiles()
	assert.NoError(t, err)

	assert.Equal(t, map[string]string{
		url1: local1,
		url2: local2,
		url3: local3,
	}, job.downloadFileMap)

	assert.Equal(t, []interface{}{url1, url2, url3}, job.remoteDownloadFiles)
	assert.Equal(t, []interface{}{local1, local2, local3}, job.localDownloadFiles)

	err = job.build()
	assert.NoError(t, err)

	assert.Equal(t, "cmd2", job.cmd.Path)
	assert.Equal(t, []string{"cmd2", uploads_dir, local1, local2, local3}, job.cmd.Args)

	// Use "invalid_key" option for the message with attrs.cmd
	attrs = map[string]string{
		"cmd":            "invalid_key",
		"download_files": generateJSON(t, []interface{}{url1, url2, url3}),
		"foo":            "ABC",
		"bar":            "DEFG",
		"baz":            "HIJKL",
	}
	job = &Job{
		config: config,
		message: &JobMessage{
			raw: &pubsub.ReceivedMessage{
				AckId: "test-ack1",
				Message: &pubsub.PubsubMessage{
					Data:       "",
					Attributes: attrs,
					MessageId:  "test-message1",
				},
			},
		},
		workspace:     workspace,
		downloads_dir: downloads_dir,
		uploads_dir:   uploads_dir,
	}

	err = job.setupDownloadFiles()
	assert.NoError(t, err)

	err = job.build()
	if assert.Error(t, err) {
		assert.Regexp(t, "Invalid command options key", err.Error())
	}
}

func TestJobSetupWithUseDataAsAttributes(t *testing.T) {
	workspace := "/tmp/workspace"
	downloads_dir := workspace + "/downloads"
	uploads_dir := workspace + "/uploads"
	bucket := "bucket1"
	path1 := "path/to/file1"
	path2 := "path/to/file2"
	path3 := "path/to/file3"
	url1 := "gs://" + bucket + "/" + path1
	url2 := "gs://" + bucket + "/" + path2
	url3 := "gs://" + bucket + "/" + path3
	// local1 := downloads_dir + "/" + bucket + "/" + path1
	// local2 := downloads_dir + "/" + bucket + "/" + path2
	// local3 := downloads_dir + "/" + bucket + "/" + path3

	job := &Job{
		config: &CommandConfig{
			Template: []string{"cmd1", "%{uploads_dir}", "%{download_files.foo}", "%{download_files.bar}"},
		},
		message: &JobMessage{
			raw: &pubsub.ReceivedMessage{
				AckId: "test-ack1",
				Message: &pubsub.PubsubMessage{
					Data: base64.StdEncoding.EncodeToString([]byte(generateJSON(t, map[string]interface{}{
						"foo":            url1,
						"download_files": []string{url2, url3},
					}))),
					Attributes: map[string]string{
						"download_files": generateJSON(t, map[string]interface{}{
							"foo": url1,
						}),
						"use-data-as-attributes": "true",
					},
					MessageId: "test-message1",
				},
			},
		},
		workspace:     workspace,
		downloads_dir: downloads_dir,
		uploads_dir:   uploads_dir,
	}

	assert.Equal(t, "", job.message.raw.Message.Attributes["foo"])

	err := job.useDataAsAttributesIfPossible()
	assert.NoError(t, err)
	assert.Equal(t, url1, job.message.raw.Message.Attributes["foo"])

	err = job.setupDownloadFiles()
	assert.NoError(t, err)

	assert.Equal(t, []interface{}{url2, url3}, job.message.DownloadFiles())
}
