# blocks-gcs-proxy basic example

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
$ cd examples/basic
$ export PIPELINE=pipeline01
$ export PROJECT=your-gcp-project
$ blocks-gcs-proxy ./app.sh %{download_files.0} %{downloads_dir} %{uploads_dir} test
```

### Terminal 3 Publish message

```
$ export PIPELINE=pipeline01
$ export PROJECT=your-gcp-project
$ export TOPIC="projects/${PROJECT}/topics/${PIPELINE}-job-topic"
$ gcloud beta pubsub topics publish $TOPIC "" --attribute='download_files=["gs://bucket1/path/to/file"]'
messageIds: '49718447408725'
```


## How to execute command on container in local docker

```
$ cd blocks-gcs-proxy/examples/basic

$ export WORKSPACE=$PWD/tmp/workspace
$ export DOWNLOADS_DIR=$WORKSPACE/downloads
$ mkdir -p $DOWNLOADS_DIR/bucket1/dir1
$ echo "test1" > $DOWNLOADS_DIR/bucket1/dir1/test.txt

$ cp test_message.json $WORKSPACE

$ docker run \
    -v $WORKSPACE:/usr/app/batch_type_example/tmp \
    groovenauts/concurrent_batch_basic_example:0.6.2-alpha3 \
    ./blocks-gcs-proxy exec \
        ./app.sh %{download_files.0} %{downloads_dir} %{uploads_dir} test \
        -w /usr/app/batch_type_example/tmp \
        -m /usr/app/batch_type_example/tmp/test_message.json

$ find $WORKSPACE/uploads
$ cat $WORKSPACE/uploads/**/*.*
```

### How to execute command locally

Download `blocks-gcs-proxy` to somewhere.

```
$ path/to/blocks-gcs-proxy exec \
    ./app.sh %{download_files.0} %{downloads_dir} %{uploads_dir} test \
    -w $WORKSPACE \
    -m $WORKSPACE/test_message.json
```
