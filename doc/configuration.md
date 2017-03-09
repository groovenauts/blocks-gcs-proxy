# config.yml

## Elements

### command/options

Use `command/options` config to choose a command from multiple commands at runtime.
`command/options` must be a map like this:

```json
{
  "command": {
    "options" : {
      "key1": ["cmd1", "%{uploads_dir}", "%{download_files}"],
      "key2": ["cmd2", "%{download_files.bar}", "%{uploads_dir}"]
    }
  }
}
```

See [How it works/Run one of multiple commands](./how_it_works.md#run-one-of-multiple-commands) also.

### command/dryrun

`blocks-gcs-proxy` doesn't run command if dryrun given.
`true`, `yes`, `on` or `1` are true. The others are false.

### job

```json
{
  "job": {
    "subscription": "projects/dummy-gcp-proj/subscriptions/test-job-subscription",
    "pull_interval": 60,
    "sustainer": {
      "delay": 600,
      "interval": 540
    }
  }
}
```

`subscription` is the name of the job subscription which receives the job message from job topic.
`pull_interval` is the number of seconds of interval between pulling job messages.


### job/sustainer

There are two configurations `delay` and `interval` for sustainer to delay ack deadline for long time job support.

Every `interval` seconds, `blocks-gcs-proxy` sends modify `modifyAckDeadline` to the subscription.
The new deadline will be `delay` seconds later at the time.

See [How it works/Long time job support](https://github.com/groovenauts/blocks-gcs-proxy/blob/features/documents/doc/how_it_works.md#long-time-job-support) also.


### progress

If you need to notify the job progresses to progress-topic, use `progress` configuration.
It must be a map which has `topic`.

```json
{
  "progress": {
    "topic": "topic_name"
  }
}
```

See [How it works/Progress notification](https://github.com/groovenauts/blocks-gcs-proxy/blob/features/documents/doc/how_it_works.md#progress-notification) also.


## Environment Variables

You can use environment variables in the `config.json` with `{{env "HOME"}}`, `{{ .HOME }}` or `{{ or .HOME default}}`.

### Examples

- [config file wih env](https://github.com/groovenauts/blocks-gcs-proxy/blob/2104fb374544c78a6c240141bc131559a48fa382/test/config_with_env.json)
- [dot style config file](https://github.com/groovenauts/blocks-gcs-proxy/blob/5d4624048aafd743f3c49cd2a4eaede026a73552/test/config_with_env2.json)
- [dot style config file with default values](https://github.com/groovenauts/blocks-gcs-proxy/blob/5257ace08db7fbd860b6a20795ae887892077f18/test/config_with_env_and_default3.json)
