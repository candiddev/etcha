---
categories:
- guide
description: How to monitor Etcha.
title: Monitoring
weight: 70
---

Running Etcha in [Listen Mode]({{< ref "/docs/guides/running-patterns#listen-mode" >}}) can expose a metrics scrape endpoint, `/etcha/v1/system/metrics`.  This scrape endpoint is designed for use with tools like Prometheus.  Setting the configuration value for {{% config run_systemMetricsSecret %}} will enable this endpoint.

You can scrape this endpoint using a path like this:

`https://etcha:4000/etcha/v1/system/metrics?key=secret`

Or using a Prometheus `scrape_config` like this:

```yaml
- job_name: etcha
  metrics_path: /etcha/v1/system/metrics
  params:
    key:
    - secret
  relabel_configs:
  scheme: https
  static_configs:
    - targets:
      - "etcha:4000"
  tls_config: # Optional, otherwise you must disable TLS verification
    ca: <Etcha certificate>
```

Where `secret` is the value for `systemMetricsKey` and `:4000` is the {{% config httpServer_listenAddress %}}.

## Metrics

Etcha exposes the following metrics (along with the default `go` metrics):

### `etcha_commands_total`

Counter of Commands that have ran.

Labels:
- `error`: If the Command has errors (0=no, 1=yes)
- `id`: [Command]({{< ref "/docs/references/commands" >}}) ID
- `mode`: Mode that was executed (`changed`, `check`, `remove`)
- `parent_id`: [Command]({{< ref "/docs/references/commands" >}}) parent ID
- `public_key`: Public key ID of JWT signer
- `remote_addr`: Remote address of JWT source (if pushed)
- `source`: {{% config sources %}} Name

### `etcha_sources_commands`

Gauage of Source commands that have ran.

Labels:
- `mode`: Mode that was executed (`changed`, `remove`)
- `name`: {{% config sources %}} Name
- `public_key`: Public key ID of JWT signer
- `remote_addr`: Remote address of JWT source (if pushed)
- `trigger`: What triggered the source (`event`, `pull`, `push`, `webhook`)

### `etcha_sources_shells`

Gauage of Source shells.  A value of 1 means the shell is active, a value of 0 means the shell is not active.

Labels:
- `name`: {{% config sources %}} Name
- `public_key`: Public key ID of the shell initiator
- `remote_addr`: Remote address of shell initiator

### `etcha_sources_total`

Counter of Sources that have ran.

Labels:
- `error`: If the Source has errors (0=no, 1=yes)
- `name`: {{% config sources %}} Name
- `public_key`: Public key ID of JWT signer
- `remote_addr`: Remote address of JWT source (if pushed)
- `trigger`: What triggered the source (`event`, `pull`, `push`, `webhook`)

## Example Rules

Here are some example Prometheus rules for sending alerts when things happen in Etcha.

### Alert When a Shell Session Starts

```yaml
annotations:
  instance: '{{ $labels.instance }}'
  public_key: '{{ $labels.public_key }}'
  remote_addr: '{{ $labels.remote_addr }}'
  source: '{{ $labels.name }}'
  summary: '{{ $labels.instance }} has an active Etcha shell session'
expr: 'etcha_sources_shells == 1'
```

### Alert When a Source is Pushed

```yaml
annotations:
  instance: '{{ $labels.node }}'
  public_key: '{{ $labels.public_key }}'
  remote_addr: '{{ $labels.remote_addr }}'
  source: '{{ $labels.name }}'
  summary: '{{ $labels.node }} ran a Pattern'
  trigger: '{{ $labels.trigger }}'
expr: 'changes(etcha_sources_total[2m]) != 0'
```
