# [config] IMAGE_NAME: "groovenauts/concurrent_batch_command_options_example"
# [config]
# [config] WORKING_DIR: "."
# [config] VERSION_SCRIPT: 'grep VERSION ../../version.go | cut -f2 -d\"'
# [config] GIT_TAG_PREFIX: 'examples/command_options/'

FROM buildpack-deps:jessie-curl

ENV APP_HOME /usr/app/batch_type_example
COPY . $APP_HOME
WORKDIR $APP_HOME

ENV BLOCKS_GCS_PROXY_VERSION 0.4.0-alpha2
RUN mkdir -p /usr/app && \
    curl -L --output ${APP_HOME}/blocks-gcs-proxy \
         https://github.com/groovenauts/blocks-gcs-proxy/releases/download/${BLOCKS_GCS_PROXY_VERSION}/blocks-gcs-proxy_linux_amd64 && \
    chmod +x ${APP_HOME}/blocks-gcs-proxy

CMD ["./blocks-gcs-proxy", "%{attrs.cmd}"]
