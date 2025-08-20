---
author: Candid Development
date: 2024-08-13
description: Release notes for Etcha v2024.08.
tags:
  - release
title: "What's New in Etcha: v2024.08"
type: blog
---

## Enhancements

- Added more [metrics]({{< ref "/docs/guides/monitoring" >}}) and metrics labels.
- Added [pull]({{< ref "/docs/references/events#pull" >}}), [push]({{< ref "/docs/references/events#push" >}}), and [shell]({{< ref "/docs/references/events#shell" >}}) events for hooking in notifications or other commands after these events occur.
- Added [signingKeyPath]({{< ref "/docs/references/config#signingkeypath" >}}) to allow for using a file-based signingKey (like Rot).

## Fixes

- Fixed {{% config vars %}} not consistently being set in various rendering pipelines.
