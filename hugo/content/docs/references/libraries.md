---
categories:
- reference
description: Reference documentation for Etcha's Jsonnet libraries.
title: Libraries
---

Etcha ships with a various Jsonnet libraries.  These libraries extend the [Jsonnet standard library](../jsonnet#standard-library) to provide additional functions and [Commands](../commands).  These libraries can be selectively downloaded from [GitHub](https://github.com/candiddev/etcha/tree/main/go/initdir/lib/etcha) or using [`etcha init`](../cli#init).

## Native Functions

[Native Functions](../jsonnet#native-functions) are custom Jsonnet functions provided by the Etcha binary.  Etcha provides these as an importable helper file via init under `lib/etcha/native.libsonnet` or from [GitHub](https://github.com/candiddev/shared/go/jsonnet/native.libsonnet).

## Commands

Etcha contains numerous importable [Command](../commands) functions that will template out Commands.  We are constantly adding more of these and encourage the community to contribute others.  All Commands shipped in Etcha are linted and tested.

See [Writing Patterns](../../guides/writing-patterns) for more information about initializing these files.

### `apt`
{{% etcha-library "apt" %}}

### `aptKey`
{{% etcha-library "aptKey" %}}

### `copy`
{{% etcha-library "copy" %}}

### `dir`
{{% etcha-library "dir" %}}

### `file`
{{% etcha-library "file" %}}

### `mount`
{{% etcha-library "mount" %}}

### `password`
{{% etcha-library "password" %}}

### `symlink`
{{% etcha-library "symlink" %}}

### `systemdUnit`
{{% etcha-library "systemdUnit" %}}




