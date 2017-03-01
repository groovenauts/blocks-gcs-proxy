# [config] IMAGE_NAME: "groovenauts/concurrent_batch_example"
# [config]
# [config] WORKING_DIR: "."
# [config] VERSION_SCRIPT: "ruby -r ../lib/magellan/gcs/proxy/version.rb -e 'puts Magellan::Gcs::Proxy::VERSION'"
# [config] GIT_TAG_PREFIX: 'example/'

FROM ruby:2.3

COPY Gemfile* /tmp/
WORKDIR /tmp
RUN bundle install

ENV APP_HOME /usr/app/batch_type_example
COPY . $APP_HOME
WORKDIR $APP_HOME
RUN bundle install --without development

CMD ["bundle", "exec", "magellan-gcs-proxy", "./app.sh", "%{download_files.0}", "%{downloads_dir}", "%{uploads_dir}", "test"]
