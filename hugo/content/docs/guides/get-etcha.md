---
categories:
- guide
description: How to install Etcha.
title: Install Etcha
weight: 1
---

Installing Etcha depends on how you want to run it.  Etcha is available as a [binary](#binary) or a [container](#container).

## Binary

Etcha binaries are available on [GitHub](https://github.com/candiddev/etcha/releases).

{{< tabpane text=true >}}
{{< tab header="Linux amd64" >}}
{{< highlight bash >}}
curl -L https://github.com/candiddev/etcha/releases/latest/download/etcha_linux_amd64.tar.gz -O
curl -L https://github.com/candiddev/etcha/releases/latest/download/etcha_linux_amd64.tar.gz.sha256 -O
sha256sum -c etcha_linux_amd64.tar.gz.sha256
tar -xzf etcha_linux_amd64.tar.gz
{{< /highlight >}}
{{< /tab >}}

{{< tab header="Linux arm" >}}
{{< highlight bash >}}
curl -L https://github.com/candiddev/etcha/releases/latest/download/etcha_linux_arm.tar.gz -O
curl -L https://github.com/candiddev/etcha/releases/latest/download/etcha_linux_arm.tar.gz.sha256 -O
sha256sum -c etcha_linux_arm.tar.gz.sha256
tar -xzf etcha_linux_arm.tar.gz
{{< /highlight >}}
{{< /tab >}}

{{< tab header="Linux arm64" >}}
{{< highlight bash >}}
curl -L https://github.com/candiddev/etcha/releases/latest/download/etcha_linux_arm64.tar.gz -O
curl -L https://github.com/candiddev/etcha/releases/latest/download/etcha_linux_arm64.tar.gz.sha256 -O
sha256sum -c etcha_linux_arm64.tar.gz.sha256
tar -xzf etcha_linux_arm64.tar.gz
{{< /tab >}}
{{< /highlight >}}

{{< tab header="macOS amd64" >}}
{{< highlight bash >}}
curl -L https://github.com/candiddev/etcha/releases/latest/download/etcha_darwin_amd64.tar.gz -O
curl -L https://github.com/candiddev/etcha/releases/latest/download/etcha_darwin_amd64.tar.gz.sha256 -O
sha256sum -c etcha_darwin_amd64.tar.gz.sha256
tar -xzf etcha_darwin_amd64.tar.gz
{{< /highlight >}}
{{< /tab >}}

{{< tab header="macOS arm64" >}}
{{< highlight bash >}}
curl -L https://github.com/candiddev/etcha/releases/latest/download/etcha_darwin_arm64.tar.gz -O
curl -L https://github.com/candiddev/etcha/releases/latest/download/etcha_darwin_arm64.tar.gz.sha256 -O
sha256sum -c etcha_darwin_arm64.tar.gz.sha256
tar -xzf etcha_darwin_arm64.tar.gz
{{< /highlight >}}
{{< /tab >}}
{{< /tabpane >}}


{{% alert title="Updating Etcha" color="info" %}}
Etcha can be updated by replacing the binary with the latest version.
{{% /alert %}}

## Container

Etcha containers are available on [GitHub](https://github.com/candiddev/etcha/pkgs/container/etcha).  They include Etcha and Busybox.

You can create an alias to run Etcha as a container:

{{< highlight bash >}}
alias etcha='docker run -u $(id -u):$(id -g) -it --rm -v $(pwd):/work -w /work ghcr.io/candiddev/etcha:latest'
{{< /highlight >}}
