---
author: Candid Development
date: 2024-11-13
description: Release notes for Etcha v2024.11.
tags:
  - release
title: "What's New in Etcha: v2024.11"
type: blog
---

## Enhancements

- Added {{% config "reload functionality" configreloadsec %}}.
- Added additional {{% config "http configurations" http %}}.
- Updated library dependencies to latest version.

## Deprecations

- Deprecated config value `run.listenAddress`, please use {{% config http.listenAddress listenaddress %}}.
- Deprecated config value `run.rateLimiterRate`, please use {{% config http.rateLimitPatterns ratelimitpatterns %}}.
- Deprecated config value `run.tlsCertificateBase64`, please use {{% config http.tlsCertificateBase64 tlscertificatebase64 %}}.
- Deprecated config value `run.tlsCertificatePath`, please use {{% config http.tlsCertificatePath tlscertificatepath %}}.
- Deprecated config value `run.tlsKeyBase64`, please use {{% config http.tlsKeyBase64 tlskeybase64 %}}.
- Deprecated config value `run.tlsKeyBasePath`, please use {{% config http.tlsKeyBasePath tlskeybasepath %}}.
