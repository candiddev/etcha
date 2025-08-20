---
categories:
- guide
description: How to use Etcha to run commands without a Patten.
title: Running Commands
weight: 60
---

In this guide, we'll go over running Commands without Patterns.

## Use Cases

Etcha can run Commands via Source's {{% config exec_commands %}} and {{% cli push %}} in an ad-hoc way.  Some reasons you might want to use this include:

- Executing long running tasks
- Remote troubleshooting and debugging
- Statically defining Commands to run for event handlers

## Static Source Commands

Static source Commands allow Etcha to run Commands for sources without having them pushed/pulled via Patterns.  Instead, the Commands live within Etcha's main configuration.  The Commands are defined under a {{% config sources %}} config block like this:

```json
{
  "sources": {
    "100reboot": {
      "commands": [
        {
          "id": "reboot",
          "change": "shutdown -r now"
        }
      ]
    }
  }
}
```

When using Source Commands:

- Etcha will run Source Commands at startup unless the source is set to {{% config sources_triggerOnly %}}.
- Source Commands will be overwritten by any Patterns Etcha has cached from a previous pull/push before being ran at startup.
- Source Commands obey {{% config sources_checkOnly %}}.
- Patterns can be pulled/pushed with Source Commands.  These will overwrite the Source Commands and trigger Remove (unless {{% config sources_noRemove %}}) is set)
- Source Commands will run periodically using {{% config sources_runFrequencySec %}}

## Push Commands

Using push-commands is similar to running `ansible -a <command>`, expect it uses Etcha's push functionality instead of SSH.

{{% alert title="Candid Commentary" color="info" %}}
Push Commands are a great way to do deployments.
{{% /alert %}}

The sender and receiver need to have certain configurations before it will work:

### Receiver

The receiver of Push Commands needs to have certain configuration values set:

- {{% config run_verifyKeys %}}
- Configure a {{% config sources %}} with the following options:
  - **Required**:
    - {{% config sources_allowPush %}} set to `true`, this enables pushing.
  - **Recommended**:
    - {{% config sources_noRemove %}} set to `true`, this prevents `remove` from being ran.
    - {{% config sources_noRestore %}} set to `true`, this prevents Etcha from running your most recently pushed Command at every startup.
  - **Optional**:
    - {{% config sources_runMulti %}} set to `true`, this allows Etcha to run multiple push Command  requests concurrently, otherwise they will be queued.

### Sender

The sender of Push Commands needs to have a corresponding {{% config build_signingKey %}} configured.  Optionally, {{% config httpClient_tlsSkipVerify %}} can be set to `true`, but it may impact security.

### Pushing Commands

Here is an example push from the Sender:

```bash
$ etcha -x httpClient_tlsSkipVerify=true push -h etcha.local mysource ls
etcha.local:
    README.md
```

In this example:

- `httpClient_tlsSkipVerify=true` skips TLS certificate checking (Etcha uses self signed certificates by default)
- `etcha.local` is the address of the remote instance
- `mysource` is the source on the remote instance we should push to
- `ls` is the command we want to run on the remote instance
- `README.md` is the output of the `ls` command __on the remote instnace__

### Precautions

Using `etcha push` can cause problems when used mixed with Sources that use Patterns.  `push` effectively pushes a new [Pattern]({{< ref "/docs/references/patterns" >}}) with the following format:

```json
{
  "run": [
    {
      "always": true,
      "change": "ls",
      "id": "etcha push"
    }
  ]
}
```

This will replace the current Pattern.  If you do not set `noRemove` in the destination Source config, the replaced Pattern will be diff'd and the `remove` Commands will be ran.  Additionally, subsequent Pattern pushes will most likely trigger their `change` values.

We recommend using dedicated `sources` for `push` and only using it for break-glass scenarios, but advanced users may be able to use it for PAM as well.

{{% alert title="Candid Commentary" color="info" %}}
We use push in combination with limited-access users via {{% config exec_user %}}
{{% /alert %}}
