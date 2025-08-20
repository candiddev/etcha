---
categories:
- reference
description: Reference documentation for Etcha's Jsonnet libraries.
title: Libraries
---

Etcha ships with a various Jsonnet libraries.  These libraries extend the [Jsonnet standard library]({{< ref "/docs/references/jsonnet#standard-library" >}}) to provide additional functions and [Commands]({{< ref "/docs/references/commands" >}}).  These libraries are created using {{% cli init %}}.

## Native Functions

[Native Functions]({{< ref "/docs/references/jsonnet#native-functions" >}}) are custom Jsonnet functions provided by the Etcha binary.  Etcha provides these as an importable helper file via init under `lib/etcha/native.libsonnet`.

## Commands

Etcha contains numerous importable [Command]({{< ref "/docs/references/commands" >}}) functions that will template out Commands.  We are constantly adding more of these and encourage the community to contribute others.  All Commands shipped in Etcha are linted and tested.

See [Writing Patterns]({{< ref "/docs/guides/writing-patterns" >}}) for more information about initializing these files.

Think we're missing something?  [Request a library on GitHub](https://github.com/candiddev/etcha/discussions/categories/feedback).

{{% etcha-library %}}
