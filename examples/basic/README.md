# magellan-gcs-proxy basic example

## How to run application locally

```
$ cd path/to/workspace
$ git clone https://github.com/groovenauts/magellan-gcs-proxy.git
$ cd magellan-gcs-proxy
$ bundle
$ export PIPELINE=akm-pipeline01
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

```
$ cd path/to/workspace/magellan-gcs-proxy
$ cd examples/basic
$ export PIPELINE=pipeline01
$ export PROJECT=your-gcp-project
$
$ blocks-gcs-proxy ./app.sh %{download_files.0} %{uploads_dir}
```

### Terminal 3 Publish message

```
$ export PIPELINE=pipeline01
$ export PROJECT=your-gcp-project
$ export TOPIC="projects/${PROJECT}/topics/${PIPELINE}-job-topic"
$ gcloud beta pubsub topics publish $TOPIC "" --attribute='download_files=["gs://bucket1/path/to/file"]'
messageIds: '49718447408725'
```