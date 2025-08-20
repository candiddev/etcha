---
categories:
- guide
description: How to configure EtchaOS for common scenarios.
title: Configure
type: docs
weight: 20
---

EtchaOS can be configured for a variety of use cases using [Etcha Patterns]({{< ref "/docs/explanations/architecture" >}}).  EtchaOS can be have patterns [pushed]({{< ref "/docs/guides/running-patterns#remote-push" >}}), [pulled]({{< ref "/docs/guides/running-patterns#remote-pull" >}}), or [statically defined]({{< ref "/docs/guides/running-commands#static-source-commands" >}}) within a configuration.

This guide provides Commands for common scenarios.  These Commands can be used within a Pattern or defined within a {{< config "source config" sources >}}.

{{% alert color="primary" title="Need Help?" %}}
Looking to run your app, IoT platform, Kubernetes, or databases? \
We can build your configurations or train and support your team. \
{{% contactus "Custom%20EtchaOS%20Configuration" %}} for consulting and support.
{{% /alert %}}

## Add Password for Root

EtchaOS can only be accessed using [Etcha Shell]({{< ref "/docs/guides/shell-access" >}}) out of the box.  You can use this Command to set a root password for console login, replacing 123 with the actual password.  Care should be taken as this will store the password in plaintext, consider using the [`user` library]({{< ref "/docs/references/libraries#user" >}}) to set a password using a hash.

```
{
  id: 'set passwd',
  always: true,
  change: 'echo -e \'123\n123\n\' | passwd',
}
```

## Add Swap

One of the common concerns with EtchaOS is having everything run in memory.  Depending on the [Variant]({{< ref "/etchaos/references/download" >}}), this could be anywhere from 400-600MB+ of RAM in use by EtchaOS.  An easy fix for this is to add a swap disk or file--Linux will seamlessly move unused files like the EtchaOS data to swap, freeing up your RAM for other things.

This is an example Command to add swap, replace the target partition (`/dev/vda1`) with the correct partition or disk:

```
{
  id: 'swap',
  commands: [
    {
      id: 'mkswap'
      check: '[[ -e /dev/disk/by-label/swap ]]',
      change: 'mkswap -L swap /dev/vda1',
    },
    {
      id: 'mount'
      check: 'swapon -s | grep /dev/vda1',
      change: 'swapon /dev/vda1',
    },
  ],
}
```

## Make `/usr` Writable

The `/usr` directory is read-only by default.  EtchaOS comes with a script, `/usr/sbin/usrmount`, which accepts two arguments: an upper directory and a work directory.  This script will use those arguments to mount an OverlayFS volume in place of `/usr`, making it read-write.  Persisting these changes across EtchaOS upgrades may cause undefined behavior.

This is an example Command to make `/usr` writable, replace the target directories (`/mnt/usr_upper` and `/mnt/usr_work`) with the correct ones:

```
{
  id: 'mount usr',
  check: '[[ -d '/mnt/usr_upper' ]],
  change: 'usrmount /mnt/usr_upper /mnt/usr_work',
}
```
