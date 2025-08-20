---
categories:
- guide
description: How to install Etcha.
title: Install Etcha
weight: 1
---

Installing Etcha depends on how you want to run it.  Etcha is available as a [binary](#binary) or a [container](#container).

## Binary

Rot binaries are available for various architectures and operating systems:

{{% release %}}

{{% alert title="Updating Etcha" color="primary" %}}
Etcha can be updated by replacing the binary with the latest version.
{{% /alert %}}

Etcha can be ran as a service:

{{< tabpane text=true >}}
{{% tab header="Linux/systemd" %}}
1. Download the Etcha binary to `/usr/local/bin/etcha` and mark it as executable.
2. Create `/etc/systemd/system/etcha.service` with this content:

{{< highlight systemd >}}
[Unit]
Description=Infinite scale configuration management for distributed platforms
Documentation=https://etcha.dev
After=network.target

[Service]
ExecStart=/usr/local/bin/etcha -c /etc/etcha.jsonnet run
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
{{< /highlight >}}

3. Start Etcha: `systemctl enable --now etcha.service`
{{% /tab %}}

{{% tab header="macOS/launchd" %}}
1. Download the Etcha binary to `/usr/local/bin/etcha` and mark it as executable.
2. Create `/Library/LaunchDaemons/dev.etcha.plist` with this content:

{{< highlight xml >}}
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.etcha</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/etcha</string>
        <string>-c</string>
        <string>/etc/etcha.jsonnet</string>
        <string>run</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>/var/lib/etcha/etcha.log</string>
    <key>StandardErrorPath</key>
    <string>/var/lib/etcha/etcha.err</string>
    <key>WorkingDirectory</key>
    <string>/var/lib/etcha</string>
</dict>
</plist>
{{< /highlight >}}

3. Start Etcha: `launchctl load /Library/LaunchDaemons/com.etcha.plist`
{{% /tab %}}
{{< /tabpane >}}


## Container

Etcha containers are available on [GitHub](https://github.com/candiddev/etcha/pkgs/container/etcha).

You can create an alias to run Etcha as a container:

{{< highlight bash >}}
alias etcha='docker run -u $(id -u):$(id -g) -it --rm -v $(pwd):/work -w /work ghcr.io/candiddev/etcha:latest'
{{< /highlight >}}

## SBOM

Etcha ships with a Software Bill of Materials (SBOM) manifest generated using [CycloneDX](https://cyclonedx.org/).  The `.bom.json` manifest is available with the other [Binary Assets](#binary).
