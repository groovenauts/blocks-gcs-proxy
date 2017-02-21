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
