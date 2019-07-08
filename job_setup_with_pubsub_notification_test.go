package main

import (
	// "encoding/base64"
	// "encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	pubsub "google.golang.org/api/pubsub/v1"
)

const (
	bucket = "bucket1"
	path1  = "path/to/file1"
	url1   = "gs://" + bucket + "/" + path1
	local1 = workspace + "/" + bucket + "/" + path1
)

var (
	BaseNotificationAttrs = map[string]string{
		"objectId":           "path/to/file1",
		"payloadFormat":      "JSON_API_V1",
		"resource":           "projects/_/buckets/bucket1/objects/path/to/file1#1495443037537696",
		"bucketId":           "bucket1",
		"eventType":          "OBJECT_FINALIZE",
		"notificationConfig": "projects/_/buckets/bucket1/notificationConfigs/3",
		"objectGeneration":   "1495443037537696",
	}

	BaseNotificationData = `{
  "kind": "storage#object",
  "id": "bucket1/path/to/file1/1495443037537696",
  "selfLink": "https://www.googleapis.com/storage/v1/b/bucket1/o/files%2Fnode-v4.5.0-linux-x64.tar.xz",
  "name": "path/to/file1",
  "bucket": "bucket1",
  "generation": "1495443037537696",
  "metageneration": "1",
  "contentType": "binary/octet-stream",
  "timeCreated": "2017-05-22T08:50:37.518Z",
  "updated": "2017-05-22T08:50:37.518Z",
  "storageClass": "REGIONAL",
  "timeStorageClassUpdated": "2017-05-22T08:50:37.518Z",
  "size": "8320540",
  "md5Hash": "bXiRq/+p9S1mnM6EcdDGKQ==",
  "mediaLink": "https://www.googleapis.com/download/storage/v1/b/bucket1/o/files%2Fnode-v4.5.0-linux-x64.tar.xz?generation=1495443037537696&alt=media",
  "crc32c": "J8Knpg==",
  "etag": "CKCLo7iPg9QCEAE="
}`
)

func TestJobSetupWithPubSubNotification1(t *testing.T) {
	job := &Job{
		config: &CommandConfig{
			Template: []string{"cmd1", "%{download_files}", "%{workspace}"},
		},
		message: &JobMessage{
			raw: &pubsub.ReceivedMessage{
				AckId: "test-ack1",
				Message: &pubsub.PubsubMessage{
					Data:       BaseNotificationData,
					Attributes: BaseNotificationAttrs,
					MessageId:  "test-message1",
				},
			},
		},
		workspace: workspace,
	}

	job.remoteDownloadFiles = job.message.DownloadFiles()
	err := job.setupDownloadFiles()
	assert.NoError(t, err)

	assert.Equal(t, map[string]string{
		url1: local1,
	}, job.downloadFileMap)

	assert.Equal(t, []interface{}{url1}, job.remoteDownloadFiles)
	assert.Equal(t, []interface{}{local1}, job.localDownloadFiles)

	err = job.build()
	assert.NoError(t, err)

	assert.Equal(t, "cmd1", job.cmd.Path)
	assert.Equal(t, []string{"cmd1", local1, workspace}, job.cmd.Args)
}
