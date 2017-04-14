# [config] IMAGE_NAME: "groovenauts/concurrent_batch_devpub_example"
# [config]
# [config] WORKING_DIR: "."
# [config] VERSION_SCRIPT: 'grep VERSION ../../version.go | cut -f2 -d\"'
# [config] GIT_TAG_PREFIX: 'examples/devpub/'

FROM buildpack-deps:jessie-curl

ENV APP_HOME /usr/app/devpub_example
WORKDIR $APP_HOME

ADD pkg/devpub_linux_amd64 $APP_HOME/devpub
ADD config.json $APP_HOME/config.json

ENV BLOCKS_GCS_PROXY_VERSION 0.5.2
RUN mkdir -p /usr/app && \
    curl -L --output ${APP_HOME}/blocks-gcs-proxy \
         https://github.com/groovenauts/blocks-gcs-proxy/releases/download/v${BLOCKS_GCS_PROXY_VERSION}/blocks-gcs-proxy_linux_amd64 && \
    chmod +x ${APP_HOME}/blocks-gcs-proxy

ENV WORKERS 10
ENV TOPIC projects/proj-dummy-999/topics/topic-test

CMD ["./blocks-gcs-proxy", "./devpub", "--filepath", "%{download_files.0}", "--loglevel", "debug"]
