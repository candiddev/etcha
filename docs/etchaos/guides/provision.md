---
categories:
- guide
description: How to provision EtchaOS on various platforms.
title: Provision
type: docs
weight: 10
---

EtchaOS can be ran on a variety of platforms.  To get started, review the [EtchaOS Variants]({{< ref "/etchaos/references/download" >}}) and determine which Variant is best for your needs.  Then, choose a platform to provision EtchaOS on.

EtchaOS currently supports these platforms:

- [Bare Metal](#baremetal)\
Physical servers and devices
- [Containers](#qemu)\
Virtual machine running on QEMU via a container on Docker, Kubernetes, or Podman
- [Libvirt](#qemu)\
Virtual machine running on Libvirt
- [PXE](#pxe)\
Run EtchaOS via network boot, including support for Secure Boot
- [Proxmox](#qemu)\
Virtual machine running on Proxmox
- [QEMU](#qemu)\
Virtual machine running on QEMU
- [VMware](#vmware)\
Virtual machine running on VMware

The following platforms are considered experimental, please {{% contactus "EtchaOS%20Platform" %}} for assistance:

- **Amazon Web Services (AWS)**\
Run EtchaOS as an EC2 instance on AWS
- **Google Cloud Platform (GCP)**\
Run EtchaOS as a Compute Engine instance on GCP
- **Microsoft Azure**\
Run EtchaOS as a VM on Azure

{{% alert color="primary" title="Need Help?" %}}
We can provision EtchaOS in your cloud or datacenter. \
{{% contactus "Custom%20EtchaOS%20Configuration" %}} for consulting and support.
{{% /alert %}}

## Bare Metal

EtchaOS should work out of the box with most bare metal platforms using either BIOS or UEFI boot.  Secure boot should work as well without any extra steps or key signing.  EtchaOS comes bundled with most commonly used firmware, but certain firmware such as GPU drivers may need to be installed separately.

To get started with EtchaOS on Bare Metal:

1. [Download the Disk]({{% ref "/etchaos/references/download#download" %}}) for your Variant.
2. Flash a USB drive with the disk image:
```bash
gzip -d etchaos_deb_amd64.raw.gz | dd of=/dev/sda
```
3. Determine your [Config]({{% ref "/etchaos/references/config-discovery" %}}) strategy for Etcha.  Using [cidata]({{% ref "/etchaos/references/config-discovery#cidata" %}}) is recommended.

### ARM Support

ARM can be tricky with EtchaOS.  While EtchaOS has been tested with various [Tow-Boot](https://tow-boot.org/) platforms, it should work with [U-Boot](https://docs.u-boot.org/en/latest/) configured for with BIOS or UEFI support.

Sharing a disk with U-Boot or Tow-Boot bootloaders can be problematic.  The EtchaOS disk image is not meant to be combined with other disk images, and splicing it onto an existing disk image can be error prone.  There are a few ways to accomplish this:

- Use two disks, one for the bootloader (U-Boot, Tow-Boot), one for EtchaOS
- Using the U-Boot or Tow-Boot disk image, add the Kernel and Initrd files from [Downloads]({{% ref "/etchaos/references/download" %}}) and configure a [bootflow](https://docs.u-boot.org/en/stable/develop/bootstd.html)

## QEMU

EtchaOS has been tested and deployed across various QEMU platforms and boot sources.

To get started with EtchaOS on QEMU:

1. Choose your [Platform](#qemu-platforms)
2. Choose your [Boot Source](#qemu-source)
3. Determine your [Config]({{% ref "/etchaos/references/config-discovery" %}}) strategy for Etcha.  Using [SMBIOS]({{% ref "/etchaos/references/config-discovery#smbios" %}}) or [cidata]({{% ref "/etchaos/references/config-discovery#cidata" %}}) is recommended.

### Platforms {#qemu-platforms}

QEMU can be used on various platforms:

#### Container

> Checkout the [Run EtchaOS in a Container Tutorial]({{% ref "/etchaos/tutorials/run-etchaos-container" %}})

Candid Development has created a [QEMU container](https://candid.dev/containers/qemu) that can easily launch an EtchaOS instance using QEMU.  For this to work, mount the location of the disk image or kernel and initrd files into the container and follow the instructions on that page.  The QEMU arguments listed below can be used with the environment variable, `QEMUARGS`.

#### Libvirt

A Libvirt VM can be created using the disk image or kernel and initrd files below.  Simply download the files to a Libvirt directory pool, such as `/var/lib/libvirt/images`.

#### Proxmox

A Proxmox VM can be created by [importing a raw disk](https://pve.proxmox.com/wiki/Migrate_to_Proxmox_VE#Import_Disk) from the disk image below.  Make sure to decompress it first.

### Boot Source {#qemu-source}

QEMU VMs can boot from either a raw disk image or the kernel and initrd files.

#### Disk Image

1. [Download the Disk]({{% ref "/etchaos/references/download#download" %}}) for your Variant.  Decompress the `.raw.gz` file using `gzip -d`.
2. Determine your [Config]({{% ref "/etchaos/references/config-discovery" %}}) strategy for Etcha.  Using [smbios]({{% ref "/etchaos/references/config-discovery#smbios" %}}) is an easy way to get started.
3. Launch a QEMU virtual machine using the disk:
```bash
-drive file=etchaos_deb_amd64.raw,media=disk
```

#### Kernel and Initrd

1. [Download the Kernel and Initrd]({{% ref "/etchaos/references/download#download" %}}) for your Variant.
2. Determine your [Config]({{% ref "/etchaos/references/config-discovery" %}}) strategy for Etcha.  Using [smbios]({{% ref "/etchaos/references/config-discovery#smbios" %}}) is an easy way to get started.
3. Launch a QEMU virtual machine using the kernel and initrd:
```bash
-kernel etchaos_deb_amd64.vmlinuz -initrd etchaos_deb_amd64.initrd
```

## PXE

EtchaOS can be deployed on the network using UEFI PXE Boot.

To get started with EtchaOS using PXE:

1. [Download the Kernel and Initrd]({{% ref "/etchaos/references/download#download" %}}) for your Variant.  Put these files in the root of your TFTP directory.
2. [Download the PXE archive]({{% ref "/etchaos/references/download#download" %}}) for your Variant.  Extract the contents of this archive to the root of your TFTP directory.
3. Configure your boot server to present the bootx64.efi file (or bootaa64.efi for arm64 hosts).
4. Determine your [Config]({{% ref "/etchaos/references/config-discovery" %}}) strategy for Etcha.  PXE booted hosts may or may not have disks, so consider setting up an [Instance User Data]({{% ref "/etchaos/references/config-discovery#user-data" %}}) service on your network.

## VMware

VMware ESXi VMs can be created from EtchaOS by converting the raw disk to a VMDK using `qemu-img`:

To get started with EtchaOS on VMware:

1. [Download the Disk]({{% ref "/etchaos/references/download#download" %}}) for your Variant.  Decompress the `.raw.gz` file using `gzip -d`.
2. Convert the disk to a VMDK:
```bash
qemu-img convert -O vmdk etchaos_deb_amd64.raw etchaos_deb_amd64.vmdk
```
3. Import the VMDK into your VMware Datastore.  The VM should support BIOS or UEFI boot.
3. Determine your [Config]({{% ref "/etchaos/references/config-discovery" %}}) strategy for Etcha.  Using [cidata]({{% ref "/etchaos/references/config-discovery#cidata" %}}) is recommended.
