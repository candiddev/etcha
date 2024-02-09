---
categories:
- reference
description: Reference documentation for Etcha Patterns
title: Patterns
---

Patterns are what Etcha uses to build and run [Commands]({{< ref "/docs/references/commands" >}}).

## Build vs Run

Patterns have two separate list of [Commands]({{< ref "/docs/references/commands" >}}): [`build`](#build) and [`run`](#run).  `build` Commands are ran on the local instance when the Pattern is built using [`etcha build`]({{< ref "/docs/references/cli#build" >}}).  `run` Commands are ran on an instance when a Pattern is push, pulled, manually applied via [CLI]({{< ref "/docs/references/cli" >}}), or triggered via [`eventsReceive`]({{< ref "/docs/references/config#eventsreceive" >}}) and [`webhookPaths`]({{< ref "/docs/references/config#webhookpaths" >}})) in a [Source]({{< ref "/docs/references/config#sources" >}}).

## Jsonnet

Patterns are typically written in [Jsonnet]({{< ref "/docs/references/jsonnet" >}}).  A Pattern will have one "entrypoint" file with the extension `.jsonnet`, such as `main.jsonnet`.  This file should return a JSON object containing the [Pattern Properties](#properties).  It can import other Jsonnet files (with the file extension `.libsonnet`) or any other file to create the Pattern.

## Properties

### `audience`

String, a list of audience values to set in the [JWT `aud` property]({{< ref "/docs/references/jwt#aud" >}}).

### `build`

A list of [Commands]({{< ref "/docs/references/commands" >}}) to execute when building a [JWT]({{< ref "/docs/references/jwt" >}}).

### `buildExec`

See [`exec`]({{< ref "/docs/references/config#exec" >}}).  Specifies a custom exec configuration for the Pattern `build` Commands.  Parent exec configurations must allow overrides for this to work.

### `expiresInSec`

Integer, specifies the number of seconds from the current time until the [JWT]({{< ref "/docs/references/jwt" >}}) expires.  Defaults to 0/JWTs do not expire.  If a JWT expires, it will not be trusted/ran.

### `id`

String, the ID to set in the [JWT `jti` property]({{< ref "/docs/references/jwt#jti" >}}).

### `issuer`

String, the Issuer to set in the [JWT `iss` property]({{< ref "/docs/references/jwt#iss" >}}).

### `run`

A list of [Commands]({{< ref "/docs/references/commands" >}}) to execute when a Pattern is push, pulled, manually applied via [CLI]({{< ref "/docs/references/cli" >}}), or triggered via [`eventsReceive`]({{< ref "/docs/references/config#eventsreceive" >}}) and [`webhookPaths`]({{< ref "/docs/references/config#webhookpaths" >}}) in a [Source]({{< ref "/docs/references/config#sources" >}}).

### `runExec`

See [`exec`]({{< ref "/docs/references/config#exec" >}}).  Specifies a custom exec configuration for the Pattern `run` Commands.  Parent exec configurations must allow overrides for this to work.

### `runVars`

A map of values that will be combined with [Vars]({{< ref "/docs/references/config#vars" >}}) when the Pattern is rendered.  These are exposed using the [Jsonnet native function, `getConfig`]({{< ref "/docs/references/jsonnet#getConfig" >}}), for rendering these values during a run:

```
{
  run: [
    id: 'run a thing',
    change: std.native('getConfig')().vars.myVar,
  ],
}
```

### `subject`

String, the Subject to set in the [JWT `sub` property]({{< ref "/docs/references/jwt#sub" >}}).

## Rendering

Patterns are rendered from Jsonnet everytime they are ran.  That means all of the Jsonnet functions, lookups, and environment variables are all executed/evaluated on the current instance that is running the Pattern.

### Variables

The native Jsonnet function, [`getConfig() object`]({{< ref "/docs/references/jsonnet#getConfig" >}}), can be used to retrieve the combined [`exec`]({{< ref "/docs/references/config#exec" >}}) and [`vars`]({{< ref "/docs/references/config#vars" >}}) for the `source`.  Given a configuration like this:

```json
{
  "sources": {
    "source1": {
      "exec": {
        "command": "/bin/bash"
      },
      "vars": {
        "var1": false,
        "var2": "value"
      }
    }
  },
  "vars": {
    "var1": true,
    "var2": "original"
  }
}
```

Running `getConfig().exec` within a Pattern for the source `source1` will render a Jsonnet object like this:

```json
{
  "command": "/bin/bash"
}
```

Running `getConfig().vars` within a Pattern for the source `source1` will render a Jsonnet object like this:

```json
{
  "var1": false,
  "var2": "value",
  "var3": "original"
}
```
