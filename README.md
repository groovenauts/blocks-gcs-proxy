# magellang-gcs-proxy

magellang-gcs-proxy is a gem for MAGELLAN BLOCKS "Batch type" IoT board.


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
export PROJECT_ID=[Project ID]
export TOPIC=projects/$PROJECT_ID/topics/[Topic name]
export SUB=projects/$PROJECT_ID/subscriptions/[Subscription name of topic]
$ gcloud beta pubsub topics create projects/$PROJECT_ID/topics/$TOPIC
$ gcloud beta pubsub topics list
$ gcloud beta pubsub subscriptions create $SUB --topic $TOPIC
$ gcloud beta pubsub subscriptions list
```

### Publish message

```
$ gcloud beta pubsub topics publish $TOPIC "" --attribute='download_files=["gs://bucket1/path/to/file"]'
```


### Run application

```
$ magellan-gcs-proxy COMMAND ARGS...
```


## Development

After checking out the repo, run `bin/setup` to install dependencies. Then, run `rake spec` to run the tests. You can also run `bin/console` for an interactive prompt that will allow you to experiment.

To install this gem onto your local machine, run `bundle exec rake install`. To release a new version, update the version number in `version.rb`, and then run `bundle exec rake release`, which will create a git tag for the version, push git commits and tags, and push the `.gem` file to [rubygems.org](https://rubygems.org).

## Contributing

Bug reports and pull requests are welcome on GitHub at https://github.com/[USERNAME]/magellan-gcs-proxy. This project is intended to be a safe, welcoming space for collaboration, and contributors are expected to adhere to the [Contributor Covenant](http://contributor-covenant.org) code of conduct.


## License

The gem is available as open source under the terms of the [MIT License](http://opensource.org/licenses/MIT).

