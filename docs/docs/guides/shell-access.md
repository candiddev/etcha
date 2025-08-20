---
categories:
- guide
description: How to open a shell using Etcha.
title: Shell Access
weight: 60
---

Etcha can open an interactive shell like Bash on remote Etcha instances for debugging and terminal access.  This can replace SSH access to instances with Etcha-based authentication.

{{% alert title="Monitoring" color="primary" %}}
Make sure to checkout [Monitoring]({{< ref "/docs/guides/monitoring" >}}) for an overview of how to monitor Etcha runs.  You can trigger Prometheus alerts when Commands run or fail, and track who is doing what.
{{% /alert %}}

## Architecture

Etcha shell access uses Server-Sent Events (SSE) and API calls to create an interactive experience.  An SSE session is created when the shell starts, and stderr/stdout is sent over it.  All user commands are sent as HTTP POST calls to the shell endpoint and passed as stdin to the shell.

## Security

Shell access uses signed JWTs similar to Patterns, where the JWTs are verified by the remote Etcha instance using `verifyKeys` or `verifyCommands`.  Once a shell session is started, a random ID is generated and exchanged with the client for sending commands--these requests are still sent using JWTs and verified as well.  All JWTs are sent with low expirations (5 seconds) to limit replability, and the server will periodically (at least every 10 seconds) send a random keepalive string to the client to provide keystroke timing obfuscation.

## Setup

Etcha must be already running on the remote instance in [Listen Mode]({{< ref "/docs/guides/running-patterns#remote-run" >}}) with appropriate `verifyKeys` or `verifyCommands`.  The remote instance will additionally need [Sources]({{< ref "/docs/guides/running-patterns#sources" >}}) configured to allow Shell access by specifying a {{% config sources_shell %}}:

```json
{
  "sources": {
    "debug": {
      "exec": {
        "group": "operator",
        "user": "operator",
      },
      "shell": "/bin/bash",
    },
  }
}
```

In this example, the source `debug` is allowed shell access and will execute the shell `/bin/bash`.  Shell obeys `exec` values in the Source (and the global exec), so the shell will be executed under the `operator` user and group.

## Access

We can now start a shell from a different Etcha instance with a matching key.  The remote instance can use [Targets]({{< ref "/docs/guides/running-patterns#targets" >}}) or specifying a host directly:

```bash
$ etcha shell debug myhost
operator@myhost~$
```

{{% alert title="Candid Commentary" color="info" %}}
We plan on adding remote file transfers by Q2 2025.
{{% /alert %}}
