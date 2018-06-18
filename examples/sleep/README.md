# blocks-gcs-proxy sleep example

## Setup

```
cd path/to/blocks-gcs-proxy/examples/sleep
source ../../.env
make generate
make release

curl -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -X POST http://$AEHOST/orgs/$ORG_ID/pipelines --data @pipeline.json
curl -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' http://$AEHOST/orgs/$ORG_ID/pipelines

export PIPELINE_ID=xxxxxx
curl -v -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -X POST "http://$AEHOST/pipelines/$PIPELINE_ID/jobs?ready=true" --data '{"message": {"attributes": {"sleep_time": "60"}}}'
```
