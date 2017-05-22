# blocks-gcs-proxy example with Cloud Pub/Sub Notifications

See [Use blocks-gcs-proxy with Cloud Pub/Sub Notifications](../../doc/pubsub_notification.md)
for more detail about background.

## How to run application locally

```
$ cd path/to/workspace
$ git clone https://github.com/groovenauts/blocks-gcs-proxy.git
$ cd blocks-gcs-proxy
$ export PIPELINE=pipeline01
$ gcloud deployment-manager deployments create $PIPELINE --config test/pubsub.jinja
```

### Terminal 1

Download [pubsub-devsub](https://github.com/akm/pubsub-devsub/releases) and put it into the directory on PATH.

```
$ export PIPELINE=pipeline01
$ export PROJECT=your-gcp-project
$ pubsub-devsub --project $PROJECT --subscription "${PIPELINE}-progress-subscription"
```

### Terminal 2

Download [blocks-gcs-proxy](https://github.com/groovenauts/blocks-gcs-proxy/releases) and put it into the directory on PATH.

```
$ cd path/to/workspace/blocks-gcs-proxy
$ cd examples/pubsub_notifications
$ export PIPELINE=pipeline01
$ export PROJECT=your-gcp-project
$ blocks-gcs-proxy echo %{attrs.eventType} gs://%{attrs.bucketId}/%{attrs.objectId} %{download_files.0} %{downloads_dir}
```

### Terminal 3 Setup notification and upload files

```
$ export PIPELINE=pipeline01
$ export PROJECT=your-gcp-project
$ export TOPIC="projects/${PROJECT}/topics/${PIPELINE}-job-topic"
$ export BUCKET="your-bucket"
$ gsutil notification create -t $TOPIC -f json -e OBJECT_FINALIZE gs://$BUCKET
$ gsutil cp [path/to/localFile] $BUCKET/path/to/remoteFile
```

gsutil notification create -t projects/scenic-doodad-617/topics/akm-pipeline01-job-topic                -f json gs://blocks-gcs-proxy-test -e OBJECT_FINALIZE
gsutil notification create -t projects/scenic-doodad-617/topics/akm-test-gcs-pubsub-notifications-topic -f json gs://akm-test
