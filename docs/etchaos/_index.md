---
description: Secure, minimal, immutable Linux for cloud, on-prem, or embedded. Automatic upgrades, flexible deployments, and consistent experience across various base distributions such as Debian or Fedora.
linkTitle: EtchaOS
menu: {main}
title: EtchaOS
type: docs
weight: 20
---

EtchaOS is a powerful, secure, and lightweight Linux distribution built for the cloud, on-premise deployments, and embedded systems. It prioritizes security and consistency, while offering incredible flexibility for your deployments.

## Features
- **Automatic Upgrades**\
Never worry about falling behind. EtchaOS can be configured for secure, air-gapped, automatic upgrades, on a schedule that fits your needs.
- **Etcha-Driven Customization**\
Etcha, the core of EtchaOS, allows for secure customization through push or pull configuration [Patterns]({{% ref "/docs/explanations/architecture" %}}).
- **Immutable Security**\
EtchaOS boots with an unmodified filesystem and a read-only `/usr` directory. Optionally, configure a fully read-only `/` for maximum security. Users can leverage bind or OverlayFS mounts for read-write functionality when needed.
- **In-Memory Efficiency**\
For ultimate performance, EtchaOS can run entirely in-memory, using a combination of static files/folders and SquashFS.
- **Minimal Footprint**\
EtchaOS comes with only the essential applications and libraries required for running common services and containers.
- **Secure by Design**\
EtchaOS integrates seamlessly with Secure Boot and doesn't expose unnecessary network services, minimizing attack vectors.
- **Flexible Variants**\
EtchaOS comes in [multiple variants]({{% ref "/etchaos/references/download" %}}) based on popular Linux distributions like AlmaLinux (`alm`), Debian (`deb`), Fedora (`fed`), and Ubuntu (`ubu`). Choose the variant that best suits your environment.

## Benefits
- **Consistent Experience**\
No matter where you deploy EtchaOS, you'll get the same consistent operating model. Whether you're using deb on-premise or ubu in the cloud, they both offer the same set of features and capabilities.
- **Container Friendly**\
EtchaOS includes industry-standard container tooling, making it ideal for running containers standalone or using Kubernetes.
- **Deployment Flexibility**\
Deploy EtchaOS on-premise, in the cloud, on your network, or even on ARM devices for maximum versatility.
- **Testing Powerhouse**\
EtchaOS is a perfect choice for CI/CD runners or test benches for validating OS, kernel, or library compatibility.
- **Seamless Upgrades**\
Effortlessly upgrade EtchaOS to the latest version with clear manifests that show you what packages have changed. You can even switch variants to rapidly reprovision to a different base OS.

{{% alert color="primary" title="Built For You" %}}
Let us build a customized EtchaOS for your app, or train and support your team. \
{{% contactus "Custom%20EtchaOS%20Configuration" %}} for consulting and support.
{{% /alert %}}
