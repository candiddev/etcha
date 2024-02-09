---
author: Mike
date: 2024-02-06
description: Release notes for Etcha v2024.02.
tags:
  - release
title: "What's New in Etcha: v2024.02"
type: blog
---

{{< etcha-release version="2024.02" >}}

## Enhancements

- Added [rotInstall]({{< ref "/docs/references/libraries#rotInstall" >}}) library for installing the latest version of [Rot](https://rotx.dev).
- Changed [Environment Variables]({{< ref "/docs/references/commands#environment-variables" >}}) to set `envPrefix` to `check` stderr/stdout if defined.
- Changed Push mode to run build commands.
- Changed RunEnv to RunVars--these can be used during Jsonnet rendering for variables defined at build time.  See [Patterns]({{< ref "/docs/references/patterns#runvars" >}}) for more information.
- Updated all Go libraries to latest version
