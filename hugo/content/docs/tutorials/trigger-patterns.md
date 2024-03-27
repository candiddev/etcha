---
categories:
- tutorial
description: How to trigger a simple Pattern using events and webhooks.
title: Trigger Patterns with Events and Webhooks
---

In this tutorial, we'll trigger a Pattern in Etcha using Webhooks and Events.

## Requirements

- Docker or Podman (we'll use Docker here, but this should work with Podman, too)
- Access to pull down Etcha from GitHub's container registry (ghcr.io)
- A text editor

## Tutorial

### Prepare Our Environment

1. Open a local, empty directory in a shell like bash where we can read/write files and mount them into a container.
2. Create a temporary bash alias for Etcha so we can use it:

```bash
alias etcha='docker run --network etcha -u $(id -u):$(id -g) --rm -v $(pwd):/work -w /work ghcr.io/candiddev/etcha:latest'
```

3. Initialize a new container network and directory with Etcha:

```bash
$ docker network create etcha
$ mkdir etcha
$ etcha init .
```

We need to use a custom network so our Etcha containers can communicate with each other.

4. Lets write a Pattern that we'll trigger from an Event and Webhook.  Create a new file under `patterns` called `handler.jsonnet`.  Add in this content:

```
{
  run: [
    {
      always: true,
      change: 'if [[ ${ETCHA_SOURCE_TRIGGER} == "event" ]]; then env > env; else env; fi',
      id: 'write env',
      onChange: [
        'etcha:webhookBody',
      ],
      remove: 'rm env',
    },
  ],
  runExec: {
    command: '/bin/sh -c'
  },
}
```

This Pattern will output the environment to a file `env` when the Pattern is ran via `event`, otherwise it will send it stdout and the event `webhookBody`.

5. Create another new file under `patterns` called `trigger.jsonnet`.  Add in this content:

```
{
  run: [
    {
      always: true,
      change: 'echo hello world',
      id: 'trigger event',
      onChange: [
        'etcha:trigger',
      ],
    },
  ],
  runExec: {
    command: '/bin/sh -c'
  },
}
```

## Running Etcha Listener

1. Lets configure Etcha with a [Source]({{< ref "/docs/guides/running-patterns#sourcesc" >}}) that will handle our Events and Webhooks, and another that allows pushes to trigger Events:

```bash
$ docker run -d --name etcha_listen \
  --network etcha -p 4000:4000 \
  -u $(id -u):$(id -g) \
  -v $(pwd):/work -w /work \
  ghcr.io/candiddev/etcha:latest \
  -x run_systemMetricsSecret=secret \
  -x run_verifyKeys='["ed25519public:MCowBQYDK2VwAyEAw7eTEuEH0+TfgtX3zB+JZVnYD0eskY6qn3n7ZCA7wWM=:reqYEklgP4"]' \
  -x sources='{
      "handler": {
        "allowPush": true,
        "eventsReceive": [
          "trigger"
        ],
        "exec": {
          "allowOverride": true
        },
        "triggerOnly": true,
        "webhookPaths": [
          "/trigger"
        ]
      },
      "trigger": {
        "allowPush": true,
        "eventsSend": ".*",
        "exec": {
          "allowOverride": true
        },
      }
    }' run-listen
```

This config has a lot going on, lets break it down:

- We define two Sources, `handler`, and `trigger`.
- `handler` allows pushes, but nothing will be ran unless via a trigger, like receiving an Event under `eventsReceive`, or a receiving a Webhook request to a `webhookPaths`.
- `trigger` allows pushes, and will always run whatever is pushed, not just a diff of it.  This source is also allowed to send any Event.

The container should've started listening:

```
$ docker logs etcha_listen
level="ERROR" function="etcha/go/pattern/jwt.go:56" status=500 success=false error="error reading JWT: error opening src: error opening src: open /work/etcha/handler.jwt: no such file or directory"
level="ERROR" function="etcha/go/pattern/jwt.go:56" status=500 success=false error="error reading JWT: error opening src: error opening src: open /work/etcha/trigger.jwt: no such file or directory"
level="INFO" function="etcha/go/run/run.go:60" status=200 success=true  message="Starting source runner..."
level="INFO" function="etcha/go/run/run.go:185" status=200 success=true  message="Generating self-signed certificate for listener..."
level="INFO" function="etcha/go/run/run.go:203" status=200 success=true  message="Starting listener..."
```

The errors at the beginning are normal, Etcha can't find an existing JWT for our source.  Etcha also generated a self-signed certificate for us to use.

2. Lets make sure we can access the web interface by pulling some metrics:

```bash
$ curl -sk https://localhost:4000/etcha/v1/system/metrics?key=secret
# HELP go_gc_duration_seconds A summary of the pause duration of garbage collection cycles.
# TYPE go_gc_duration_seconds summary
go_gc_duration_seconds{quantile="0"} 0.000104377
go_gc_duration_seconds{quantile="0.25"} 0.000104377
go_gc_duration_seconds{quantile="0.5"} 0.000104377
go_gc_duration_seconds{quantile="0.75"} 0.000104377
...
```

We should see a bunch of [metrics]({{< ref "/docs/guides/monitoring" >}}).  Nothing interesting yet--we haven't triggered any commands.

3. Lets Push our `handler` Pattern:

```bash
$ etcha -x build_signingKey=ed25519private:MC4CAQAwBQYDK2VwBCIEIBq+BhDRYk8OJv1ksMwKtf0td5p3FGwypXq96gHKefGS:reqYEklgP4 \
    -x build_pushTLSSkipVerify=true push-pattern patterns/handler.jsonnet https://etcha_listen:4000/etcha/v1/push/handler
```

This shouldn't have ran any Commands:
```bash
$ docker logs -n 5 etcha_listen
 $ docker logs -n 5 etcha_listen
level="INFO" function="etcha/go/run/run.go:203" status=200 success=true  message="Starting listener..."
2023/10/14 19:54:55 http: TLS handshake error from 172.19.0.3:57738: remote error: tls: bad certificate
level="INFO" function="etcha/go/run/run.go:97" status=200 success=true path="/etcha/v1/push/handler" sourceAddress="172.19.0.3" sourceName="handler" sourceTrigger="push" sourceName="handler" message="Updating config for handler..."
```

## Triggering Via Events

1. Lets push the Trigger Pattern:

```bash
$ etcha -x build_signingKey=ed25519private:MC4CAQAwBQYDK2VwBCIEIBq+BhDRYk8OJv1ksMwKtf0td5p3FGwypXq96gHKefGS:reqYEklgP4 \
    -x build_pushTLSSkipVerify=true push-pattern patterns/trigger.jsonnet https://etcha_listen:4000/etcha/v1/push/trigger
INFO  etcha/go/run/push.go:37
Pushing config to https://etcha_listen:4000/etcha/v1/push/trigger...
INFO  candiddev/etcha/go/push.go:25
Changed: trigger event
```

The Command `trigger event` range, did it trigger `handler`?

```bash
$ docker logs -n 5 etcha_listen
level="INFO" function="etcha/go/commands/command.go:95" status=200 success=true path="/etcha/v1/push/trigger" sourceAddress="172.19.0.3" sourceName="trigger" sourceTrigger="push" sourceName="trigger" commandID="trigger event" commandMode="check" message="Always changing trigger event..."
level="INFO" function="etcha/go/commands/command.go:129" status=200 success=true path="/etcha/v1/push/trigger" sourceAddress="172.19.0.3" sourceName="trigger" sourceTrigger="push" sourceName="trigger" commandID="trigger event" commandMode="check" commandMode="change" message="Changing trigger event..."
level="INFO" function="etcha/go/run/handlers.go:41" status=200 success=true path="/etcha/v1/push/trigger" sourceAddress="172.19.0.3" sourceName="trigger" sourceTrigger="push" sourceName="trigger" sourceTrigger="event" sourceName="handler" message="Running source handler for event trigger from ID trigger event..."
level="INFO" function="etcha/go/commands/command.go:95" status=200 success=true path="/etcha/v1/push/trigger" sourceAddress="172.19.0.3" sourceName="trigger" sourceTrigger="push" sourceName="trigger" sourceTrigger="event" sourceName="handler" commandID="write env" commandMode="check" message="Always changing write env..."
level="INFO" function="etcha/go/commands/command.go:129" status=200 success=true path="/etcha/v1/push/trigger" sourceAddress="172.19.0.3" sourceName="trigger" sourceTrigger="push" sourceName="trigger" sourceTrigger="event" sourceName="handler" commandID="write env" commandMode="check" commandMode="change" message="Changing write env..."
```

Looks good!  Lets check the env file and see what `ETCHA_` environment variables were set:

```bash
$ cat env | grep ETCHA
ETCHA_EVENT_NAME=trigger
ETCHA_SOURCE_NAME=handler
ETCHA_EVENT_ID=trigger event
ETCHA_SOURCE_TRIGGER=event
ETCHA_EVENT_OUTPUT=hello world
```

We can see the output is the same as the `echo` command in `trigger`.  We can build a more useful Pattern to handle these Events using environment variables.

## Triggering via Webhook

1. Lets fire cURL the Webhook endpoint with some data:

```bash
$  $ curl -H 'x-my-header: hello' -X POST -d 'some data' -k https://localhost:4000/trigger
HOSTNAME=34504152a2b1
SHLVL=1
HOME=/
ETCHA_WEBHOOK_QUERY=
ETCHA_WEBHOOK_METHOD=POST
_CHECK=1
PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
ETCHA_WEBHOOK_PATH=/trigger
ETCHA_WEBHOOK_BODY=c29tZSBkYXRh
ETCHA_SOURCE_NAME=handler
PWD=/work
ETCHA_WEBHOOK_HEADERS=Accept: */*
Content-Length: 9
Content-Type: application/x-www-form-urlencoded
User-Agent: curl/7.88.1
X-My-Header: hello
ETCHA_SOURCE_TRIGGER=webhook
```

There's the environment!  Etcha also added the header we sent, `x-my-header`.  We can base64 decode the `ETCHA_WEBHOOK_BODY` to get the data we sent:

```bash
$ $ base64 -d <<< 'c29tZSBkYXRh'
some data
```

8. Remove the Etcha container and network:

```bash
$ docker rm etcha_listen
$ docker network rm etcha
```

## Summary

We've successfully triggered events and webhooks to execute Patterns.  This concludes the tutorials for now.  You should start looking through the ({{< ref "/docs/guides" >}}).
