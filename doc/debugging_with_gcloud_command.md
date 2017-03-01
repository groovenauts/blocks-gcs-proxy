# Debugging with gcloud

## Setup

```
$ export PIPELINE=pipeline01
gcloud deployment-manager deployments create $PIPELINE --config test/pubsub.jinja
```

### Cleanup

```
gcloud deployment-manager deployments list
gcloud deployment-manager deployments delete $PIPELINE
```

## Listen to progress subscription

Download [pubsub-devsub](https://github.com/akm/pubsub-devsub/releases) and put it into the directory on PATH.

Open a new terminal:

```
export PIPELINE=pipeline01
export PROJECT=your-gcp-project
pubsub-devsub --project $PROJECT --subscription "${PIPELINE}-progress-subscription"
```

## Run application

```
export PROJECT=your-gcp-project
export PIPELINE=pipeline01
export JOB_SUB="projects/${PROJECT}/subscriptions/${PIPELINE}-job-subscription"
export PROGRESS_TOPIC="projects/${PROJECT}/topics/${PIPELINE}-progress-topic"
echo "{\"job\":{\"subscription\":\"${JOB_SUB}\"},\"progress\":{\"topic\":\"${PROGRESS_TOPIC}\"}}" > config.json
blocks-gcs-proxy COMMAND ARGS...
```

## Publish message

```
export PROJECT=your-gcp-project
export PIPELINE=pipeline01
export JOB_TOPIC="projects/${PROJECT}/topics/${PIPELINE}-job-topic"
gcloud beta pubsub topics publish $JOB_TOPIC "" --attribute='download_files=["gs://bucket1/path/to/file"]'
```
