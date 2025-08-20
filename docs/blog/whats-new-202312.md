---
author: Candid Development
date: 2023-12-11
description: Release notes for Etcha v2023.12.
tags:
  - release
title: "What's New in Etcha: v2023.12"
type: blog
---

## Features

- Config: {{% config configPath configPath %}} can be used to set a custom configuration path, along with {{% cli c %}}.
- Config: {{% config sources.commands sourcecommands %}} can be used to set a static list of Commands for sources.

## Enhancements

- CLI: {{% cli gen-keys %}} can generate PBKDF-encrypted signing keys
- Config: {{% config build.signingKey signingkey %}}) can be encrypted using a Password Based Key Derivation Function (PBKDF)

## Removals

- CLI: {{% cli c %}} no longer accepts a comma separated list of configuration files.  Instead, you can import the configuration files into one Jsonnet file.
