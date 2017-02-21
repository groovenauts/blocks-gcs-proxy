## Debugging with gcloud

### Setup

```
magellan-gcs-proxy-dev-setup [Project ID]
```


### Listen to progress subscription

Download [pubsub-devsub](https://github.com/akm/pubsub-devsub/releases) and put it into the directory on PATH.

```
$ pubsub-devsub --project [Project ID] --subscription [Progress subscription name]
```

`Progress subscription name` is the name of `Progress subscription`.
It created by `magellan-gcs-proxy-dev-setup`.
You can see it by `gcloud beta pubsub subscriptions list`.
It starts with `test-progress-` and ends with '-sub'.

### Run application

```
$ export BLOCKS_BATCH_PROJECT_ID=[Project ID]
$ export BLOCKS_BATCH_PUBSUB_SUBSCRIPTION=[Job subscription name]
$ magellan-gcs-proxy COMMAND ARGS...
```

`Job subscription name` is the name of `Job subscription`.
It created by `magellan-gcs-proxy-dev-setup`.
You can see it by `gcloud beta pubsub subscriptions list`.
It starts with `test-job-` and ends with '-sub'.

### Publish message

```
$ export JOB_TOPIC=[Job topic name]
$ gcloud beta pubsub topics publish $JOB_TOPIC "" --attribute='download_files=["gs://bucket1/path/to/file"]'
```

`Job topic name` is the name of `Job topic`.
It created by `magellan-gcs-proxy-dev-setup`.
You can see it by `gcloud beta pubsub topics list`.
It starts with `test-job-` and ends with your user name.
