# error handling example

## Setup

```bash
$ cd path/to/blocks-gcs-proxy
$ export BLOCKS_GCS_PROXY_VERSION=$(make version)
$ cd examples/error_handling
$ export BLOCKS_GCS_PROXY_VERSION=$(make version)
$ export GCP_PROJECT=[YOUR PROJECT]
$ export PUBSUB_BASE_NAME=test1
$ export PUBSUB_TOPIC=$PUBSUB_BASE_NAME-topic
$ export PUBSUB_SUBSCRIPTION=$PUBSUB_BASE_NAME-sub
$ gcloud --project $GCP_PROJECT pubsub topics create $PUBSUB_TOPIC
$ gcloud --project $GCP_PROJECT pubsub subscriptions create $PUBSUB_SUBSCRIPTION --topic=$PUBSUB_TOPIC
$ gcloud --project $GCP_PROJECT pubsub topics list
$ gcloud --project $GCP_PROJECT pubsub subscriptions list
```

```
$ export BLOCKS_GCS_PROXY_FILE=blocks-gcs-proxy_darwin_amd64
```

```
$ export BLOCKS_GCS_PROXY_FILE=blocks-gcs-proxy_linux_amd64
```

```
$ curl -L --output ./blocks-gcs-proxy https://github.com/groovenauts/blocks-gcs-proxy/releases/download/v${BLOCKS_GCS_PROXY_VERSION}/$BLOCKS_GCS_PROXY_FILE
$ chmod +x ./blocks-gcs-proxy
$ ./blocks-gcs-proxy --version


## Without NackOnError

Start blocks-gcs-proxy with app.rb

```
$ ./blocks-gcs-proxy --config config_without_nack_on_error.json ./app.rb %{attrs.exit_code}
```

Send messages from another terminal.

```
$ gcloud --project $GCP_PROJECT pubsub topics publish projects/$GCP_PROJECT/topics/$PUBSUB_TOPIC --attribute exit_code=0
$ gcloud --project $GCP_PROJECT pubsub topics publish projects/$GCP_PROJECT/topics/$PUBSUB_TOPIC --attribute exit_code=1
```

Each of the messages must show you that app.rb is called once.


## With NackOnError

Start blocks-gcs-proxy with app.rb

```
$ ./blocks-gcs-proxy --config config_with_nack_on_error.json ./app.rb %{attrs.exit_code}
```

Send messages from another terminal.

```
$ gcloud --project $GCP_PROJECT pubsub topics publish projects/$GCP_PROJECT/topics/$PUBSUB_TOPIC --attribute exit_code=0
$ gcloud --project $GCP_PROJECT pubsub topics publish projects/$GCP_PROJECT/topics/$PUBSUB_TOPIC --attribute exit_code=1
```

The 2nd message must show you that app.rb is called repeatedly.
Stop the process by `Ctrl+C`.
