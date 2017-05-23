# [config] IMAGE_NAME: "groovenauts/concurrent_batch_pubsub_notifications_example"
# [config]
# [config] WORKING_DIR: "."
# [config] VERSION_SCRIPT: 'grep VERSION ../../version.go | cut -f2 -d\"'
# [config] GIT_TAG_PREFIX: 'examples/pubsub_notifications/'

FROM buildpack-deps:jessie-curl

ENV APP_HOME /usr/app/batch_type_example
COPY . $APP_HOME
WORKDIR $APP_HOME

ENV BLOCKS_GCS_PROXY_VERSION 0.6.0
RUN mkdir -p /usr/app && \
    curl -L --output ${APP_HOME}/blocks-gcs-proxy \
         https://github.com/groovenauts/blocks-gcs-proxy/releases/download/v${BLOCKS_GCS_PROXY_VERSION}/blocks-gcs-proxy_linux_amd64 && \
    chmod +x ${APP_HOME}/blocks-gcs-proxy

CMD ["./blocks-gcs-proxy", "echo", "%{attrs.eventType}", "gs://%{attrs.bucketId}/%{attrs.objectId}", "%{download_files.0}", "%{downloads_dir}"]
