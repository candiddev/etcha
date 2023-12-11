---
author: Mike
date: 2023-12-11
description: Release notes for Etcha v2023.12.
tags:
  - release
title: "What's New in Etcha: v2023.12"
type: blog
---

{{< etcha-release version="2023.12" >}}

## Features

- Config: [`cli.configPath`](../../docs/references/config#configpath) can be used to set a custom configuration path, along with [`etcha -c`](../../docs/references/cli#c)

## Enhancements

- CLI: [`etcha generate-keys`](../../docs/references/cli#generate-keys) can generate PBKDF-encrypted signing keys
- Config: [`build.signingKey`](../../docs/references/config#signingkey) can be encrypted using a Password Based Key Derivation Function (PBKDF)

## Removals

- CLI: [`etcha -c`](../../docs/references/cli#c) no longer accepts a comma separated list of configuration files.  Instead, you can import the configuration files into one Jsonnet file.
