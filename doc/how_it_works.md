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

magellan-gcs-proxy builds command to run you application from job message by extending
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

If the value match `/\A\[.*\]\z/` or `/\A\{.*\}\z/`, magellan-gcs-proxy tries to parse as JSON.
When it succeeds magellan-gcs-proxy use it as an array or a map. When it fails magellan-gcs-proxy
use it as a string.


### Run one of multiple commands

If you have to run some commands in a docker container image, you can use `commands` in your `config.yml`.

#### Precondition

You have `config.yml`:

```yaml
commands:
  key1: "cmd1 %{uploads_dir} %{download_files}"
  key2: "cmd2 %{download_files.bar}" %{uploads_dir}
```

And run `magellan-gcs-proxy` like this:
```
magellan-gcs-proxy %{attrs.foo}
```

#### Case 1. foo is key1

```yaml
foo: key1
download_files:
- gs://bucket1/file1
- gs://bucket1/file2
```

`magellan-gcs-proxy` call `cmd1`:

```
cmd1 path/to/workspace/uploads path/to/workspace/downloads/file1 path/to/workspace/downloads/file2
```

#### Case 2. foo is key2

```yaml
foo: key2
download_files:
  foo: gs://bucket1/file1
  bar: gs://bucket1/file1
```

`magellan-gcs-proxy` call `cmd2`:

```
cmd1 path/to/workspace/downloads/file2 path/to/workspace/uploads
```

For more detail, see [config.yml/commands](./configuration.md#commands) also.


## Directories

`magellan-gcs-proxy` makes temporary `workspace` directory.

When a message which has the following attribute:

```
download_files: ["gs://bucket1/path/to/file1"]
```

`magellan-gcs-proxy` makes the following directories and download the file before calling user application.

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

`magellan-gcs-proxy` uploads `workspace/uploads/bucket1/path/to/file-a` to `gs://bucket1/path/to/file-a`
and `workspace/uploads/bucket2/path/to/file-b` to `gs://bucket2/path/to/file-b`.

## Long time job support

If your application takes a long time over [acknowledgement deadline](https://cloud.google.com/pubsub/docs/subscriber#ack_deadline),
`magellan-gcs-proxy` sends `modifyAckDeadline` request to job-subscription automatically.

For more detail, see [config.yml/sustainer](./configuration.md#sustainer) also.
