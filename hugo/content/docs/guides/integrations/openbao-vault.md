---
categories:
- guide
description: How to use Etcha with OpenBao/Vault
title: OpenBao/Vault
---

## Overview

[OpenBao](https://github.com/openbao/openbao) and [Vault](https://www.vaultproject.io/) are platforms for managing secrets like passwords and cryptographic keys.  Etcha can leverage their cryptographic functions to securely sign and verify [Patterns]({{< ref "/docs/references/patterns" >}}) to limit access to private keys via [`signingCommands`]({{< ref "/docs/guides/building-patterns#signingcommands" >}}) and [`verifyCommands`]({{< ref "/docs/guides/running-patterns#verifycommands" >}}).

These examples use OpenBao, but it should apply to Vault, too.

## OpenBao Setup

Etcha uses OpenBao's [Transit Secrets Engine](https://developer.hashicorp.com/vault/docs/secrets/transit) to secure sign and verify JWTs.

Here is a guide to setup the Secrets engine:

```bash
# Enable the Transit secrets engine
openbao secrets enable transit

# Create an Ed25519 key named etchaos
openbao write transit/keys/etchaos type=ed25519
```

## Etcha Configuration

### Signing

We'll configure Etcha to sign Patterns using OpenBao via [`signingCommands`]({{< ref "/docs/references/config#signingcommands" >}}).  Here is an example configuration file:

**config.jsonnet**
```
{
  build: {
    signingCommands: [
      {
        id: 'openbao',
        always: true,
        change: |||
          header='{"alg":"EdDSA","typ":"JWT"}'
          token="$(basenc --base64url -w0 <<<${header} | cut -d= -f1 | tr -d '\n').${ETCHA_PAYLOAD}"
          sig=$(openbao write -format=json transit/sign/etchaos input=$(echo -n ${token} | base64 -w0 | tr -d '\n') marshaling_algorithm=jws)
          printf '%s.%s' ${token} $(jq -r .data.signature <<<${sig} | cut -d: -f3)
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
- Using OpenBao, it signs the `token` using the key we created above and stores it in `sig`
- Combines the `sig` and `token` into a JWT and prints it to stdout
- Using the [`jwt` event]({{< ref "/docs/references/events#jwt" >}}), Etcha will take the output of this as the JWT for the build

Continue following the [Building Patterns guide]({{< ref "/docs/guides/building-patterns" >}}).

### Verifying

Etcha can verify OpenBao signatures [online](#online) using `verifyCommands` to communicate directly to OpenBao, or [offline](#offline) using an exported public key from OpenBao.

#### Online

We'll configure Etcha to verify Patterns using OpenBao via [`verifyCommands`]({{< ref "/docs/references/config#verifycommands" >}}).  Here is an example configuration file:

**config.jsonnet**
```
{
  run: {
    verifyCommands: [
      {
        id: 'openbao',
        always: true,
        change: |||
          header=$(cut -d. -f1 <<<${ETCHA_JWT})
          payload=$(cut -d. -f2 <<<${ETCHA_JWT})
          sig=$(cut -d. -f3 <<<${ETCHA_JWT})
          if [[ $(openbao write -field=valid transit/verify/etchaoss input=$(echo -n "${header}.${payload}" | base64 -w0 | tr -d '\n') marshaling_algorithm=jws signature="vault:v1:${sig}") == true ]]; then
            printf '%s' ${ETCHA_JWT}
          fi
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

- Constructs a [JWT]({{< ref "/docs/references/jwt" >}}) header, payload, and signature using the [`ETCHA_JWT`]({{< ref "/docs/references/commands#etcha_jwt" >}}) variable
- Using OpenBao, it verifies the `token` using the key and the `sig` we created above
- **If it verifies successfully**, print the JWT
- Using the [`jwt` event]({{< ref "/docs/references/events#jwt" >}}), Etcha will take the output of this as the JWT and assume it was verified successfully.  If we didn't output a JWT or exited non-zero, Etcha will reject the JWT.

Continue following the [Running Patterns guide]({{< ref "/docs/guides/running-patterns" >}}).

#### Offline

We can use the public key returned from `openbao read -keys transit/export/public-key/etchaos`:

```bash
$ openbao read -field keys transit/export/public-key/etchaos
map[1:gBykng9f71hnl54iBxadY6mUTEU058EhFJyT3C3RIjE=]
```

We can add the key to [`verifyKeys`]({{< ref "/docs/references/config#verifykeys" >}}) by adding the prefix `ed25519public:`

```
{
  run: {
    verifyKeys: [
      'ed25519public:gBykng9f71hnl54iBxadY6mUTEU058EhFJyT3C3RIjE=:vault',
    ],
  },
}
```

Continue following the [Running Patterns guide]({{< ref "/docs/guides/running-patterns" >}}).
