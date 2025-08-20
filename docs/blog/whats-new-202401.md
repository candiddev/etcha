---
author: Candid Development
date: 2024-01-01
description: Release notes for Etcha v2024.01.
tags:
  - release
title: "What's New in Etcha: v2024.01"
type: blog
---

## Features

### Declarative Helpers

Etcha now ships with declarative helper arguments:

- `etcha copy` to copy files.
- `etcha dir` to manage directories.
- `etcha file` to manage files.
- `etcha get` to retrieve files locally or from HTTP endpoints.
- `etcha line` to manage lines in files.
- `etcha link` to manage symlinks.

[Libraries]({{< ref "/docs/references/libraries" >}}) have been updated to use these native functions instead of system helpers.

See [CLI]({{< ref "/docs/references/cli" >}}) for details.

### Variable Evaluation

Patterns will now evaluate shell variables (`${VARIABLE}`) before executing a Command if the variable is defined.  This applies to variables specified in `check`, `change`, and `remove` Commands, as well as the new `stdin` value.

## Enhancements

- **CLI:** Changed and combined CLI arguments to be more concise and reduce the number of similar arguments.  See [CLI]({{< ref "/docs/references/cli" >}}) for details.
- **Commands:** Stdin can now be set for Commands.  See [Commands]({{< ref "/docs/references/commands" >}}) for details.
- **Config:** Container networks can now be set for Exec configs.  See [Commands]({{< ref "/docs/references/commands" >}}) for details.
- **Jsonnet:** Added [`getArch`]({{< ref "/docs/references/jsonnet#getarch" >}}), a jsonnet library for retrieving the current arch.
- **Jsonnet:** Added [`getOS`]({{< ref "/docs/references/jsonnet#getarch" >}}), a jsonnet library for retrieving the current OS.
- **Libraries:** Added [`line`]({{< ref "/docs/references/libraries#line" >}}), a library for managing lines in files.
- **Libraries:** Added [`etchaInstall`]({{< ref "/docs/references/libraries#etchaInstall" >}}), a library for managing lines in files.

## Fixes

- **CLI:** Fixed {{% cli render %}} not parsing JWTs correctly.
- **Lint:** Fixed linting reporting duplicate results.
- **Patterns:** Fixed Pattern imports using the Jsonnet function `getEnv` not obeying the `exec.envInherit` or `exec.env` values within the main configuration or source configuration.
- **Sources:** Fixed Pattern loading and execution at startup not obeying naming order of sources.
