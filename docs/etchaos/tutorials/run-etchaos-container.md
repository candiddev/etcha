---
description: How to run EtchaOS in a container for testing and experimenting.
title: Run EtchaOS in a Container
type: docs
---

An easy way to try out EtchaOS is using a container.  In this tutorial, we will download an [EtchaOS Variant]({{< ref "/etchaos/references/download" >}}) and run it  as a QEMU VM via a container.

## 0. Prerequisites

Before starting this, you should:

- Install [Docker](https://docker.com) or [Podman](https://podman.io).  This tutorial will use Docker but the commands _should be the same_ Podman.
- [Install Etcha]({{< ref "/docs/guides/install-etcha" >}})

## 1. Download a Variant

Head over to [Download]({{< ref "/etchaos/references/download" >}}) and download the **Kernel** and **Initrd** files for a Variant.  In this example we're using the `deb` variant, `amd64` arch:

```bash
mkdir etchaos
etcha copy change https://etcha.dev/releases/etchaos_deb_amd64.initrd etchaos/etchaos_deb_amd64.initrd
etcha copy change https://etcha.dev/releases/etchaos_deb_amd64.vmlinuz etchaos/etchaos_deb_amd64.vmlinuz
```

## 2. Create some keys

Etcha uses cryptographic keys to validate configurations.  We need to generate a keypair that we can use signing and verifying:

```bash
$ etcha gen-keys
New Password (empty string skips PBKDF): 
Confirm Password (empty string skips PBKDF): 
{
  "privateKey": "ed25519private:MC4CAQAwBQYDK2VwBCIEIA9o98iCtk+/TDRLt5/aKcoistxbKOo94G5alApDPRnE:jjKByT4bnU",
  "publicKey": "ed25519public:MCowBQYDK2VwAyEA5noJdcrHmhEO8mQA89kSdd4/GQZbDz0kbxoPKNApTkc=:jjKByT4bnU"
}
```

Lets export the `privateKey` for future use as the {{% config build_signingKey %}}:

```bash
export etcha_build_signingKey=ed25519private:MC4CAQAwBQYDK2VwBCIEIA9o98iCtk+/TDRLt5/aKcoistxbKOo94G5alApDPRnE:jjKByT4bnU
```

## 2. Run the Container

Next, we'll run the [`qemu` container](https://candid.dev/containers/qemu).  This container wraps QEMU with some convenience functions to make it work nicely in a container.  This will start EtchaOS as VM and allow us to [configure it]({{< ref "/etchaos/references/config-discovery" >}}).

Replace `ARCH` with the arch you chose and `verifyKeys` if you want to use your own.

```bash
docker run -d $(if [[ -e /dev/kvm ]]; then echo "--device /dev/kvm --user root"; fi) -e ARCH=amd64 -e QEMUARGS="-initrd $(pwd)/etchaos/etchaos_deb_amd64.initrd -kernel $(pwd)/etchaos/etchaos_deb_amd64.vmlinuz -nic user,hostfwd=tcp::4000-:4000 -nographic -nodefaults" -e SMBIOS='etchaos={run:{verifyKeys:["ed25519public:MCowBQYDK2VwAyEA5noJdcrHmhEO8mQA89kSdd4/GQZbDz0kbxoPKNApTkc=:jjKByT4bnU"]},sources:{test:{allowPush:true,shell:{command:"machinectl shell root@"}}}}' -p 4000:4000 -v $(pwd)/etchaos:$(pwd)/etchaos --cap-add NET_ADMIN --name etchaos_test ghcr.io/candiddev/qemu
```

This command does a few things:
1. Adds the `/dev/kvm` device to the container if it exists.  This will greatly improve performance if the host and variant architecture match.
2. Specifies a path for the initrd and kernel files we downloaded earlier.
3. Forwards port 4000 from the VM, through QEMU, to the host.  This is the port Etcha listens on by default.
4. Injects [smbios]({{< ref "/etchaos/references/config-discovery#smbios" >}}) configuration data:
   1. Adds the public key from above to EtchaOS.
   2. Adds a {{% config sources %}} named `test`:
      1. Push is allowed
      2. Shell is allowed and configured to run `machinectl shell root@`

The logs should look like this:

```bash
$ docker logs -f etchaos_test
Starting web listener...
Starting TPM...
Starting VM...
```

## 3. Interact with EtchaOS

Lets try getting a shell to the instance:

```bash
$ etcha shell test localhost
Connected to the local host. Press ^] three times within 1s to exit session.
root@localhost:~# 
```

We now have a shell on the local Etcha instance.  Feel free to poke around on it.

Exit out of the shell and push an example command:

```bash
 $ etcha push -h localhost test ls /etc/apt
localhost:
    apt.conf.d
    auth.conf.d
    keyrings
    preferences.d
    sources.list
    sources.list.d
    trusted.gpg.d
```

Continue with the [Etcha Guides]({{< ref "/docs/guides" >}}) for more details on how to use Etcha.
