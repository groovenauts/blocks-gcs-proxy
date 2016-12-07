# magellan-gcs-proxy example

## How to run application locally

```
$ cd path/to/somewhere
$ git clone https://github.com/groovenauts/magellan-gcs-proxy.git
$ cd magellan-gcs-proxy/example
$ bundle
$ export BLOCKS_BATCH_PROJECT_ID=[Your project ID]
$ magellan-gcs-proxy-dev-setup $BLOCKS_BATCH_PROJECT_ID
```

Follow the steps `magellan-gcs-proxy-dev-setup` shows.

## Result

### Terminal 1

```
$ magellan-gcs-proxy-dev-progress-listener $BLOCKS_BATCH_PROGRESS_SUBSCRIPTION

Listening to projects/scenic-doodad-617/subscriptions/test-progress-topic-akm-sub
total:12	job_message_id:49719039110847	progress:1	level:info	Processing message: #<Google::Cloud::Pubsub::ReceivedMessage:0x007f9c6b2127d0 @subscription=#<Google::Cloud::Pubsub::Subscription:0x007f9c6b386f58 @service=Google::Cloud::Pubsub::Service(scenic-doodad-617), @grpc=<Google::Pubsub::V1::Subscription: name: "projects/scenic-doodad-617/subscriptions/test-job-topic-akm-sub", topic: "projects/scenic-doodad-617/topics/test-job-topic-akm", push_config: <Google::Pubsub::V1::PushConfig: push_endpoint: "", attributes: {}>, ack_deadline_seconds: 10>, @name=nil, @exists=nil>, @grpc=<Google::Pubsub::V1::ReceivedMessage: ack_id: "OUVBXkASTDYMRElTK0MLKlgRTgQhIT4wPkVTRFAGFixdRkhRNxkIaFEOT14jPzUgKEUSAAgUBXx9clNcdV8zdQdRDRlzeGkmP1lGU1ARB3ReURsfWVx-SwNYChh7fWJ8a1sTCQZCe4m8xtU_Zhs9XxJLLD5-PQ", message: <Google::Pubsub::V1::PubsubMessage: data: "", attributes: {"download_files"=>"[\"gs://bucket1/dir1/database.yml\"]"}, message_id: "49719039110847", publish_time: <Google::Protobuf::Timestamp: seconds: 1480929118, nanos: 504000000>>>>
total:12	job_message_id:49719039110847	progress:2	level:info	Download starting
total:12	job_message_id:49719039110847	progress:3	level:info	Download completed
total:12	job_message_id:49719039110847	progress:5	level:info	Command starting
total:12	job_message_id:49719039110847	progress:8	level:info	Upload starting
total:12	job_message_id:49719039110847	progress:6	level:info	Command completed
total:12	job_message_id:49719039110847	progress:9	level:info	Upload completed
total:12	job_message_id:49719039110847	progress:11	level:info	Acknowledged
total:12	job_message_id:49719039110847	progress:12	level:info	Cleanup
```

### Terminal 2

```
$ bundle exec magellan-gcs-proxy ./app.sh %{download_files.0} %{uploads_dir}

====================================================================================================
progress_notification:
  topic: projects/scenic-doodad-617/topics/test-progress-topic-akm

loggers:
- type: stdout
- type: cloud_logging
  log_name: magellan-gcs-proxy-example-logging
----------------------------------------------------------------------------------------------------
I, [2016-12-05T18:12:32.892158 #22285]  INFO -- : Start listening
I, [2016-12-05T18:12:34.077139 #22285]  INFO -- : subscription: #<Google::Cloud::Pubsub::Subscription:0x007f9c6b386f58 @service=Google::Cloud::Pubsub::Service(scenic-doodad-617), @grpc=<Google::Pubsub::V1::Subscription: name: "projects/scenic-doodad-617/subscriptions/test-job-topic-akm-sub", topic: "projects/scenic-doodad-617/topics/test-job-topic-akm", push_config: <Google::Pubsub::V1::PushConfig: push_endpoint: "", attributes: {}>, ack_deadline_seconds: 10>, @name=nil, @exists=nil>
I, [2016-12-05T18:12:35.572740 #22285]  INFO -- : job_message_id:49719039110847	progress:1	total:12	data:Processing message: #<Google::Cloud::Pubsub::ReceivedMessage:0x007f9c6b2127d0 @subscription=#<Google::Cloud::Pubsub::Subscription:0x007f9c6b386f58 @service=Google::Cloud::Pubsub::Service(scenic-doodad-617), @grpc=<Google::Pubsub::V1::Subscription: name: "projects/scenic-doodad-617/subscriptions/test-job-topic-akm-sub", topic: "projects/scenic-doodad-617/topics/test-job-topic-akm", push_config: <Google::Pubsub::V1::PushConfig: push_endpoint: "", attributes: {}>, ack_deadline_seconds: 10>, @name=nil, @exists=nil>, @grpc=<Google::Pubsub::V1::ReceivedMessage: ack_id: "OUVBXkASTDYMRElTK0MLKlgRTgQhIT4wPkVTRFAGFixdRkhRNxkIaFEOT14jPzUgKEUSAAgUBXx9clNcdV8zdQdRDRlzeGkmP1lGU1ARB3ReURsfWVx-SwNYChh7fWJ8a1sTCQZCe4m8xtU_Zhs9XxJLLD5-PQ", message: <Google::Pubsub::V1::PubsubMessage: data: "", attributes: {"download_files"=>"[\"gs://bucket1/dir1/database.yml\"]"}, message_id: "49719039110847", publish_time: <Google::Protobuf::Timestamp: seconds: 1480929118, nanos: 504000000>>>>
I, [2016-12-05T18:12:35.785083 #22285]  INFO -- : job_message_id:49719039110847	progress:2	total:12	data:Download starting
D, [2016-12-05T18:12:35.942864 #22285] DEBUG -- : Downloading: gs://bucket1/dir1/database.yml to /var/folders/70/5yxt0jys7ss_m71y5pz_l_qc0000gn/T/workspace20161205-22285-18l5pmo/downloads/dir1/database.yml
I, [2016-12-05T18:12:37.559498 #22285]  INFO -- : Download OK: gs://bucket1/dir1/database.yml to /var/folders/70/5yxt0jys7ss_m71y5pz_l_qc0000gn/T/workspace20161205-22285-18l5pmo/downloads/dir1/database.yml
I, [2016-12-05T18:12:37.796599 #22285]  INFO -- : job_message_id:49719039110847	progress:3	total:12	data:Download completed
I, [2016-12-05T18:12:38.072803 #22285]  INFO -- : job_message_id:49719039110847	progress:5	total:12	data:Command starting
I, [2016-12-05T18:12:38.223959 #22285]  INFO -- : executing: ./app.sh /var/folders/70/5yxt0jys7ss_m71y5pz_l_qc0000gn/T/workspace20161205-22285-18l5pmo/downloads/dir1/database.yml /var/folders/70/5yxt0jys7ss_m71y5pz_l_qc0000gn/T/workspace20161205-22285-18l5pmo/uploads 2>&1
I, [2016-12-05T18:12:38.465296 #22285]  INFO -- : SUCCESS: ./app.sh /var/folders/70/5yxt0jys7ss_m71y5pz_l_qc0000gn/T/workspace20161205-22285-18l5pmo/downloads/dir1/database.yml /var/folders/70/5yxt0jys7ss_m71y5pz_l_qc0000gn/T/workspace20161205-22285-18l5pmo/uploads 2>&1
I, [2016-12-05T18:12:38.859797 #22285]  INFO -- : job_message_id:49719039110847	progress:6	total:12	data:Command completed
I, [2016-12-05T18:12:39.081097 #22285]  INFO -- : job_message_id:49719039110847	progress:8	total:12	data:Upload starting
I, [2016-12-05T18:12:39.205161 #22285]  INFO -- : Uploading: database-20161205-181238.yml to gs://bucket1/database-20161205-181238.yml
I, [2016-12-05T18:12:41.102121 #22285]  INFO -- : Upload OK: database-20161205-181238.yml to gs://bucket1/database-20161205-181238.yml
I, [2016-12-05T18:12:41.346263 #22285]  INFO -- : job_message_id:49719039110847	progress:9	total:12	data:Upload completed
I, [2016-12-05T18:12:41.662179 #22285]  INFO -- : job_message_id:49719039110847	progress:11	total:12	data:Acknowledged
I, [2016-12-05T18:12:41.881730 #22285]  INFO -- : job_message_id:49719039110847	progress:12	total:12	data:Cleanup
```

### Terminal 3 Publish message

```
$  gcloud beta pubsub topics publish $BLOCKS_BATCH_PUBSUB_TOPIC "" --attribute='download_files=["gs://bucket1/dir1/database.yml"]'
messageIds: '49718447408725'
```
