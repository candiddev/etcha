---
categories:
- guide
description: How to upgrade EtchaOS.
title: Upgrade
type: docs
weight: 30
---

EtchaOS is delivered as two files: initrd containing the root filesystem, and vmlinuz containing the kernel image.  The upgrade process for EtchaOS involves replacing these files with newer versions (or a different [Variant]({{% ref "/etchaos/references/download" %}}), if you want to try something new).

## Automatic Upgrade

EtchaOS disk images come with an EtchaOS {{% config sources %}} configuration under `/boot/efi/etcha.libsonnet`.  The source periodically pulls the latest EtchaOS JWT from `https://etcha.dev/releases` for your current EtchaOS arch and variant, checks it against the EtchaOS signing key, and updates EtchaOS if it is changed.

### Config

This configuration **is not enabled by default**.  You must merge the source into one of your configs above to enable it, i.e.:

```
std.mergePatch((import '/boot/efi/etchaos.libsonnet'), {
  ...your config...
})
```

You can modify/override this source in your config to change various aspects:

- Point {{% config sources_pullPaths %}} to an internal server your control (default: `https://etcha.dev/releases/etchaos_<variant name>_<variant arch>.jwt`)
- Change {{% config sources_runFrequencySec %}} to different interval (default: `86400`)

### Variables

Additionally, you can set the following {{% config vars %}} in your config to control the behavior of the upgrade process:

- `etchaosArch` controls the [Variant]({{% ref "/etchaos/references/download" %}}) architecture that will be installed.  Changing this is not recommended.
- `etchaosReboot` is a timeframe provided to the reboot command after EtchaOS is upgraded.  The current reboot command is `systemd-run -u echaos-reboot %s reboot` where `%s` is replaced with the value of this variable.  The default value is `--on-active=1s` (reboot immediately).  If this variable is set to `never`, EtchaOS will never reboot after an upgrade.  The syntax is described in [systemd.time](https://www.freedesktop.org/software/systemd/man/latest/systemd.time.html).  A calendar value can be provided, i.e. `--on-calendar=04:00` to reboot at 4 AM.  To cancel the reboot, run `systemctl stop etchaos-reboot`.  To view the reboot status, run `systemctl status etchaos-reboot.timer`.
- `etchaosURL` sets the base URL to download the EtchaOS files (like `initrd` and `vmlinuz`).  This can be repointed to an internal server to distribute EtchaOS files from.
- `etchaosVariant` sets the [Variant]({{% ref "/etchaos/references/download" %}}) type to be installed.  This can be changed to switch between variants.

### Manual Upgrades

EtchaOS can be manually upgraded by replacing the `initrd` and `vmlinuz` files with newer versions.
