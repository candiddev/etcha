---
author: Candid Development
date: 2024-05-07
description: Release notes for Etcha v2024.05.
tags:
  - release
title: "What's New in Etcha: v2024.05"
type: blog
---

## Enhancements

- Added a new library: [dnf]({{% ref "/docs/references/libraries#dnf" %}}).
- Added a new library: [container]({{% ref "/docs/references/libraries#container" %}}).
- Added a new library: [package]({{% ref "/docs/references/libraries#package" %}}).
- Etcha's CLI is now available in English, عربي, Deutsch, Español, Francais, हिन्दी, 日本語, Nederlands, Português, русский, and 中文.
- {{% cli dir %}} can now optionally purge non-empty directories.
- {{% cli shell %}} is now configured using an Exec value instead of a binary path.
- Command `onChange` and `onRemove` will now trigger if any sub `commands` change or remove.

## Fixes

- Fixed nested Command diffing returning duplicate removals
