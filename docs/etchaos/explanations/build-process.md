---
description: An overview of how EtchaOS is built.
title: Build Process
type: docs
---

EtchaOS is built every week from various distributions using Etcha on GitHub Runners.  The build process is driven by an Etcha Pattern--this runs the build process and outputs the [JWT]({{< ref "/docs/references/jwt" >}}) used for managing updates and manifest details.

Here is an overview of how the build process works:

## 1. Initialization

Etcha creates a container using the official image of the source distribution (i.e. `docker.io/debian`).  This container has various packages added to it to facilitate bootstrapping, like `debootstrap` and disk utilities.

## 2. Bootstrap

Etcha executes into the container and runs the bootstrap process.  This varies amongst distributions, but the goal is to bootstrap a minimum viable OS for the distribution.

## 3. Refinement

Etcha switches over to a standard refinement script that installs and reconfigures the OS to become EtchaOS.  During this step, tools like Etcha, Docker, and Rot are added, systemd units are enabled/disabled, files are cleaned up, and archives are made to get the OS ready for packaging.

## 4. Packaging

Etcha creates an initrd image containing scripts, directories, and files used to setup the EtchaOS filesystem.  Etcha will also copy the Linux kernel vmlinuz file.  Finally, Etcha creates a raw disk image containing the kernel, image, and a bootloader.  After this step, Etcha generates a JWT containing a manifest of the packages, their versions, and any other data that would be useful to understand the Variant.

## 5. Testing

Once the build is packaged, the artifacts are rigorously tested to ensure the build works.  The raw disk image is booted using QEMU in UEFI Secure Boot mode and a Pattern is pushed to validate the build works and contains the correct binaries.  Then, Etcha deletes the kernel and initrd file from the raw disk and performs an "auto update" to reinstall them.  Etcha then restarts the disk image, waits for it to reboot, and validates the image still boots successfully.  For amd64 variants, Etcha will also test that BIOS boot works, too.

## 6. Release

With the build packaged and tested, Etcha performs a comparison against the built JWT and the existing JWT for the Variant.  If they match, the build is not released and the process ends.  If they do not match, Etcha pushes the new version of the build to our CDNs.  Existing EtchaOS servers will periodically retrieve this JWT file and compare it against their existing build, if it doesn't match, a new EtchaOS version will be installed.
