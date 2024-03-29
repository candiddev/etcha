---
author: Mike
date: 2024-03-14
description: Release notes for Etcha v2024.03.
tags:
  - release
title: "What's New in Etcha: v2024.03"
type: blog
---

{{< etcha-release version="2024.03" >}}

## Features

### Push Targets

[Push Targets]({{< ref "/docs/references/config#pushtargets" >}}) can now be configured to create groups of push endpoints, similar to Ansible inventories or Puppet targets.

### Sub Commands

[Commands]({{< ref "/docs/references/commands" >}}) can now contain sub Commands.  These Commands are executed within their own scope for `onChange/Fail/Remove`.

## Enhancements

- Changed the ordering of change and remove.  By default, all removes will now happen **before changes**.  A Command can be configured for the old behavior (remove after change) using the property [`removeAfter`]({{< ref "/docs/references/commands#removeafter" >}}).
- Changed Commands to trigger a `remove` and `change` cycle if the `change` value is modified for an ID.  See [`changeIgnore`]({{< ref "/docs/references/commands#changeignore" >}}) for details on how to disable this.
- Changed [`exec.env`]({{< ref "/docs/references/config#env" >}}) to be a map of strings.
- Changed [`onChange`, `onFail`, and `onRemove`]({{< ref "/docs/references/patterns#on" >}}) to support RegExp values.
- Changed `etcha local`, `etcha push`, and `etcha test` to allow filtering for parent Command IDs for targeting and testing.
- Changed [`etcha local`]({{< ref "/docs/references/cli#local" >}}) to support rendering and running ad-hoc Jsonnet.  See [Render and Run]({{< ref "/docs/guides/running-patterns#render-and-run" >}}) for more information.
- Changed `etcha push` and `etcha run` to include the raw JWT as a var.
- Fixed [`etcha lint`]({{< ref "/docs/references/cli#lint" >}}) not excluding directories correctly.

## Removals

- Removed [`sources.runAll`]({{< ref "/docs/references/config#runall" >}}) toggle, Patterns will always run all Commands by default.
