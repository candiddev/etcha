---
categories:
- reference
description: Reference documentation for Etcha Commands
title: Commands
---

## Command

A [Command](#command) is the smallest unit of work within Etcha.  [Patterns]({{< ref "/docs/references/patterns" >}}) contain build and run properties which are lists of [Commands](#commands), as well as Signing and Verify commands for integrating JWT signing/verification with other systems.

Within a list of Commands, **Commands are executed in the order they are listed**.  Additionally, a list of Commands can contain nested lists of Commands.  Etcha will flatten the lists into one ordered list automatically.

## Execution

A Command is executed using the values within [`exec`]({{< ref "/docs/references/config#exec" >}}).  Exec overrides mean the Command may need to support, handle, or not run within other Exec configurations.

## Environment Variables

A Command is passed environment variables within the `exec` configuration (or inherited values from the parent process).  Etcha will attempt to resolve environment variables before running commands, e.g. if `${MYVAR}` is present in a command, Etcha will resolve this if a match environment variable exists before passing it to the underlying command.

Each command can also add environment variables to subsequent check/change/command executions:

### `ETCHA_EVENT_ID`

This variable will be set for Commands in [`eventReceive`]({{< ref "/docs/references/config#eventreceive" >}}) Pattern run lists.  It contains the [`id`](#id) of the Command that triggered the event.

### `ETCHA_EVENT_NAME`

This variable will be set for Commands in [`eventReceive`]({{< ref "/docs/references/config#eventreceive" >}}) Pattern run lists.  It contains the event name that triggered the event.

### `ETCHA_EVENT_OUTPUT`

This variable will be set for Commands in [`eventReceive`]({{< ref "/docs/references/config#eventreceive" >}}) Pattern run lists.  It contains the stdout/stderr of the `change` that triggered the event.

### `ETCHA_JWT`

This variable will be set for [`verifyCommands`]({{< ref "/docs/references/config#verifycommands" >}}).  It contains the JWT that needs to be verified.

### `ETCHA_PAYLOAD`

This variable will be set for [`signCommands`]({{< ref "/docs/references/config#signcommands" >}}).  It contains the base64 JWT payload that needs to be signed.

### `ETCHA_SOURCE_NAME`

This variable will be set for Commands in [`eventReceive`]({{< ref "/docs/references/config#eventreceive" >}}) or [`webhookPaths`]({{< ref "/docs/references/config#webhookPaths" >}}) Pattern run lists.  It contains the name of the [`source`]({{< ref "/docs/references/config#sources" >}}) receiving the Event or Webhook Pattern.

### `ETCHA_SOURCE_TRIGGER`

This variable will be set for Commands in [`eventReceive`]({{< ref "/docs/references/config#eventreceive" >}}) or [`webhookPaths`]({{< ref "/docs/references/config#webhookPaths" >}}) Pattern run lists.  It contains the type for the trigger, `event` or `webhook`.

### `ETCHA_WEBHOOK_BODY`

This variable will be set for Commands in [`webhookPaths`]({{< ref "/docs/references/config#webhookpaths" >}}) Pattern run lists.  It contains the base64 encoded body of the webhook request.

### `ETCHA_WEBHOOK_HEADERS`

This variable will be set for Commands in [`webhookPaths`]({{< ref "/docs/references/config#webhookpaths" >}}) Pattern run lists.  It contains a list of all webhook headers, separated with a newline (`\n`).

### `ETCHA_WEBHOOK_METHOD`

This variable will be set for Commands in [`webhookPaths`]({{< ref "/docs/references/config#webhookpaths" >}}) Pattern run lists.  It contains the name of the webhook method (`DELETE|GET|POST|PUT`).

### `ETCHA_WEBHOOK_PATH`

This variable will be set for Commands in [`webhookPaths`]({{< ref "/docs/references/config#webhookpaths" >}}) Pattern run lists.  It contains the request path for the webhook.

### `ETCHA_WEBHOOK_QUERY`

This variable will be set for Commands in [`webhookPaths`]({{< ref "/docs/references/config#webhookpaths" >}}) Pattern run lists.  It contains the request query params separated with a `&`.

### `envPrefix`

This variable will be set to the stdout and stderr of the check execution of a Command with [`envPrefix`](#envPrefix).  This variable will not be set if the Command wasn't checked.  If no envPrefix is defined, this variable will not be set.

### `<envPrefix>_CHECK`

This variable will be set to 0 if a Command is checked without any error or skipped checking due to no check value, `always` set to false, or not changed by anything.  It will be 1 if it had errors while checking.  If no [`envPrefix`](#envPrefix) is defined, the variable will be `_CHECK` and will show the previously run command.

### `<envPrefix>_CHECK_OUT`

This variable will be set to the stdout and stderr of the check execution of a Command.  This variable will not be set if the Command wasn't checked.  If no [`envPrefix`](#envPrefix) is defined, the variable will be `_CHECK_OUT` and will show the previously run command.

### `<envPrefix>_CHANGE`

This variable will be set to 0 if a Command is changed without any error, or 1 if it had errors.  This variable will not be set if the Command didn't have any change executed.  If no [`envPrefix`](#envPrefix) is defined, the variable will be `_CHANGE` and will show the previously run command.

### `<envPrefix>_CHANGE_OUT`

This variable will be set to the stdout and stderr of the change execution of a Command.  This variable will not be set if the Command didn't have any change executed.  If no [`envPrefix`](#envPrefix) is defined, the variable will be `_CHANGE_OUT` and will show the previously run command.

### `<envPrefix>_REMOVE`

This variable will be set to 0 if a Command is removed without any error, and 1 if it had errors.  If no [`envPrefix`](#envPrefix) is defined, the variable will be `_REMOVE` and will show the previously run command.

### `<envPrefix>_REMOVE_OUT`

This variable will be set to the stdout and stderr of the remove execution for a Command.  If no [`envPrefix`](#envPrefix) is defined, the variable will be `_REMOVE_OUT` and will show the previously run command.

## Operating Modes

A Command is ran within three different operating modes:

### Change (default) {#change-mode}

The default, will always run [`check`](#check) if specified, and run [`change`](#change) if [`always`](#always) is true, `check` is non-zero, or the [`id`](#id) is [`changed by`](#on) another command.

For push and pull, Etcha by default diff Patterns and run checks and changes for Commands that have a different [`change`](#change) or [`check`](#check) value, as well as the `change` value of any Command with [`always`](#always) set to `true`.

### Check {#check-mode}

Will always run [`check`](#check) if specified only.  [Sources]({{< ref "/docs/references/config#sources" >}}) can be forced to run in check mode, and patterns can be ran in check mode using [`checkOnly`]({{< ref "/docs/references/config#checkonly" >}})

### Remove {#remove-mode}

Will always run [`check`](#check) if specified, and run [`remove`](#remove) if [`always`](#always) is true, `check` is zero, or the [`id`](#id) is [`removed by`](#on) another command.

For push and pull, Etcha by default diff Patterns and run checks and removes for Commands that are no longer present in the new Pattern.

## Properties

### `always`

Boolean, when true, [`change`](#change) will always be ran during [Change Mode](#change-mode)

### `change`

String, the commands or executable to run during [Change Mode](#change-mode).  Can be multiple lines.  Will be appended to `exec.command`.  Should return 0 if successful, otherwise it will produce an error.

### `check`

String, the commands or executable to run during [Change Mode](#change-mode) or [Check Mode](#check-mode).  Can be multiple lines.  Will be appended to `exec.command`.  If this returns 0, [`remove`](#remove) will be ran in [Remove Mode](#remove-mode).  If this does not return 0, [`change`](#change) will be ran in [Change Mode](#change-mode).  If omitted, [`change`](#change) or [`remove`](#remove) will never run unless [`always`](#always) is `true` or [`id`](#id) is changed by another Command via [`onChange`](#onChange) or removed by another Command via [`onRemove`](#onRemove)

### `commands`

A list of sub Commands.  Other properties for this Command will be ignored except `id`.  These Commands will be ran in a group and not affect other groups.

### `envPrefix`

String, an environment variable name prefix to add to all [Environment Variables](#environment-variables) created by this command.  Must be a valid environment variable (does not start with a number, must only contain word characters or _).

### `exec`

See [`exec`]({{< ref "/docs/references/config#exec" >}}).  Specifies a custom exec configuration for this command.  Parent exec configurations must allow overrides for this to work.

### `id` (required) {#id}

An ID for the Command.  Must be specified.  Can overlap with other Commands.

### `onChange`, `onFail`, `onRemove` {#on}

A list of:
- Other Command [`id`s](#id) to run
- Regular expressions to match Command [`id`s](#id) to run
- [Events]({{< ref "/docs/references/events" >}}) to trigger, if this Command changes, removes or fails.  Event names must be prefixed with `etcha:`.

Cannot specify the current command ID (can't target self).  For onChange, targets must exist and occur after the current Command in the Command list (onRemove is the opposite, must occur before), or there will be an error during compilation.

These IDs can only target IDs within the current Command list:

```json
[
  {
    id: "a",
    commands: [
      {
        id: "b"
      },
      {
        id: "c"
      },
    ],
  },
  {
    id: "d"
  }
]
```

In this example, `b` can target `c` but cannot target `d`.

### `remove`

String, the commands or executable to run during [Remove Mode](#remove-mode).  Can be multiple lines.  Will be appended to `exec.command`.  Should return 0 if successful, otherwise it will produce an error.

### `stdin`

String, sets the stdin for the Command for `change`, `check`, and `remove`.
