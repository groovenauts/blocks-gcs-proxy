# [config] IMAGE_NAME: "groovenauts/concurrent_batch_devpub_example"
# [config]
# [config] WORKING_DIR: "."
# [config] VERSION_SCRIPT: 'grep VERSION ../../version.go | cut -f2 -d\"'
# [config] GIT_TAG_PREFIX: 'examples/devpub/'

FROM buildpack-deps:jessie-curl

ENV APP_HOME /usr/app/devpub_example
WORKDIR $APP_HOME

ENV PUBSUB_DEVPUB_VERSION 0.0.1
RUN mkdir -p /usr/app && \
    curl -L --output ${APP_HOME}/pubsub-devpub \
         https://github.com/groovenauts/pubsub-devpub/releases/download/v${PUBSUB_DEVPUB_VERSION}/pubsub-devpub_linux_amd64 && \
    chmod +x ${APP_HOME}/pubsub-devpub

ENV BLOCKS_GCS_PROXY_VERSION 0.5.2
RUN mkdir -p /usr/app && \
    curl -L --output ${APP_HOME}/blocks-gcs-proxy \
         https://github.com/groovenauts/blocks-gcs-proxy/releases/download/v${BLOCKS_GCS_PROXY_VERSION}/blocks-gcs-proxy_linux_amd64 && \
    chmod +x ${APP_HOME}/blocks-gcs-proxy

ADD config.json $APP_HOME/config.json

ENV WORKERS 10
ENV TOPIC projects/proj-dummy-999/topics/topic-test

CMD ["./blocks-gcs-proxy", "./pubsub-devpub", "--filepath", "%{download_files.0}", "--loglevel", "debug"]
