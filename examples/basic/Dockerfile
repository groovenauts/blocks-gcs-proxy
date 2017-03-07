# [config] IMAGE_NAME: "groovenauts/concurrent_batch_basic_example"
# [config]
# [config] WORKING_DIR: "."
# [config] VERSION_SCRIPT: 'grep VERSION ../../version.go | cut -f2 -d\"'
# [config] GIT_TAG_PREFIX: 'examples/basic/'

FROM buildpack-deps:jessie-curl

ENV APP_HOME /usr/app/batch_type_example
COPY . $APP_HOME
WORKDIR $APP_HOME

ENV BLOCKS_GCS_PROXY_VERSION 0.4.0-alpha2
RUN mkdir -p /usr/app && \
    curl -L --output /usr/app/blocks-gcs-proxy \
         https://github.com/groovenauts/blocks-gcs-proxy/releases/download/v${BLOCKS_GCS_PROXY_VERSION}/blocks-gcs-proxy-${BLOCKS_GCS_PROXY_VERSION}_linux-amd64 && \
    chmod +x /usr/app/blocks-gcs-proxy

CMD ["blocks-gcs-proxy", "./app.sh", "%{download_files.0}", "%{downloads_dir}", "%{uploads_dir}", "test"]