---
author: Mike
date: 2023-11-07
description: Release notes for Etcha v2023.11.
tags:
  - release
title: "What's New in Etcha: v2023.11"
type: blog
---

{{< etcha-release version="2023.11" >}}

## Enhancements

- Commands can now specify `onRemove` to trigger other Commands during removal.
- Remove mode will now run `check` Commands if `always` is false.  If `check` Commands do not return an error, `remove` will run.

## Removals

- Removed `local-check`, `change` and `remove` modes can be forced into check-only by specifying `checkOnly` in `sources`.
