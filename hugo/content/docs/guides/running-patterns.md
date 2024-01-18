---
categories:
- guide
description: How to run Patterns in Etcha.
title: Running Patterns
weight: 60
---

In this guide, we'll go over running Patterns via CLI, push, and pull.

## Run Process

Given a Pattern like this:

```
// patterns/myapp.jsonnet

local app_name = 'myapp';
local restart = function(name)
  {
    change: 'systemctl restart %s' % name,
    id: 'restart %s' % name,
  };

{
  build: [
    {
      always: true,
      change: 'make %s' % app_name,
      id: 'build %s' % app_name,
      onChange: [
        'etcha:buildManifest',
        'etcha:runEnv_myapp',
      ]
    },
  ],
  buildExec: {
    command: 'sudo /bin/bash -c'
  },
  subject: app_name,
  run: [
    {
      change: 'curl -L https://s3.example.com/myapp_v2 -o /myapp',
      check: "myapp --version | grep v2",
      id: "copy %s v2' app_name,
      onChange: [
        'restart myapp',
      ],
    },
    restart(app_name),
  ],
  runExec: {
    command: 'sudo /bin/bash -c'
  },
  runEnv: {
    hello: 'world',
  }
}
```

Etcha will always perform the following during the run phase:

1. Import the Pattern files, either from the [JWT property etchaPattern]({{< ref "/docs/references/JWT" >}}) or local files.
2. Render the Pattern, executing any [Native Functions]({{< ref "/docs/references/jsonnet#native-functions" >}}).
3. Set [Environment Variables]({{< ref "/docs/references/commands#environment-variables" >}}) for the [Commands]({{< ref "/docs/references/commands" >}}) from the `runEnv` property.
4. Use the configured [`buildExec`]({{< ref "/docs/references/config#exec" >}}) or [`runExec`]({{< ref "/docs/references/config#exec" >}}) if specified and allowed by [`allowedOverride`]({{< ref "/docs/references/config#allowoverride" >}}).
5. Execute the `run` [Commands]({{< ref "/docs/references/commands" >}}) in the order specified.

## Local Run via CLI

Etcha can run Patterns via one-off applies using the [CLI]({{< ref "/docs/references/cli" >}}).  These runs will always be local to the current Etcha instance.  They're performed using the following CLI commands:

- [`etcha local change [pattern path]`]({{< ref "/docs/references/cli#local" >}}), executes the `run` in [**Change Mode**]({{< ref "/docs/references/commands#change-mode" >}}).
- [`etcha local remove [pattern path]`]({{< ref "/docs/references/cli#local" >}})), executes the `run` in [**Remove Mode**]({{< ref "/docs/references/commands#remove-mode" >}}).

## Remote Run

Etcha can push and pull [Pattern]({{< ref "/docs/references/patterns" >}}) [JWTs]({{< ref "/docs/references/jwt" >}}) to/from remote instances.  In order for this to work, we need to have a way to verify the JWT, and the remote Etcha instance needs to be configured to accept JWTs through a particular method via [`sources`]({{< ref "/docs/references/config#sources" >}}).

### Listen Mode

Etcha needs to be running in listen mode to receive pushes, execute webhooks, and handle events.  You'll typically run Etcha in listen mode as a container or via a service manager like systemd.  [`etcha run`]({{< ref "/docs/references/cli#run" >}}) will start Etcha in listening mode.

An example systemd unit might look like this:
```
[Unit]
Description=Infinite scale configuration management for distributed platforms
Documentation=https://etcha.dev
After=network.target

[Service]
ExecStart=/usr/local/bin/etcha -c /etc/etcha.jsonnet run
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

By default, Etcha will generate a self-signed certificate and listen on tcp/4000.  You can specify a certificate and different listen ports in the [`run`]({{< ref "/docs/references/config#run" >}}) configuration.

When Etcha is ran in listening mode, and is configured to have web services available like push or monitoring, [rate limiting]({{< ref "/docs/references/config#rateLimitRate" >}}) is enforced on all endpoints.

### Verifying JWTs and Patterns

Similar to [Signing Patterns]({{< ref "/docs/guides/building-patterns#signing-patterns" >}}) during [`etcha build`]({{< ref "/docs/references/cli#build" >}}), we need to verify the signed JWTs when we push or pull them.  We do this by providing configuration values on the remote Etcha instance for [`verifyKeys`]({{< ref "/docs/references/config#verifykeys" >}}) or [`verifyCommands`]({{< ref "/docs/references/config#verifycommands" >}}).

#### `verifyKeys`

`verifyKeys` will use a list of static, public keys to verify the Pattern JWT.  These values can be hardcoded in the remote Etcha instance, or you can retrieve them from environment variables, a remote URL, a DNS record, etc.  These public keys **must match** the private key used to sign the JWT.  Unlike the private, `signingKey`, these keys do not need to be kept secret.

These keys can be generated using [`etcha generate-keys sign-verify`]({{< ref "/docs/references/cli#generate-keys" >}}).  You can also bring your own keys, see [Cryptography]({{< ref "/docs/references/cryptography" >}}) for details on formatting.

Here is an example Etcha configuration showing static `verifyKeys`:

```json
{
  "run": {
    "verifyKeys": [
      "ed25519public:MCowBQYDK2VwAyEA1vXebFnhBwKJG/n2njHqx6waTsw5JpMZwSh0rRyC61Y=:pedKuY9EjQ",
      "ed25519public:MCowBQYDK2VwAyEAJl0hUXi9y5qF78QJPJ8W33DqB7WxAyXIb6/dpw0LtYM=:nh93JrtnbI",
    ]
  }
}
```

We can also leverage some of the [dynamic Jsonnet functions]({{< ref "/docs/references/jsonnet#native-functions" >}}) to pull the key from somewhere else, like say a DNS record:

```
local getRecord(type, name, fallback=null) = std.native('getRecord')(type, name, fallback);

{
  run: {
    verifyKeys: [
      getRecord('TXT', 'etcha_public.example.com'),
    ],
  },
}
```

This example looks a little intense, lets walk through it:

- We declared a function at the top to leverage [Jsonnet native functions]({{< ref "/docs/references/jsonnet#native-functions" >}}), `getRecord` which will retrieve a DNS record.
- We defined our run object and verifyKeys using those function, working inside out:
  - We lookup the TXT record for `etcha_public.example.com`.


#### `verifyCommands`

Some organizations may need to perform verification in a more secure, restrictive manner, like delegating signing to a HSM or HashiCorp Vault.  We can use `verifyCommands` for this.  `verifyCommands` are a list of Commands that will use environment variables to verify a Token.

Etcha will set the [Environment Variable]({{< ref "/docs/references/commands#environment-variables" >}}) [`ETCHA_JWT`]({{< ref "/docs/references/commands#etcha_jwt" >}}) containing the base64 raw URL encoded string that needs to be verified.

Our verify Commands need to verify the [JWT]({{< ref "/docs/references/jwt" >}}) contained with `ETCHA_JWT`, and then print the JWT during a Command `change` that triggers the event [`jwt`]({{< ref "/docs/references/events#jwt" >}}), like this:

```json
[
  {
    "always": true,
    "change": "<commands to verify JWT>",
    "id": "verify JWT",
    "onChange": [
      "etcha:jwt"
    ]
  }
]
```

### Sources

[Sources]({{< ref "/docs/references/config#sources" >}}) are how we configure an Etcha instance to push or pull Patterns.  Sources will typically have a 1:1 mapping with Patterns.

Sources can have their own [`exec configuration`]({{< ref "/docs/references/config#exec" >}}), `verifyKeys`(#verifykeys), and other options.  Sources can also be configured to always run in [**Check Mode**]({{< ref "/docs/references/commands#check-mode" >}}) via [`checkOnly`]({{< ref "/docs/references/config#checkonly" >}}).

During push/pull mode, Etcha will perform a diff against the current Pattern and the new Pattern.  **Only Commands that have changed will be `checked`, and any Commands not in the new Pattern will be `removed`**.  This behavior can be overriden using [`runAll`]({{< ref "/docs/references/config#runall" >}}).

Etcha can be configured for multiple sources:

```json
{
  "sources": {
    "myapp1": {
      "allowPush": true,
      "pullTargets": [
        "https://s3.example.com/myapp1.jwt",
        "/mnt/nfs/myapp1.jwt",
      ]
    },
    "myapp2": {
      "allowPush": true,
      "pullTargets": [
        "https://s3.example.com/myapp2.jwt",
        "/mnt/nfs/myapp2.jwt",
      ]
    }
  }
}
```

### Remote Run via Push

Etcha can push Patterns to remote Etcha instances.  This is similar to tools like Ansible, however it uses HTTPS instead of SSH.

For this to work, the remote instance needs to allow pushes via [`allowPush`]({{< ref "/docs/references/config#allowpush" >}}).  From there, we can push Patterns using [`etcha push pattern/myapp.jsonnet https://server:4000/etcha/v1/push/myapp1`].  The path `/etcha/v1/push/myapp1` is the push path, where `myapp1` is the name of a valid source on the remote Etcha instance.  Etcha will create and sign the JWT from the Pattern file `pattern/myapp.pattern`.  **When using Push mode, Pattern `build` commands are not executed**.  Switch to [Pull Mode](#remote-run-via-pull) to use `build` commands.

On the remote Etcha instance, all JWTs pushed are validated, rendered into a Pattern, and then the Pattern's `run` is executed according to the [Sources](#sources) configuration.  The client that sent the push will receive a 404 if the Source doesn't exist or didn't validate the JWT.  A successful validation by a source will return a Result object, containing information on what was changed or removed, if Etcha will exit after sending the response via [`exitEtcha`]({{< ref "/docs/references/config#exitetcha" >}}), as well as any errors:

```json
{
  "changed": [
    "run myapp1"
  ],
  "err": "myapp2 couldn't be removed",
  "exit": true,
  "removed": [
    "run myapp2"
  ]
}
```

**All clients are subject to rate limiting by the remote Etcha instance**, configured via [`rateLimiterRate`]({{< ref "/docs/references/config#ratelimiterrate" >}}).

### Remote Run via Pull

Etcha can pull Patterns from any local path or web service.  This scales far better than [Push](#remote-run-via-push), and should be used for anything more than a handful of instances.

To get started with pull mode, [Build your pattern]({{< ref "/docs/guides/building-patterns" >}}).  Copy the JWT to the remote instance, or put it somewhere the remote instance can access like a S3 bucket, web server, or any other HTTP provider.

On the remote instance, configure a source to pull Patterns via [`pullPaths`]({{< ref "/docs/references/config#pullpaths" >}}):

```json
{
  "myapp1": {
    "pullPaths": [
      "https://s3.example.com/myapp1.jwt"
    ]
  }
}
```

Then decide how you want to pull the JWT:

- Use [Listen Mode](#listen-mode) to have Etcha pull the JWT at startup, and then set a value for [`runFrequencySec`]({{< ref "/docs/references/config#runfrequencysec" >}}) to periodically pull and execute the JWT.
- Use [`etcha run`]({{< ref "/docs/references/cli#run" >}}) to have Etcha pull the JWT.  This allows you to trigger Etcha pulling based on a separate tool/service, like a cron job or systemd timer.

On the remote Etcha instance, all JWTs pulled are validated, rendered into a Pattern, and then the Pattern's `run` is executed and then ran according to the [Sources](#sources) configuration.  You can monitor the outcome of the pull/apply via logs, or in listen mode via metrics, see [Monitoring]({{< ref "/docs/guides/monitoring" >}}) for more information.

### Remote Run via Events

[Sources](#sources) can be configured to send and receive [Events]({{< ref "/docs/references/events" >}}).  This allows Etcha to trigger other Sources/Patterns dynamically.  When a Source is triggered via an Event listed in the [`eventsReceive`]({{< ref "/docs/references/config#eventsreceive" >}}) configuration, the associated Pattern's `run` Commands are executed with various [Environment Variables]({{< ref "/docs/references/commands#environment-variables" >}}) set related to the Event.

Events are handled in the order they were sent from the Commands, and Sources are dispatched events in alphabetically order, i.e. Source `pattern1` will execute before `zpattern1`.  Any errors when executing a Pattern to handle an event will not stop Event execution.  Additionally:

- Sources **cannot trigger themselves**
- Events created by Commands when handling an Event **will not trigger Sources**

### Remote Run via Webhooks

[Sources](#sources) can be configured to receive **Webhooks**.  Each Source can define [`webhookPaths`]({{< ref "/docs/references/config#webhookpaths" >}}) that Etcha will listen for requests on.  The requests can use any HTTP method, like `GET` or `POST`.  Etcha will accept the request, convert it into base64, add it and other request values into various [Environment Variables]({{< ref "/docs/references/commands#environment-variables" >}}), and execute the associated Pattern's `run` Commands.

These `Commands` must emit the [`webhookBody`]({{< ref "/docs/references/events#webhookbody" >}}) Event, which will be sent back to the Webhook client.  If no `webhookBody` is received from any source, or a Source is not found associated with the Webhook path, Etcha will send a 404.

Sources will handle Webhooks alphabetically:

```json
{
  "pattern2": {
    "webhookPaths": [
      "/mypath"
    ]
  },
  "pattern1": {
    "webhookPaths": [
      "/mypath"
    ],
  }
}
```

In this example, `pattern1` will receive the Webhook for `/mypath` first.  If it does not receive a `webhookBody`, it will send the Webhook to `pattern2`.  Events created by Commands ran when handling a Webhook **will not trigger Sources**.
