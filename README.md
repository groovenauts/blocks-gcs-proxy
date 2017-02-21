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

## How it works

### Overview

1. Receive a job message from job subscription of cloud pub/sub
2. Make temporary `workspace` directory to process the job
2. Download the files specified in the job message form Google Cloud Storage into `downloads` directory under `workspace` directory
3. Build command to run your application
4. If your application returns exit code `0`...
    1. Upload the files which are created under `uploads` directory
    2. Send acknowledgement message to pipeline-job-subscription
    3. Clean up `workspace` directory
5. If your application returns exit code not `0`...
    1. Clean up `workspace` directory

### Build command

magellan-gcs-proxy builds command to run you application from job message by extending
parameters in `%{...}`.

#### Parameters

| Parameter     | Type          | Description |
|---------------|---------------|-------------|
| workspace     | string        | The workspace directory for the job |
| downloads_dir | string        | The `downloads` directory under `workspace` |
| uploads_dir   | string        | The `uploads` directory under `workspace` |
| download_files/local_download_files | array or map | The donwloaded file names on local |
| remote_download_files | array or map | The donwloaded file names on GCS |
| attrs/attributes | map    | The attributes of the job message |
| data             | string | The data of the job message |

#### Array Parameter

When the job message has an attribute named `array_data`:

```json
[
  "foo",
  "bar",
  "baz"
]
```

You can specify an element of an array by `attrs.array_data.INDEX`. `INDEX` must be numeric charactors.

| description          | result     |
|----------------------|------------|
| `attrs.array_data.0` | "foo"      |
| `attrs.array_data.1` | "bar"      |
| `attrs.array_data.2` | "baz"      |

If you can use `attrs.array_data` also. It extends the data joined with spaces.

| description        | result        |
|--------------------|---------------|
| `attrs.array_data` | "foo bar baz" |


#### Map Parameter

When the job message has an attribute named `map_data` :

```json
{
  "A": "foo",
  "B": "bar",
  "C": "baz"
}
```

You can specify an element of an array by `attrs.map_data.KEY`.

| description        | result   |
|--------------------|----------|
| `attrs.map_data.A` | "foo"    |
| `attrs.map_data.B` | "bar"    |
| `attrs.map_data.C` | "baz"    |

If you can use `attrs.map_data` also. It extends the values joined with spaces.

| description      | result        |
|------------------|---------------|
| `attrs.map_data` | "foo bar baz" |


#### Recognizing attribute as array or hash

If the value match `/\A\[.*\]\z/` or `/\A\{.*\}\z/`, magellan-gcs-proxy tries to parse as JSON.
When it succeeds magellan-gcs-proxy use it as an array or a map. When it fails magellan-gcs-proxy
use it as a string.


## Debugging with gcloud

### Setup

```
magellan-gcs-proxy-dev-setup [Project ID]
```


### Listen to progress subscription

Download [pubsub-devsub](https://github.com/akm/pubsub-devsub/releases) and put it into the directory on PATH.

```
$ pubsub-devsub --project [Project ID] --subscription [Progress subscription name]
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
