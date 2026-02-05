---
description: Download links for EtchaOS Variants.
type: docs
title: Download
---

EtchaOS is distributed as "Variants"--different flavors of EtchaOS represented by their base OS.  The configurations for these Variants start out different but eventually converge into a common configuration and finalization process.

{{% alert color="warning" title="Custom Variants" %}}
Let us design a custom Variant for your software or application. \
{{% contactus "Custom%20EtchaOS%20Variant" %}} for consulting and support.
{{% /alert %}}

## Comparison

| Variant | Base OS | amd64 | arm64 | UEFI | BIOS | Secure Boot | Container Engine | Container Runtime Interface (CRI) |
|-|-|-|-|-|-|-|-|-|-|
| arc | Arch Linux | {{% asterisk %}} | Not Supported | {{% check %}} | {{% check %}} | {{% check %}} | docker | containerd |
| alm | AlmaLinux 9 | {{% asterisk %}} | {{% asterisk %}} | {{% check %}} | {{% check %}} | {{% check %}} | docker | containerd |
| deb | Debian 13 "trixie" | {{% check %}} | {{% check %}} | {{% check %}} | {{% check %}} | {{% check %}} | docker | containerd |
| fed | Fedora Linux 42 | {{% check %}} | {{% check %}} | {{% check %}} | {{% asterisk %}} | {{% check %}} | docker | containerd |
| sus | openSUSE Leap 15 | {{% asterisk %}} | {{% asterisk %}} | {{% check %}} | {{% check %}} | {{% check %}} | docker | containerd |
| ubu | Ubuntu 24.04 "Noble Numbat" | {{% check %}} | {{% asterisk %}} | {{% check %}} | {{% check %}} | {{% check %}} | docker | containerd |

{{% asterisk %}} - Requires customizations or considerations before using.  {{%  contactus "EtchaOS%20Varian%20Support" %}} for access.

## Download

{{% etchaos-variants-download %}}
