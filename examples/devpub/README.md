# block-gcs-proxy devpub example

## Setup blocks-concurrent-batch-agent

See https://github.com/groovenauts/blocks-concurrent-batch-agent

## Setup pipeline

Make `pipeline.json` like this:

```json
{
  "name":"devpub-pipeline01",
  "project_id":"proj-dummy-999",
  "zone":"us-central1-f",
  "source_image":"https://www.googleapis.com/compute/v1/projects/google-containers/global/images/gci-stable-55-8872-76-0",
  "machine_type":"f1-micro",
  "target_size":1,
  "container_size":1,
  "container_name":"groovenauts/concurrent_batch_devpub_example:0.0.1",
  "command":"",
  "run_options": [
    "-e", "TOPIC=projects/proj-dummy-999/topics/test-topic01"
  ]
}
```

```bash
$ export TOKEN=....
$ export AEHOST=....
$ curl -v -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -X POST http://$AEHOST/pipelines --data @pipeline.json
$ curl -H "Authorization: Bearer $TOKEN" http://$AEHOST/pipelines
$ curl -H "Authorization: Bearer $TOKEN" http://$AEHOST/pipelines/refresh
$ curl -H "Authorization: Bearer $TOKEN" http://$AEHOST/pipelines
```


### Teardown pipeline

```
$ export ID="[id of the result]"
$ curl -v -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -X PUT http://$AEHOST/pipelines/$ID/close --data ""
```

```
$ curl -H "Authorization: Bearer $TOKEN" http://$AEHOST/pipelines/refresh
$ curl -H "Authorization: Bearer $TOKEN" http://$AEHOST/pipelines
```

Wait until the pipeline's status becomes 7.


```
$ curl -v -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -X DELETE http://$AEHOST/pipelines/$ID
$ curl -H "Authorization: Bearer $TOKEN" http://$AEHOST/pipelines
```

## Generate and upload jsonl file

```
$ export BUCKET=bucket1
$ ruby -r json -e 'tmpl = JSON.generate({"attributes":{"download_files":"[\"gs://bucket1/path/to/file%06d\"]"}}); (1..1000).each{|i| puts tmpl % i}' > test1.jsonl
$ ruby -r json -e 'tmpl = JSON.generate({"attributes":{"download_files":"[\"gs://bucket1/path/to/file%06d\"]"}}); (1001..2000).each{|i| puts tmpl % i}' > test2.jsonl
$ ruby -r json -e 'tmpl = JSON.generate({"attributes":{"download_files":"[\"gs://bucket1/path/to/file%06d\"]"}}); (2001..3000).each{|i| puts tmpl % i}' > test3.jsonl
$ gsutil cp ./test*.jsonl gs://$BUCKET/
```

## Publish job messages

```
$ gcloud beta pubsub topics publish devpub-pipeline01-job-topic '' --attribute='download_files=["gs://'$BUCKET'/test1.jsonl"]'
$ gcloud beta pubsub topics publish devpub-pipeline01-job-topic '' --attribute='download_files=["gs://'$BUCKET'/test2.jsonl"]'
$ gcloud beta pubsub topics publish devpub-pipeline01-job-topic '' --attribute='download_files=["gs://'$BUCKET'/test3.jsonl"]'
```
