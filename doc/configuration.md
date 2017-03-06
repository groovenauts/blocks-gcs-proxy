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

`magellan-gcs-proxy` doesn't run command if dryrun given.
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

Every `interval` seconds, `magellan-gcs-proxy` sends modify `modifyAckDeadline` to the subscription.
The new deadline will be `delay` seconds later at the time.

See [How it works/Long time job support](https://github.com/groovenauts/magellan-gcs-proxy/blob/features/documents/doc/how_it_works.md#long-time-job-support) also.


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

See [How it works/Progress notification](https://github.com/groovenauts/magellan-gcs-proxy/blob/features/documents/doc/how_it_works.md#progress-notification) also.


## Environment Variables

You can use environment variables in the `config.json` with `{{env "HOME"}}`.
See [the file for test](https://github.com/groovenauts/magellan-gcs-proxy/blob/2104fb374544c78a6c240141bc131559a48fa382/test/config_with_env.json) if you need examples.
