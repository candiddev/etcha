---
description: Reference documentation for EtchaOS configuration discovery process.
type: docs
title: Config Discovery
---

EtchaOS is configured via [Etcha]({{< ref "/docs" >}}), and the service starts as soon as the network is online (or after a 20 second timeout).  When the service starts, Etcha will parse a [config file]({{< ref "/docs/references/config" >}}) under `/etc/etcha.jsonnet`.

The service will keep restarting until a valid config is discovered:

- At least one {{% config run_verifyKey %}} is defined.
- At least one {{% config sources %}} is defined.  It can be a push, pull, or static Commands source.

## Config Discovery

EtchaOS's config imports JSON or Jsonnet from various sources, merging them onto the previous configuration.  This process is meant to support bare-metal, virtual machines, and cloud provider sources--if something is missing, please {{% contactus "EtchaOS%20Config%20Sources" %}}.

The current configuration looks like this:

```
{{% file "config.libsonnet" %}}
```

Lets break down the sources individually:

### 1. Base Configuration

Etcha is configured with a base configuration defining a {{% config run_stateDir %}} and a {{% config vars %}} for future usage by Patterns.

### 2. SMBIOS (QEMU) {#smbios}

Etcha then attempts to render a Jsonnet or JSON value from SMBIOS data located under `/sys/firmware/dmi/tables/DMI`.  It looks for a string starting with `etchaos=` and attempts to parse the remaining text.  If no text is found, or the config value `fallthrough` is `true`, Etcha tries to resolve configs using the next step.

The SMBIOS data is typically used by virtual machines, such as QEMU.  The value can be specified using the QEMU command line argument `-smbios type=11`.  It's recommended to put the SMBIOS data as a single line Jsonnet value and store it in a file.

Given a file named `/tmp/bios` with this content:
```
etchaos={fallthrough:true,sources:{push:{allowPush:true}}}
```

And using QEMU with this argument:
```
-smbios type=11,path=/tmp/smbios
```

Etcha would render and apply this config to the configuration at boot:

```json
{
  "fallthrough": true,
  "sources": {
    "push": {
      "allowPush": true
    }
  }
}
```

### 3. cidata (any) {#cidata}

If SMBIOS fails to resolve or allows fallthrough, Etcha then attempts to render a Jsonnet or JSON value from `cidata`:

- EtchaOS's `/etc/fstab` is configured to mount any disk with the label `cidata` to `/mnt/cidata`.
- Etcha will attempt to read `/mnt/cidata/user-data` and parse it as Jsonnet or JSON.
- If no text is found, or the config value `fallthrough` is `true`, Etcha tries to resolve configs using the next step.

You can create a `cidata` volume using almost any filesystem supported by Linux out of the box.  Most commonly, this volume is an `ext4`, `vfat`, or a CD-ROM, ISO, or floppy disk.

Given a file named `user-data` with these values:

```
{
  cli: {
    logLevel: 'debug',
  },
  fallthrough: true,
}
```

An ISO named `cidata.iso` can be created with the ID of `cidata` using `genisoimage`:

```
genisoimage -output cidata.iso -joliet -rock user-data
```

The ISO can be provided to QEMU with this argument:
```
-drive file=cidata.sio,media=cdrom
```

Etcha would render and apply this config to the configuration at boot:

```json
{
  "cli": {
    "logLevel": "debug"
  },
  "fallthrough": true,
}
```

### 4. Instance User Data (AWS, OpenStack) {#user-data}


If cidata fails to resolve or allows fallthrough, Etcha then attempts to render a Jsonnet or JSON value from Instance User Data:

- Etcha will attempt to read `http://169.254.169.254/latest/user-data` and parse it as Jsonnet or JSON.
- If no text is found, or the config value `fallthrough` is `true`, Etcha tries to resolve configs using the next step.

This is primarily used on Amazon Web Services (AWS) and OpenStack, however the same functionality can be replicated by routing the IP address `169.254.169.254` to a web server and serving a config file with the path `/latest/user-data`.

### 5. Instance Metadata (GCP)


If Instance User Data fails to resolve or allows fallthrough, Etcha then attempts to render a Jsonnet or JSON value from Instance Metadata on Google Cloud Platform (GCP):

- Etcha will attempt to read `http://169.254.169.254/latest/user-data` and parse it as Jsonnet or JSON.
- If no text is found, or the config value `fallthrough` is `true`, Etcha tries to resolve configs using the next step.

## EtchaOS Source

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
- `etchaosReboot` is a timeframe provided to the reboot command after EtchaOS is upgraded.  The current reboot command is `systemd-run --on-active=%s reboot` where `%s` is replaced with the value of this variable.  If this variable is set to `never`, EtchaOS will never reboot after an upgrade.  The syntax is described in [systemd.time](https://www.freedesktop.org/software/systemd/man/latest/systemd.time.html).
- `etchaosURL` sets the base URL to download the EtchaOS files (like `initrd` and `vmlinuz`).  This can be repointed to an internal server to distribute EtchaOS files from.
- `etchaosVariant` sets the [Variant]({{% ref "/etchaos/references/download" %}}) type to be installed.  This can be changed to switch between variants.
