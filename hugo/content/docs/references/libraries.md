---
categories:
- reference
description: Reference documentation for Etcha's Jsonnet libraries.
title: Libraries
---

Etcha ships with a various Jsonnet libraries.  These libraries extend the [Jsonnet standard library]({{< ref "/docs/references/jsonnet#standard-library" >}}) to provide additional functions and [Commands]({{< ref "/docs/references/commands" >}}).  These libraries can be selectively downloaded from [GitHub](https://github.com/candiddev/etcha/tree/main/go/initdir/lib/etcha) or using [`etcha init`]({{< ref "/docs/references/cli#init" >}}).

## Native Functions

[Native Functions]({{< ref "/docs/references/jsonnet#native-functions" >}}) are custom Jsonnet functions provided by the Etcha binary.  Etcha provides these as an importable helper file via init under `lib/etcha/native.libsonnet` or from [GitHub](https://github.com/candiddev/shared/go/jsonnet/native.libsonnet).

## Commands

Etcha contains numerous importable [Command]({{< ref "/docs/references/commands" >}}) functions that will template out Commands.  We are constantly adding more of these and encourage the community to contribute others.  All Commands shipped in Etcha are linted and tested.

See [Writing Patterns]({{< ref "/docs/guides/writing-patterns" >}}) for more information about initializing these files.

### `apt`
{{% etcha-library "apt" %}}

### `aptKey`
{{% etcha-library "aptKey" %}}

### `copy`
{{% etcha-library "copy" %}}

### `dir`
{{% etcha-library "dir" %}}

### `etchaInstall`
{{% etcha-library "etchaInstall" %}}

### `file`
{{% etcha-library "file" %}}

### `group`
{{% etcha-library "group" %}}

### `line`
{{% etcha-library "line" %}}

### `mount`
{{% etcha-library "mount" %}}

### `symlink`
{{% etcha-library "symlink" %}}

### `systemdUnit`
{{% etcha-library "systemdUnit" %}}

### `user`
{{% etcha-library "user" %}}
