---
categories:
- guide
description: How to use Etcha to run commands without a Patten.
title: Running Commands
weight: 60
---

In this guide, we'll go over running Commands without Patterns.

## Use Cases

Etcha can run Commands via [Source's `commands`]({{< ref "/docs/references/config#commands" >}}) and [`etcha push`]({{< ref "/docs/references/cli#push" >}}) in an ad-hoc way.  Some reasons you might want to use this include:

- Executing long running tasks
- Remote troubleshooting and debugging
- Statically defining Commands to run for event handlers

## Static Source Commands

Static source Commands allow Etcha to run Commands for sources without having them pushed/pulled via Patterns.  Instead, the Commands live within Etcha's main configuration.  The Commands are defined under a [`source`]({{< ref "/docs/references/config#sources" >}}) config block like this:

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

- Etcha will run Source Commands at startup unless the source is set to [`triggerOnly`]({{< ref "/docs/references/config#triggeronly">}}).
- Source Commands will be overwritten by any Patterns Etcha has cached from a previous pull/push before being ran at startup.
- Source Commands obey [`checkOnly`]({{< ref "/docs/references/config#checkonly" >}}).
- Patterns can be pulled/pushed with Source Commands.  These will overwrite the Source Commands and trigger Remove (unless [`noRemove`]({{< ref "/docs/references/config#noremove" >}}) is set)
- Source Commands will run periodically using [`runFrequencySec`]({{< ref "/docs/references/config#runfrequencysec" >}})

## Push Commands

Using push-commands is similar to running `ansible -a <command>`, expect it uses Etcha's push functionality instead of SSH.

The sender and receiver need to have certain configurations before it will work:

### Receiver

The receiver of Push Commands needs to have certain configuration values set:

- Add [`verifyKeys`]({{< ref "/docs/references/config#verifyKeys" >}})
- Configure a [`source`]({{< ref "/docs/references/config#sources" >}}) with the following options:
  - **Required**:
    - [`allowPush`]({{< ref "/docs/references/config#allowPush" >}}) set to `true`, this enables pushing.
  - **Recommended**:
    - [`noRemove`]({{< ref "/docs/references/config#noRemove" >}}) set to `true`, this prevents `remove` from being ran.
    - [`noRestore`]({{< ref "/docs/references/config#noRestore" >}}) set to `true`, this prevents Etcha from running your most recently pushed Command at every startup.
  - **Optional**:
    - [`runMulti`]({{< ref "/docs/references/config#runMulti" >}}) set to `true`, this allows Etcha to run multiple push Command  requests concurrently, otherwise they will be queued.

### Sender

The sender of Push Commands needs to have a corresponding [`signingKey`]({{< ref "/docs/references/config#signingKey" >}}) configured.  Optionally, [`pushTLSSkipVerify`]({{< ref "/docs/references/config#pushTLSSkipVerify" >}}) can be set to `true`, but it may impact security.

### Pushing Commands

Here is an example push from the Sender:

```bash
$ etcha push ls https://etcha.local:4000/etcha/v1/mysource
README.md
```

In this example:

- `ls` is the command we want to run on the remote instance
- `https://etcha.local:4000` is the address of the remote instance
- `mysource` is the source on the remote instance we should push to
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
