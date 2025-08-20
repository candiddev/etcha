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
    id: 'my app',
    commands: [
      {
        change: 'systemctl restart %s' % name,
        id: 'restart %s' % name,
      },
    ],
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
4. Use the configured {{% config exec buildExec %}} or {{% config exec runExec %}} if specified and allowed by {{% config exec_allowOverride %}}.
5. Execute the `run` [Commands]({{< ref "/docs/references/commands" >}}) in the order specified.

## Local Run via CLI

Etcha can run Patterns via one-off applies using the [CLI]({{< ref "/docs/references/cli" >}}).  These runs will always be local to the current Etcha instance.  They're performed using the following CLI commands:

- [`etcha local [pattern path]`]({{< ref "/docs/references/cli#local" >}}), executes the `run` in [**Change Mode**]({{< ref "/docs/references/commands#change-mode" >}}).
- [`etcha local -r [pattern path]`]({{< ref "/docs/references/cli#local" >}})), executes the `run` in [**Remove Mode**]({{< ref "/docs/references/commands#remove-mode" >}}).

{{% alert title="Candid Commentary" color="info" %}}
Local run works great for dotfiles.
{{% /alert %}}

### Render and Run

Local can also be passed raw Jsonnet that it will wrap in a `{run: [<your jsonnet>]}` string and render a Pattern on the fly.  You can use this to run [Libraries]({{< ref "/docs/references/libraries" >}}) or test other Jsonnet things:

```bash
$ etcha local "(import 'lib/etcha/etchaInstall.libsonnet')(dst='/tmp/etcha')"
INFO  Changed download Etcha to /tmp/etcha [2s]
INFO  Always changed etcha version [50ms]
```

## Remote Run

Etcha can push and pull [Pattern]({{< ref "/docs/references/patterns" >}}) [JWTs]({{< ref "/docs/references/jwt" >}}) to/from remote instances.  In order for this to work, we need to have a way to verify the JWT, and the remote Etcha instance needs to be configured to accept JWTs through a particular method via {{% config sources %}}.

{{% alert title="Monitoring" color="primary" %}}
Make sure to checkout [Monitoring]({{< ref "/docs/guides/monitoring" >}}) for an overview of how to monitor Etcha runs.  You can trigger Prometheus alerts when Commands run or fail, and track who is doing what.
{{% /alert %}}

### Listen Mode

Etcha needs to be running in listen mode to receive pushes, execute webhooks, and handle events.  You'll typically run Etcha in listen mode as a container or via a service manager like systemd.  {{% cli run %}} will start Etcha in listening mode.

By default, Etcha will generate a self-signed certificate and listen on tcp/4000.  You can specify a certificate and different listen ports in the {{% config run %}} configuration.

When Etcha is ran in listening mode, and is configured to have web services available like push or monitoring, {{% config httpServer_rateLimitRate %}} is enforced on all endpoints.

{{% alert title="Candid Commentary" color="info" %}}
Etcha can also be ran in a container or Kubernetes cluster, and can provide access to remote resources this way as well.
{{% /alert %}}

### Verifying JWTs and Patterns

Similar to [Signing Patterns]({{< ref "/docs/guides/building-patterns#signing-patterns" >}}) during {{% cli build %}}, we need to verify the signed JWTs when we push or pull them.  We do this by providing configuration values on the remote Etcha instance for {{% config run_verifyKeys  %}} or {{% config run_verifyCommands %}}.

#### `verifyKeys`

`verifyKeys` will use a list of static, public keys to verify the Pattern JWT.  These values can be hardcoded in the remote Etcha instance, or you can retrieve them from environment variables, a remote URL, a DNS record, etc.  These public keys **must match** the private key used to sign the JWT.  Unlike the private, `signingKey`, these keys do not need to be kept secret.

These keys can be generated using {{% cli gen-keys %}}.  You can also bring your own keys, see [Cryptography]({{< ref "/docs/references/cryptography" >}}) for details on formatting.

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

{{% alert title="License Required" color="warning" %}}
This requires an [Unlimited License]({{< ref "/pricing" >}})
{{% /alert %}}

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

{{% alert title="Candid Commentary" color="info" %}}
This is behind a license purely from a footgun standpoint.  If this is something you need to do, we want to make sure you're doing it correctly.
{{% /alert %}}

### Sources

{{% config sources %}} are how we configure an Etcha instance to push or pull Patterns.  Sources will typically have a 1:1 mapping with Patterns.

Sources can have their own {{% config exec "exec configuration" %}}, `verifyKeys`(#verifykeys), and other options.  Sources can also be configured to always run in [**Check Mode**]({{< ref "/docs/references/commands#check-mode" >}}) via {{% config sources_checkOnly %}}.

During push/pull mode, Etcha will perform a diff against the current Pattern and the new Pattern.  **Any Commands not in the new Pattern will be `removed`, as well as any Commands with a modified `change` value**.  This behavior can be overriden using {{% config sources_noRemove %}}.

After a source receives a push or a pull, it will cache the JWT in the {{% config run_stateDir %}}.  On startup, Etcha will restore these JWTs after validating them.  You can disable this behavior using {{% config sources_noRestore %}}.

Sources will periodically pull (if {{% config sources_pullPaths %}} are set) and run Commands (either from the Source's {{% config sources_commands %}}, `pullPaths`, or via pushes) if {{% config sources_runFrequencySec %}} is defined.

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

### Targets

{{% cli push %}} and {{% cli shell %}} can use adhoc targets provided on the command line using `-h hostname`, or you can defined a list of targets in the Etcha config and target them instead:

```json
{
  "targets": {
    "server1": {
      "sourcePatterns": [
        "core": "etcha/patterns/core.jsonnet",
        "debug": "",
        "nginx": "etcha/patterns/nginx.jsonnet",
      ]
    },
    "server2": {
      "sources": [
        "core": "etcha/patterns/core.jsonnet",
        "debug": "",
        "mysql": "etcha/patterns/mysql.jsonnet",
      ]
    }
  }
}
```

We can run `etcha push core patterns/core.jsonnet` and it will push to both `server1` and `server2`, or we can target the mysql source (`etcha push core mysql patterns/mysql.jsonnet`) and only target `server2`.  Both servers have an empty string Source, `debug`, that allows any Pattern or Command to be pushed.

### Remote Run via Push {#remote-push}

Etcha can push Patterns to remote Etcha instances.  This is similar to tools like Ansible, however it uses HTTPS instead of SSH.

For this to work, the remote instance needs to allow pushes via {{% config sources_allowPush %}}.  From there, we can push Patterns using {{% cli push %}}.

{{% alert color="primary" title="Mono Patterns" %}}
**You don't have to use multiple sources**.  With how flexible Patterns are, the `core` pattern could contain the `mysql` Pattern and only run it if a host's var ({{% config sources_pushTargets %}}) has `mysql: true`.
{{% /alert %}}

On the remote Etcha instance, all JWTs pushed are validated, rendered into a Pattern, and then the Pattern's `run` is executed according to the [Sources](#sources) configuration.  The client that sent the push will receive a 404 if the Source doesn't exist or didn't validate the JWT.  A successful validation by a source will return a Result object, containing information on what was changed or removed, as well as any errors:

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

**All clients are subject to rate limiting by the remote Etcha instance**, configured via {{% config rateLimiterRate %}}.

### Remote Run via Pull {#pull}

Etcha can pull Patterns from any local path or web service.  This scales far better than [Push](#remote-run-via-push), and should be used for anything more than a handful of instances.

To get started with pull mode, [Build your pattern]({{< ref "/docs/guides/building-patterns" >}}).  Copy the JWT to the remote instance, or put it somewhere the remote instance can access like a S3 bucket, web server, or any other HTTP provider.

On the remote instance, configure a source to pull Patterns via {{% config pullPaths %}}:

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

- Use [Listen Mode](#listen-mode) to have Etcha pull the JWT at startup, and then set a value for {{% config runFrequencySec %}} to periodically pull and execute the JWT.
- Use {{% cli run %}} to have Etcha pull the JWT.  This allows you to trigger Etcha pulling based on a separate tool/service, like a cron job or systemd timer.

On the remote Etcha instance, all JWTs pulled are validated, rendered into a Pattern, and then the Pattern's `run` is executed and then ran according to the [Sources](#sources) configuration.  You can monitor the outcome of the pull/apply via logs, or in listen mode via metrics, see [Monitoring]({{< ref "/docs/guides/monitoring" >}}) for more information.

### Remote Run via Events

[Sources](#sources) can be configured to send and receive [Events]({{< ref "/docs/references/events" >}}).  This allows Etcha to trigger other Sources/Patterns dynamically.  When a Source is triggered via an Event listed in the {{% config eventsReceive %}} configuration, the associated Pattern's `run` Commands are executed with various [Environment Variables]({{< ref "/docs/references/commands#environment-variables" >}}) set related to the Event.

Events are handled in the order they were sent from the Commands, and Sources are dispatched events in alphabetically order, i.e. Source `pattern1` will execute before `zpattern1`.  Any errors when executing a Pattern to handle an event will not stop Event execution.  Additionally:

- Sources **cannot trigger themselves**
- Events created by Commands when handling an Event **will not trigger Sources**

### Remote Run via Webhooks

[Sources](#sources) can be configured to receive **Webhooks**.  Each Source can define {{% config webhookPaths %}} that Etcha will listen for requests on.  The requests can use any HTTP method, like `GET` or `POST`.  Etcha will accept the request, convert it into base64, add it and other request values into various [Environment Variables]({{< ref "/docs/references/commands#environment-variables" >}}), and execute the associated Pattern's `run` Commands.

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
