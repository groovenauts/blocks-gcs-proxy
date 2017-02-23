# config.yml

## commands

Use `commands` config to choose a command from multiple commands at runtime.
`commands` must be a map like this:

```yaml
commands:
  key1: "cmd1 %{uploads_dir} %{download_files}"
  key2: "cmd2 %{download_files.bar}" %{uploads_dir}
```

See [How it works/Run one of multiple commands](./how_it_works.md#run-one-of-multiple-commands) also.

## dryrun

`magellan-gcs-proxy` doesn't run command if dryrun given.
`true`, `yes`, `on` or `1` are true. The others are false.


## loggers

If you need logging, use `loggers` configuration.
It must be an array of map which has `type`.

`stdout` type logger outputs logs to stdout.
`stderr` type logger outputs logs to stderr.
`cloud_logging` type logger outputs logs to cloud logging named by `log_name`.

```yaml
loggers:
- type: stdout
- type: cloud_logging
  log_name: pipeline01-log-name
```

## progress_notification

If you need to notify the progress of job to progress-topic, use `progress_notification` configuration.
It must be a map which has `topic`.

```yaml
progress_notification:
  topic: topic_name
```

See [How it works/Progress notification](https://github.com/groovenauts/magellan-gcs-proxy/blob/features/documents/doc/how_it_works.md#progress-notification) also.


## sustainer

There are `delay` and `interval` configurations for sustainer to delay ack deadline for long time job support.

```yaml
sustainer:
  delay: 6000
  interval: 540
```

Every `interval` seconds, `magellan-gcs-proxy` sends modify `modifyAckDeadline` to the subscription.
The new deadline will be `delay` seconds later at the time.

See [How it works/Long time job support](https://github.com/groovenauts/magellan-gcs-proxy/blob/features/documents/doc/how_it_works.md#long-time-job-support) also.
