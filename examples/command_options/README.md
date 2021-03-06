# blocks-gcs-proxy basic example

## How to run application locally

```
$ cd path/to/workspace
$ git clone https://github.com/groovenauts/blocks-gcs-proxy.git
$ cd blocks-gcs-proxy
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

Download [blocks-gcs-proxy](https://github.com/groovenauts/blocks-gcs-proxy/releases) and put it into the directory on PATH.

```
$ cd path/to/workspace/blocks-gcs-proxy
$ cd examples/command_options
$ export PIPELINE=pipeline01
$ export PROJECT=your-gcp-project
$ blocks-gcs-proxy %{attrs.cmd}
```

### Terminal 3 Publish message

```
$ export PIPELINE=pipeline01
$ export PROJECT=your-gcp-project
$ export TOPIC="projects/${PROJECT}/topics/${PIPELINE}-job-topic"
$ gcloud beta pubsub topics publish $TOPIC "" --attribute='download_files=["gs://bucket1/path/to/file"]'
$ gcloud beta pubsub topics publish $TOPIC "" --attribute='cmd=sleep,sleep=60'
```
