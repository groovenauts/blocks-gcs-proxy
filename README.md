# blocks-gcs-proxy

[![Build Status](https://secure.travis-ci.org/groovenauts/blocks-gcs-proxy.png)](https://travis-ci.org/groovenauts/blocks-gcs-proxy)

blocks-gcs-proxy is a proxy for MAGELLAN BLOCKS `concurrent batch board`.


## Installation

Download the file from https://github.com/groovenauts/blocks-gcs-proxy/releases and put it somewhere on PATH.


## Usage

Create `config.json` like this:

```json
{
  "job": {
    "subscription": "projects/proj-dummy-999/subscriptions/pipeline01-job-subscription"
  },
  "progress": {
    "topic": "projects/proj-dummy-999/topics/pipeline01-progress-topic"
  }
}
```

Run `blocks-gcs-proxy` with command

```
blocks-gcs-proxy COMMAND ARGS...
```

`blocks-gcs-proxy` calls `COMMAND` with `ARGS` for each message from Pubsub subscription
specified by `job.subscription` in `config.json`.

## config.json

| Key     | Type | Required | Description  |
|---------|------|----------|--------------|
| job     | map | True |   |
| job.subscription | string | True | The subscription name to pull job messages |
| job.pull_interval | int | False | The interval time in second to pull when it gets no job message. Default 0 |
| job.sustainer     | map | False |
| job.sustainer.delay | int | False | The new deadline in second to extend deadline to ack |
| job.sustainer.interval | int | False | The interval in second to send the message which extends deadline to ack |
| progress | map | True |
| progress.topic | string | True | The topic name to publish job progress messages |
| progress.level | string | False | Log level to publish job progress. You can set one of `debug`, `info`, `warn`, `error`, `fatal` and `panic`. |
| log       | map    | False | |
| log.level | string | False | Log level of processing of `blocks-gcs-proxy`. You can set one of `debug`, `info`, `warn`, `error`, `fatal` and `panic`. |
| command   | map | False |  |
| command.dryrun | bool | False | Don't run the command if this is true. |
| command.downloaders | int | False | The number of thread to download. Default: 1.|
| command.uploaders | int | False | The number of thread to upload. Default: 1.|
| command.options | map[key][]string | False | Define if you have to run one of multiple command. See [Multiple command options](#multiple-command-options) for more detail. |


### Multiple command options

If you have commands data in your config.json like the following:

```json
{
  "commands": {
    "options": {
      "key1": ["cmd1", "%{download_files}"],
      "key2": ["cmd2", "%{download_files.bar}", "%{uploads_dir}", "%{download_files.baz}"]
    }
  }
}
```

And when you run the command by
```
$ bundle exec magellan-gcs-proxy %{attrs.foo}
```

you can choose which command is executed by message attribute named `foo` at runtime.

| message attributes | command `magellan-gcs-proxy` calls  |
|--------------------|----------------------|
| `{"foo": "key1"}`  | `cmd1 %{download_files}` |
| `{"foo": "key2"}`  | `cmd2 %{download_files.bar} %{uploads_dir} %{download_files.baz}` |

If the attribute value is not defined in commands keys, the message is ignored with error message.

## blocks-gcs-proxy check

Check the `config.json` is valid.
You can give other file with `--config` or `-c` option.

```bash
$ ./blocks-gcs-proxy check -c config2.json
Error to load config.json.bak cause of invalid character '}' looking for beginning of object key string
```


## blocks-gcs-proxy download

```bash
NAME:
   blocks-gcs-proxy download - Download the files from GCS to downloads directory

USAGE:
   blocks-gcs-proxy download [command options] [arguments...]

OPTIONS:
   --downloads_dir value, -d value  Path to the directory which has bucket_name/path/to/file
   --downloaders value, -n value    Number of downloaders (default: 6)
$ ./blocks-gcs-proxy download --help

```

### Example

```bash
$ ./blocks-gcs-proxy download -d tmp/downloads -n 5 gs://bucket1/path/to/file1  gs://bucket1/path/to/file2  gs://bucket1/path/to/file3
```


## blocks-gcs-proxy upload

```bash
$ ./blocks-gcs-proxy upload --help
NAME:
   blocks-gcs-proxy upload - Upload the files under uploads directory

USAGE:
   blocks-gcs-proxy upload [command options] [arguments...]

OPTIONS:
   --uploads_dir value, -d value  Path to the directory which has bucket_name/path/to/file
   --uploaders value, -n value    Number of uploaders (default: 6)
```

### Example

```bash
$ ./blocks-gcs-proxy upload -d tmp/uploads -n 5
```
