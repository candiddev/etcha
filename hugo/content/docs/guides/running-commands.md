---
categories:
- guide
description: How to use Etcha to run commands without a Patten.
title: Running Commands
weight: 60
---

In this guide, we'll go over running Commands without Patterns.

## Use Cases

Etcha can run Commands via [`etcha push-command`](../../references/cli#push-command) in an ad-hoc way.  Some reasons you might want to use this include:

- Executing long running tasks
- Remote troubleshooting and debugging

Using push-commands is similar to running `ansible -a <command>`, expect it uses Etcha's push functionality instead of SSH.

## Configuring Etcha for Push Commands

The sender and receiver need to have certain configurations before it will work:

### Receiver

The receiver of Push Commands needs to have certain configuration values set:

- Add [`verifyKeys`](../../references/config#verifyKeys)
- Configure a [`source`](../../references/config#sources) with the following options:
  - **Required**:
    - [`allowPush`](../../references#allowPush) set to `true`, this enables pushing.
  - **Recommended**:
    - [`noRemove`](../../references#noRemove) set to `true`, this prevents `remove` from being ran.
    - [`noRestore`](../../references#noRestore) set to `true`, this prevents Etcha from running your most recently pushed Command at every startup.
    - [`runAll`](../../references#runAll) set to `true`, this forces Etcha to always run your push Command (otherwise repeated Commands will only run once).
  - **Optional**:
    - [`runMulti`](../../references#runMulti) set to `true`, this allows Etcha to run multiple push Command  requests concurrently, otherwise they will be queued.

### Sender

The sender of Push Commands needs to have a corresponding [`signingKey`](../../references/config#signingKey) configured.  Optionally, [`pushTLSSkipVerify`](../../references/config#pushTLSSkipVerify) can be set to `true`, but it may impact security.

## Pushing Commands

Here is an example push from the Sender:

```bash
$ etcha push-command ls https://etcha.local:4000/etcha/v1/mysource
README.md
```

In this example:

- `ls` is the command we want to run on the remote instance
- `https://etcha.local:4000` is the address of the remote instance
- `mysource` is the source on the remote instance we should push to
- `README.md` is the output of the `ls` command __on the remote instnace__

## Precautions

Using `etcha push-command` can cause problems when used mixed with Sources that use Patterns.  `push-command` effectively pushes a new [Pattern](../../references/patterns) with the following format:

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

We recommend using dedicated `sources` for `push-command` and only using it for break-glass scenarios, but advanced users may be able to use it for PAM as well.
