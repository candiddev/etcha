---
author: Mike
date: 2024-01-01
description: Release notes for Etcha v2024.01.
tags:
  - release
title: "What's New in Etcha: v2024.01"
type: blog
---

{{< etcha-release version="2024.01" >}}

## Fixes

- CLI: Fixed [`etcha show-pattern`](../../docs/references/cli#show-pattern) not parsing JWTs correctly.
- Patterns: Fixed Pattern imports using the Jsonnet function `getEnv` not obeying the `exec.envInherit` or `exec.env` values within the main configuration or source configuration.
- Sources: Fixed Pattern loading and execution at startup not obeying naming order of sources.
