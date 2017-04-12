# block-gcs-proxy devpub example

## Setup blocks-concurrent-batch-agent

See https://github.com/groovenauts/blocks-concurrent-batch-agent

## Setup pipeline

Make `pipeline.json` like this:

```
{
  "name":"devpub-pipeline01",
  "project_id":"proj-dummy-999",
  "zone":"us-central1-f",
  "source_image":"https://www.googleapis.com/compute/v1/projects/google-containers/global/images/gci-stable-55-8872-76-0",
  "machine_type":"f1-micro",
  "target_size":1,
  "container_size":1,
  "container_name":"groovenauts/concurrent_batch_devpub_example:0.0.1",
  "command":""
}
```
