# Use blocks-gcs-proxy with Cloud Pub/Sub Notifications

## Overview

If you run your containers after a file uploaded to the specified bucket,
You can use blocks-gcs-proxy with [Cloud Pub/Sub Notifications for Google Cloud Storage](https://cloud.google.com/storage/docs/pubsub-notifications).

[Cloud Pub/Sub Notifications for Google Cloud Storage](https://cloud.google.com/storage/docs/pubsub-notifications) publishes
a message to the specified topic after file is uploaded to the specified bucket. blocks-gcs-proxy accepts the messages and
run your command.

## How to setup

1. Create notification config to your bucket
2. Deploy/Run your container

### Create notification config to your bucket

```
gsutil notification create -t projects/[Your-Project]/topics/[Your-Pipeline-Job-Topic] -f json gs://[Your-Bucket] -e OBJECT_FINALIZE
```

`blocks-gcs-proxy` doesn't support `OBJECT_DELETE` yet. So dont't forget append `-e OBJECT_FINALIZE`.

You can see other eventTypes at https://cloud.google.com/storage/docs/pubsub-notifications#events .
Type `gsutil notification create --help` for more detail.


### Deploy/Run your container

After setup the notification, `blocks-gcs-proxy` works well.

### Tips

#### Get eventTypes from each message

You can pass the `eventType` of the message to your command by using `%{attrs.eventType}`.

For example, the following command shows you them.

```
$ blocks-gcs-proxy echo "%{attrs.eventType}" "%{attrs.bucketId}" "%{attrs.objectId}"

You can see other attributes at https://cloud.google.com/storage/docs/pubsub-notifications#format .


## Examples

- [blocks-gcs-proxy example with Cloud Pub/Sub Notifications](../examples/pubsub_notifications/README.md)
