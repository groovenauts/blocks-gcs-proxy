# How it works

## Overview

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

## Build command

blocks-gcs-proxy builds command to run you application from job message by extending
parameters in `%{...}`.

### Parameters

| Parameter     | Type          | Description |
|---------------|---------------|-------------|
| workspace     | string        | The workspace directory for the job |
| downloads_dir | string        | The `downloads` directory under `workspace` |
| uploads_dir   | string        | The `uploads` directory under `workspace` |
| download_files/local_download_files | array or map | The donwloaded file names on local |
| remote_download_files | array or map | The donwloaded file names on GCS |
| attrs/attributes | map    | The attributes of the job message |
| data             | string | The data of the job message |

### Array Parameter

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

See an example at [example/Dockerfile](https://github.com/groovenauts/blocks-gcs-proxy/blob/4bb4d0f60d1e62ba0e06b2180ec5835726e0fc57/example/Dockerfile#L18).


### Map Parameter

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


### Recognizing attribute as array or hash

If the value match `/\A\[.*\]\z/` or `/\A\{.*\}\z/`, blocks-gcs-proxy tries to parse as JSON.
When it succeeds blocks-gcs-proxy use it as an array or a map. When it fails blocks-gcs-proxy
use it as a string.


### Run one of multiple commands

If you have to run some commands in a docker container image, you can use `command/options` in your `config.json`.

#### Precondition

You have `config.json`:

```json
{
  "command": {
    "options": {
      "key1": ["cmd1", "%{uploads_dir}", "%{download_files}"],
      "key2": ["cmd2", "%{download_files.b}", "%{uploads_dir}"]
    }
  }
}
```

And run `blocks-gcs-proxy` like this:
```
blocks-gcs-proxy %{attrs.foo}
```

#### Case 1. foo is key1

```json
{
  "foo": "key1",
  "download_files": [
    "gs://bucket1/file1",
    "gs://bucket1/file2",
    "gs://bucket1/file3"
  ]
}
```

When `blocks-gcs-proxy` gets the message above, it calls `cmd1`:

```
cmd1 path/to/workspace/uploads path/to/workspace/downloads/file1 path/to/workspace/downloads/file2 path/to/workspace/downloads/file3
```

#### Case 2. foo is key2

```json
{
  "foo": "key2",
  "download_files": {
    "a": "gs://bucket1/file1",
    "b": "gs://bucket1/file2",
    "c": "gs://bucket1/file3"
  }
}
```

When `blocks-gcs-proxy` gets the message above, it calls `cmd2`:

```
cmd2 path/to/workspace/downloads/file2 path/to/workspace/uploads
```

For more detail, see [config.json/command](./configuration.md#commandoptions) also.


## Directories

`blocks-gcs-proxy` makes temporary `workspace` directory.

When a message which has the following attribute:

```
download_files: ["gs://bucket1/path/to/file1"]
```

`blocks-gcs-proxy` makes the following directories and download the file before calling user application.

```
workspace/
workspace/downloads/
workspace/downloads/bucket1/
workspace/downloads/bucket1/path/
workspace/downloads/bucket1/path/to/
workspace/downloads/bucket1/path/to/file1
workspace/uploads/
```

If user application makes the following directories and files:

```
workspace/
workspace/downloads/
(snip)
workspace/uploads/
workspace/uploads/bucket1/
workspace/uploads/bucket1/path/
workspace/uploads/bucket1/path/to/
workspace/uploads/bucket1/path/to/file-a
workspace/uploads/bucket2/
workspace/uploads/bucket2/path/
workspace/uploads/bucket2/path/to/
workspace/uploads/bucket2/path/to/file-b
```

`blocks-gcs-proxy` uploads `workspace/uploads/bucket1/path/to/file-a` to `gs://bucket1/path/to/file-a`
and `workspace/uploads/bucket2/path/to/file-b` to `gs://bucket2/path/to/file-b`.

## Long time job support

If your application takes a long time over [acknowledgement deadline](https://cloud.google.com/pubsub/docs/subscriber#ack_deadline),
`blocks-gcs-proxy` sends `modifyAckDeadline` request to job-subscription automatically.

For more detail, see [config.yml/sustainer](./configuration.md#jobsustainer) also.


## Progress notification

`blocks-gcs-proxy` sends a message for each progress to the topic if given.

To set the topic, see [config.yml/progress](./configuration.md#progress) also.

The message has the following attributes:

| Attribute name | Value  |
|----------------|--------|
| progress       | number from 1 to 14 |
| total          | 14                  |
| completed      | "false" or "true"   |
| job_message_id | Job message ID      |
| level          | "INFO"              |

### Progresses

| Value | data of message |
|------:|-----------------|
|     1 | "Processing message: <message inspect>" |
|     2 | "Download starting"  |
|     3 | "Download completed" |
|     4 | "Download error: <error detail>"
|     5 | "Command starting"  |
|     6 | "Command completed" |
|     7 | "Command error: <error detail>"
|     8 | "Upload starting"  |
|     9 | "Upload completed" |
|    10 | "Upload error: <error detail>"
|    11 | "Acknowledge starting"  |
|    12 | "Acknowledge completed" |
|    13 | "Acknowledge error: <error detail>"
|    14 | "Cleanup" |
