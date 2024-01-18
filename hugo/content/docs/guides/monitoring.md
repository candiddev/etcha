---
categories:
- guide
description: How to monitor Etcha.
title: Monitoring
weight: 70
---

Running Etcha in [Listen Mode]({{< ref "/docs/guides/running-patterns#listen-mode" >}}) can expose a metrics scrape endpoint, `/etcha/v1/system/metrics`.  This scrape endpoint is designed for use with tools like Prometheus.  Setting the configuration value for [`systemMetricsSecret`]({{< ref "/docs/references/config#systemmetricssecret" >}}) will enable this endpoint.

You can scrape this endpoint using a path like this:

`https://etcha:4000/etcha/v1/system/metrics?key=secret`

Where `secret` is the value for `systemMetricsSecret` and `:4000` is the [`listenAddress`]({{< ref "/docs/references/config#listenaddress" >}}).


## Metrics

Etcha exposes the following metrics (along with the default `go` metrics):

### `etcha_commands_total`

Counter of Commands that have ran.

Labels:
- `error`: If the Command has errors (0=no, 1=yes)
- `id`:  [Command]({{< ref "/docs/references/commands" >}}) ID
- `mode`: Mode that was executed (`changed`, `check`, `remove`)
- `source`: [Source]({{< ref "/docs/references/config#source" >}}) Name

### `etcha_sources_total`

Counter of Sources that have ran.

Labels:
- `error`: If the Source has errors (0=no, 1=yes)
- `name`: [Source]({{< ref "/docs/references/config#source" >}}) Name
- `trigger`: What triggered the source (`event`, `pull`, `push`, `webhook`)

### `etcha_sources_commands`

Gauage of Source commands that have ran.

Labels:
- `mode`: Mode that was executed (`changed`, `remove`)
- `name`: [Source]({{< ref "/docs/references/config#source" >}}) Name
- `trigger`: What triggered the source (`event`, `pull`, `push`, `webhook`)
