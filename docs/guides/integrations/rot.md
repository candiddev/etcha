---
categories:
- guide
description: How to use Etcha with Rot
title: Rot
---

## Overview

[Rot](https://rotx.dev) is a platform for managing secrets like passwords and cryptographic keys.  Etcha can leverage Rot's cryptographic functions to securely sign and verify [Patterns]({{< ref "/docs/references/patterns" >}}) to limit access to private keys via [`signingCommands`]({{< ref "/docs/guides/building-patterns#signingcommands" >}}) and [`verifyCommands`]({{< ref "/docs/guides/running-patterns#verifycommands" >}}).

{{% alert title="Candid Commentary" color="info" %}}
Rot is a Candid Development product and was originally designed as a simple keystore for early versions of Etcha.
{{% /alert %}}


## Rot Install

Etcha can install the latest version of Rot from its [libraries]({{< ref "/docs/references/libraries#rotinstall" >}}):

```bash
$ etcha local change "(import 'lib/etcha/rotInstall.libsonnet')(dst='/tmp/rot')"
INFO  Changed download Rot to /tmp/rot [1s]
INFO  Always changed rot version [50ms]
```

## Rot Setup

To use Etcha with Rot, we need to initialize Rot and add a private key for signing.  We left the password blank for the keys, but you should use passwords or a secure enclave in production.

```bash
$ rot init
New Password (empty string skips PBKDF): 
Confirm Password (empty string skips PBKDF): 
```

## Etcha Configuration

### Signing via Rot Private Key (simpler, private key leaves Rot)

We can have Rot inject a private key for Etcha to sign Patterns by creating a private key with the name {{% config build_signingKey etcha_build_signingKey %}}:

```bash
$ rot val-add-prv rot/etcha_build_signingKey
$ rot val-dis -m publicKey rot/etcha_build_signingKey
ed25519public:MCowBQYDK2VwAyEAJb9C2YlvNY2liircHRc4sresVmxCkPekzrHFTGQfHWg=:etcha_build_signingKey
```

The public key returned by `rot val-ids` is the {{% config run_verifyKey %}} we will provide to Etcha to trust the Patterns when we run them.

In order to use this signing key, we need to wrap Etcha commands with Rot:

```bash
$ rot run rot ...
```

Continue following the [Running Patterns guide]({{< ref "/docs/guides/running-patterns" >}}), wrapping the commands with `rot run`.

### Signing via Rot (more complex, private key stays in Rot)

We'll configure Etcha to sign Patterns using Rot via {{% config build_signingCommands %}}.

Lets generate a private key:
```bash
$ rot val-add-prv rot/etcha
$ rot val-dis -m publicKey rot/etcha
ed25519public:MCowBQYDK2VwAyEAJb9C2YlvNY2liircHRc4sresVmxCkPekzrHFTGQfHWg=:etcha
```

The public key returned by `rot val-dis` is the {{% config run_verifyKeys %}} we will provide to Etcha to trust the Patterns when we run them.

Here is an example configuration file containing the `signingCommands` for Rot:

**config.jsonnet**
```
{
  build: {
    signingCommands: [
      {
        id: 'rot',
        always: true,
        change: |||
          header='{"alg":"EdDSA","typ":"JWT"}'
          token="$(rot base64 -r -u <<<${header}).${ETCHA_PAYLOAD}"
          printf '%s.%s' ${token} $(rot sig-new rot/etcha ${token})
        |||,
        onChange: [
          'etcha:jwt',
        ]
      }
    ]
  }
}
```

This configuration does a few things:

- Constructs a [JWT header]({{< ref "/docs/references/jwt" >}}) using JSON and stores it in the variable `header`
- Base64 encrypts the header and combines it with the [`ETCHA_PAYLOAD`]({{< ref "/docs/references/commands#etcha_payload" >}}) variable into a `token`
- Using Rot, it signs the `token` using the key we created above and stores it in `sig`
- Combines the `sig` and `token` into a JWT and prints it to stdout
- Using the [`jwt` event]({{< ref "/docs/references/events#jwt" >}}), Etcha will take the output of this as the JWT for the build

Continue following the [Building Patterns guide]({{< ref "/docs/guides/building-patterns" >}}).

### Verifying

Etcha can verify Rot signatures [online](#online) using `verifyCommands` to communicate directly to Rot, or [offline](#offline) using an exported certificate from Rot.

#### Offline

We can use the public key returned from `rot val-dis -m publicKey rot/etcha`:

```bash
$ rot val-dis -m publicKey rot/etcha
ed25519public:MCowBQYDK2VwAyEAJb9C2YlvNY2liircHRc4sresVmxCkPekzrHFTGQfHWg=:etcha
```

We can add the key to {{% config run_verifyKeys %}}:

```
{
  run: {
    verifyKeys: [
      'ed25519public:MCowBQYDK2VwAyEAJb9C2YlvNY2liircHRc4sresVmxCkPekzrHFTGQfHWg=:etcha',
    ],
  },
}
```

Continue following the [Running Patterns guide]({{< ref "/docs/guides/running-patterns" >}}).
