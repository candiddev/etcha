---
categories:
- reference
description: Reference documentation for Etcha Commands
title: Commands
---

## Command

A [Command](#command) is the smallest unit of work within Etcha.  [Patterns]({{< ref "/docs/references/patterns" >}}) contain build and run properties which are lists of [Commands](#commands), as well as Signing and Verify commands for integrating JWT signing/verification with other systems.

### Commands

Commands are a list of Command objects in an array.  Commands can be specified as objects, or as just a string that will be interpreted as a Command that will always be changed:

```json
[
  "apt-get install postgresql",
  {
    "always": true,
    "id": "start postgresql",
    "change": "systemctl start postgresql"
  }
]
```

In this example, `apt-get install postgresql` would become this Command:

```json
[
  {
    "always": true,
    "change": "apt-get install postgresql",
    "id": "apt-get install postgresql"
  },
  {
    "always": true,
    "id": "start postgresql",
    "change": "systemctl start postgresql"
  }
]
```

Command IDs must be unique, however Etcha will deduplicate Commands that aren't unique.  If the Commands do not match, Etcha will throw an error during rendering.

### Ordering

Commands within a list are evaluated mostly in the order listed.  Etcha generates a Directed Acyclic Graph (DAG), which considers the order of Commands but also whether a Command depends on another via [`after`](#after) or [`before`](#before), or if a Command is [`async`](#async).  The DAG can be previewed using {{% cli graph %}}, which can be helpful to understand ordering and troubleshoot dependency ccyles.

### Flattening

Lists of Commands can contain nested arrays within them:

```json
[
  {
    "id": "my command"
  },
  [
    [
      [
        {
          "id": "my nested command"
        }
      ]
    ]
  ]
]
```

Etcha will flatten these Commands:

```json
[
  {
    "id": "my command"
  },
  {
    "id": "my nested command"
  }
]
```


## Execution

A Command is executed using the values within {{% config exec %}}.  Exec overrides mean the Command may need to support, handle, or not run within other Exec configurations.

## Environment Variables

A Command is passed environment variables within the `exec` configuration (or inherited values from the parent process).  Etcha will attempt to resolve environment variables before running commands, e.g. if `${MYVAR}` is present in a command, Etcha will resolve this if a match environment variable exists before passing it to the underlying command.

Each command can also add environment variables to subsequent check/change/command executions:

### `ETCHA_EVENT_ID`

This variable will be set for Commands in {{% config sources_eventReceive %}} Pattern run lists.  It contains the [`id`](#id) of the Command that triggered the event.

### `ETCHA_EVENT_NAME`

This variable will be set for Commands in {{% config sources_eventReceive %}} Pattern run lists.  It contains the event name that triggered the event.

### `ETCHA_EVENT_PUBLIC_KEY`

This variable will be set for Commands in {{% config sources_eventReceive %}} Pattern run lists.  It contains the public key ID used to sign the pattern or shell that triggered the event.

### `ETCHA_EVENT_SRC`

This variable will be set for Commands in {{% config sources_eventReceive %}} Pattern run lists.  It contains the remote address (IPv4 **or** IPv6, where applicable) of the source push or shell event.

### `ETCHA_EVENT_OUTPUT`

This variable will be set for Commands in {{% config sources_eventReceive %}} Pattern run lists.  It contains the stdout/stderr of the `change` that triggered the event.

### `ETCHA_JWT`

This variable will be set for {{% config run_verifyCommands %}}.  It contains the JWT that needs to be verified.

### `ETCHA_PAYLOAD`

This variable will be set for {{% config build_signingCommands %}}.  It contains the base64 JWT payload that needs to be signed.

### `ETCHA_SOURCE_NAME`

This variable will be set for Commands in {{% config sources_eventReceive %}} or {{% config sources_webhookPaths %}} Pattern run lists.  It contains the name of the {{% config sources %}} receiving the Event or Webhook Pattern.


### `ETCHA_SOURCE_TRIGGER`

This variable will be set for Commands in {{% config sources_eventReceive %}} or {{% config sources_webhookPaths %}} Pattern run lists.  It contains the type for the trigger, `event` or `webhook`.

### `ETCHA_WEBHOOK_BODY`

This variable will be set for Commands in {{% config sources_webhookPaths %}} Pattern run lists.  It contains the base64 encoded body of the webhook request.

### `ETCHA_WEBHOOK_HEADERS`

This variable will be set for Commands in {{% config sources_webhookPaths %}} Pattern run lists.  It contains a list of all webhook headers, separated with a newline (`\n`).

### `ETCHA_WEBHOOK_METHOD`

This variable will be set for Commands in {{% config sources_webhookPaths %}} Pattern run lists.  It contains the name of the webhook method (`DELETE|GET|POST|PUT`).

### `ETCHA_WEBHOOK_PATH`

This variable will be set for Commands in {{% config sources_webhookPaths %}} Pattern run lists.  It contains the request path for the webhook.

### `ETCHA_WEBHOOK_QUERY`

This variable will be set for Commands in {{% config sources_webhookPaths %}} Pattern run lists.  It contains the request query params separated with a `&`.

### `envPrefix`

This variable will be set to the stdout and stderr of the check execution of a Command with [`envPrefix`](#envPrefix).  This variable will not be set if the Command wasn't checked.  If no [`envPrefix`](#envPrefix) is defined, this variable will not be set.

### `<envPrefix>_CHECK`

This variable will be set to 0 if a Command is checked without any error or skipped checking due to no check value, `always` set to false, or not changed by anything.  It will be 1 if it had errors while checking.  If no [`envPrefix`](#envPrefix) is defined, the variable will be `_CHECK` and will show the previously run command.

### `<envPrefix>_CHANGE`

This variable will be set to 0 if a Command is changed without any error, or 1 if it had errors.  This variable will not be set if the Command didn't have any change executed.  If no [`envPrefix`](#envPrefix) is defined, the variable will be `_CHANGE` and will show the previously run command.

### `<envPrefix>_CHANGE_OUT`

This variable will be set to the stdout and stderr of the change execution of a Command.  This variable will not be set if the Command didn't have any change executed.  If no [`envPrefix`](#envPrefix) is defined, this variable will not be set.

### `<envPrefix>_REMOVE`

This variable will be set to 0 if a Command is removed without any error, and 1 if it had errors.  If no [`envPrefix`](#envPrefix) is defined, the variable will be `_REMOVE` and will show the previously run command.

### `<envPrefix>_REMOVE_OUT`

This variable will be set to the stdout and stderr of the remove execution for a Command.  If no [`envPrefix`](#envPrefix) is defined, this variable will not be set.

## Operating Modes

A Command is ran within three different operating modes:

### Change (default) {#change-mode}

The default, will always run [`check`](#check) if specified, and run [`change`](#change) if [`always`](#always) is true, `check` is non-zero, or the [`id`](#id) is [`changed by`](#on) another command.

### Check {#check-mode}

Will always run [`check`](#check) if specified only.  {{% config sources %}} can be forced to run in check mode, and patterns can be ran in check mode using {{% config checkOnly %}}.

### Remove {#remove-mode}

Will always run [`check`](#check) if specified, and run [`remove`](#remove) if [`always`](#always) is true, `check` is zero, or the [`id`](#id) is [`removed by`](#on) another command.

For push and pull, Etcha by default diff Patterns and run checks and removes for Commands that are no longer present in the new Pattern, as well as for any Commands that have a modified `change` values (unless `changeIgnore` is specified).

## Properties

### `after`

String or list of strings, IDs or Provides of Commands that must occur after this Command.  When Commands are defined in a top level or [`commands`](#commands) block without [`parallel`](#parallel) set to `true`, command IDs will be added to `after` based on the order they are defined.

### `always`

Boolean, when true, [`change`](#change) will always be ran during [Change Mode](#change-mode)

### `async`

Boolean, when true, follow on Commands will not wait for this Command, unless explicitly referenced in `after`.

### `before`

String or list of strings, IDs or Provides of Commands that must before this Command.  When Commands are defined in a top level or [`commands`](#commands) block without [`parallel`](#parallel) set `true`, command IDs will be added to `before` based on the order they are defined.

### `change`

String, the commands or executable to run during [Change Mode](#change-mode).  Can be multiple lines.  Will be appended to `exec.command`.  Should return 0 if successful, otherwise it will produce an error.

### `changeIgnore`

Boolean, will ignore changes to the `change` Command.  By default, `change` differences will trigger a `remove` and `change` cycle for the Command.

### `check`

String, the commands or executable to run during [Change Mode](#change-mode) or [Check Mode](#check-mode).  Can be multiple lines.  Will be appended to `exec.command`.  If this returns 0, [`remove`](#remove) will be ran in [Remove Mode](#remove-mode).  If this does not return 0, [`change`](#change) will be ran in [Change Mode](#change-mode).  If omitted, [`change`](#change) or [`remove`](#remove) will never run unless [`always`](#always) is `true` or [`id`](#id) is changed by another Command via [`onChange`](#onChange) or removed by another Command via [`onRemove`](#onRemove)

### `commands`

A list of sub Commands.  Other properties for this Command will be ignored except `id`.  These Commands will be ran in a group and not affect other groups.

### `env`

Map of strings keys and string values, environment variables that will be set for this Command specifically.

### `envPrefix`

String, an environment variable name prefix to add to all [Environment Variables](#environment-variables) created by this command.  Must be a valid environment variable (does not start with a number, must only contain word characters or _).

### `exec`

See {{% config exec %}}.  Specifies a custom exec configuration for this command.  Parent exec configurations must allow overrides for this to work.

### `id` (required) {#id}

An ID for the Command.  Must be specified.  Can overlap with other Commands.

### `locks`

String or list of strings, locks that must succeed before this Command is run.  Locks are global within an Etcha instance, so Commands running in different sources will share the same locks.  Locks will be sorted alphabetically to avoid deadlocks.

### `onChange`, `onRemove` {#on}

A list of other Command [`id`s](#id) or [`provides`](#provides) to run or [Events]({{< ref "/docs/references/events" >}}) to trigger, if this Command changes or removes.  Event names must be prefixed with `etcha:`.  Any matching Commands will be automatically added to this Command's `after` list.

Cannot specify the current command ID (can't target self).

### `parallel`

Boolean, when `true` and used with [`commands`](#commands), the Commands will be executed in parallel.

### `provides`

String or list of strings, additional matches for this Command to allow other Commands to target it using [`after`](#after), [`before`](#before), [`onChange` and `onRemove`](#on).

### `remove`

String, the commands or executable to run during [Remove Mode](#remove-mode).  Can be multiple lines.  Will be appended to `exec.command`.  Should return 0 if successful, otherwise it will produce an error.

### `removeAfter`

Boolean, will change the ordering of `remove` to be executed after the Command's `change` is ran.  By default, `remove` is executed before `change`.

### `stdin`

String, sets the stdin for the Command for `change`, `check`, and `remove`.
