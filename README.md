# magellan-gcs-proxy

[![Gem Version](https://badge.fury.io/rb/magellan-gcs-proxy.png)](https://rubygems.org/gems/magellan-gcs-proxy) [![Build Status](https://secure.travis-ci.org/groovenauts/magellan-gcs-proxy.png)](https://travis-ci.org/groovenauts/magellan-gcs-proxy)

magellan-gcs-proxy is a gem for MAGELLAN BLOCKS "Batch type" IoT board.


## Installation

Add this line to your application's Gemfile:

```ruby
gem 'magellan-gcs-proxy'
```

And then execute:

    $ bundle

Or install it yourself as:

    $ gem install magellan-gcs-proxy

## Usage

```
magellan-gcs-proxy COMMAND ARGS...
```

### TODO How it works

### TODO Expanding variables


## Debugging with gcloud

### Setup

```
magellan-gcs-proxy-dev-setup [Project ID]
```


### Listen to progress subscription

```
$ export BLOCKS_BATCH_PROJECT_ID=[Project ID]
$ magellan-gcs-proxy-dev-progress-listener [Progress subscription name]
```

`Progress subscription name` is the name of `Progress subscription`.
It created by `magellan-gcs-proxy-dev-setup`.
You can see it by `gcloud beta pubsub subscriptions list`.
It starts with `test-progress-` and ends with '-sub'.

### Run application

```
$ export BLOCKS_BATCH_PROJECT_ID=[Project ID]
$ export BLOCKS_BATCH_PUBSUB_SUBSCRIPTION=[Job subscription name]
$ magellan-gcs-proxy COMMAND ARGS...
```

`Job subscription name` is the name of `Job subscription`.
It created by `magellan-gcs-proxy-dev-setup`.
You can see it by `gcloud beta pubsub subscriptions list`.
It starts with `test-job-` and ends with '-sub'.

### Publish message

```
$ export JOB_TOPIC=[Job topic name]
$ gcloud beta pubsub topics publish $JOB_TOPIC "" --attribute='download_files=["gs://bucket1/path/to/file"]'
```

`Job topic name` is the name of `Job topic`.
It created by `magellan-gcs-proxy-dev-setup`.
You can see it by `gcloud beta pubsub topics list`.
It starts with `test-job-` and ends with your user name.


## Development

After checking out the repo, run `bin/setup` to install dependencies. Then, run `rake spec` to run the tests. You can also run `bin/console` for an interactive prompt that will allow you to experiment.

To install this gem onto your local machine, run `bundle exec rake install`. To release a new version, update the version number in `version.rb`, and then run `bundle exec rake release`, which will create a git tag for the version, push git commits and tags, and push the `.gem` file to [rubygems.org](https://rubygems.org).

## Contributing

Bug reports and pull requests are welcome on GitHub at https://github.com/[USERNAME]/magellan-gcs-proxy. This project is intended to be a safe, welcoming space for collaboration, and contributors are expected to adhere to the [Contributor Covenant](http://contributor-covenant.org) code of conduct.


## License

The gem is available as open source under the terms of the [MIT License](http://opensource.org/licenses/MIT).
