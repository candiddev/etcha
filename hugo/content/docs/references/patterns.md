---
categories:
- reference
description: Reference documentation for Etcha Patterns
title: Patterns
---

Patterns are what Etcha uses to build and run [Commands](../commands).

## Build vs Run

Patterns have two separate list of [Commands](../commands): [`build`](#build) and [`run`](#run).  `build` Commands are ran on the local instance when the Pattern is built using [`etcha build`](../cli#build).  `run` Commands are ran on an instance when a Pattern is push, pulled, manually applied via [CLI](../cli), or triggered via [`eventsReceive`](../config#eventsReceive) and [`webhookPaths`](../config#webhookPaths) in a [Source](../config#sources).

## Jsonnet

Patterns are typically written in [Jsonnet](../jsonnet).  A Pattern will have one "entrypoint" file with the extension `.jsonnet`, such as `main.jsonnet`.  This file should return a JSON object containing the [Pattern Properties](#properties).  It can import other Jsonnet files (with the file extension `.libsonnet`) or any other file to create the Pattern.

## Properties

### `audience`

String, a list of audience values to set in the [JWT `aud` property](../jwt#aud).

### `build`

A list of [Commands](../commands) to execute when building a [JWT](../jwt).

### `buildExec`

See [`exec`](../config#exec).  Specifies a custom exec configuration for the Pattern `build` Commands.  Parent exec configurations must allow overrides for this to work.

### `expiresInSec`

Integer, specifies the number of seconds from the current time until the [JWT](../jwt) expires.  Defaults to 0/JWTs do not expire.  If a JWT expires, it will not be trusted/ran.

### `id`

String, the ID to set in the [JWT `jti` property](../jwt#jti).

### `issuer`

String, the Issuer to set in the [JWT `iss` property](../jwt#iss).

### `run`

A list of [Commands](../commands) to execute when a Pattern is push, pulled, manually applied via [CLI](../cli), or triggered via [`eventsReceive`](../config#eventsReceive) and [`webhookPaths`](../config#webhookPaths) in a [Source](../config#sources).

### `runEnv`

A map of [Environment Variables](../commands#environment-variables) that will be added to [Commands](../commands) when the Patter is run.

The keys must be valid Environment Variable names.  `etcha_run_` will be prepended to the key names.

Additional environment variables will be added to this property  from the JWT during a run via build commands that fire the [`run_env_`](../events#run_env_) event.

### `runExec`

See [`exec`](../config#exec).  Specifies a custom exec configuration for the Pattern `run` Commands.  Parent exec configurations must allow overrides for this to work.

### `subject`

String, the Subject to set in the [JWT `sub` property](../jwt#sub).

## Rendering

Patterns are rendered from Jsonnet everytime they are ran.  That means all of the Jsonnet functions, lookups, and environment variables are all executed/evaluated on the current instance that is running the Pattern.
