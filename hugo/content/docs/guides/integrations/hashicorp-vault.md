---
categories:
- guide
description: How to integrate Etcha and HashiCorp Vault
title: HashiCorp Vault
---

## Overview

[HashiCorp Vault](https://www.vaultproject.io/) is a platform for managing secrets like passwords and cryptographic keys.  Etcha can leverage Vault's cryptographic functions to securely sign and verify [Patterns](../../../explanations/patterns) to limit access to private keys via [`signingCommands`](../../building-patterns#signingcommands) and [`verifyCommands`](../../running-patterns#verifycommands).

## Vault Setup

Etcha uses Vault's [Transit Secrets Engine](https://developer.hashicorp.com/vault/docs/secrets/transit) to secure sign and verify certificates.

Here is a guide to setup the Secrets engine:

```bash
# Enable the Transit secrets engine
vault secrets enable transit

# Create an Ed25519 key named etchaos
vault write transit/keys/etchaos type=ed25519
```

## Etcha Configuration

### Signing

We'll configure Etcha to sign Patterns using Vault via [`signingCommands`](../../../references/config#signingcommands).  Here is an example configuration file:

**config.jsonnet**
```
{
  build: {
    signingCommands: [
      {
        id: 'vault',
        always: true,
        change: |||
          header='{"alg":"EdDSA","typ":"JWT"}'
          token="$(basenc --base64url -w0 <<<${header} | cut -d= -f1 | tr -d '\n').${ETCHA_PAYLOAD}"
          sig=$($(pwd)/.bin/vault write -format=json transit/sign/etchaos input=$(echo -n ${token} | base64 -w0 | tr -d '\n') marshaling_algorithm=jws)
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

- Constructs a [JWT header](../../../references/jwt) using JSON and stores it in the variable `header`
- Base64 encrypts the header and combines it with the [`ETCHA_PAYLOAD`](../../../references/commands#etcha_payload) variable into a `token`
- Using Vault, it signs the `token` using the key we created above and stores it in `sig`
- Combines the `sig` and `token` into a JWT and prints it to stdout
- Using the [`jwt` event](../../../references/events#jwt), Etcha will take the output of this as the JWT for the build

Continue following the [Building Patterns guide](../../building-patterns).

### Verifying

Etcha can verify Vault signatures [online](#online) using `verifyCommands` to communicate directly to Vault, or [offline](#offline) using an exported certificate from Vault.

#### Online

We'll configure Etcha to verify Patterns using Vault via [`verifyCommands`](../../../references/config#verifycommands).  Here is an example configuration file:

**config.jsonnet**
```
{
  run: {
    verifyCommands: [
      {
        id: 'vault',
        always: true,
        change: |||
          header=$(cut -d. -f1 <<<${ETCHA_JWT})
          payload=$(cut -d. -f2 <<<${ETCHA_JWT})
          sig=$(cut -d. -f3 <<<${ETCHA_JWT})
          if [[ $($(pwd)/.bin/vault write -field=valid transit/verify/etchaoss input=$(echo -n "${header}.${payload}" | base64 -w0 | tr -d '\n') marshaling_algorithm=jws signature="vault:v1:${sig}") == true ]]; then
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

- Constructs a [JWT](../../../references/jwt) header, payload, and signature using the [`ETCHA_JWT`](../../../references/commands#etcha_jwt) variable
- Using Vault, it verifies the `token` using the key and the `sig` we created above
- **If it verifies successfully**, print the JWT
- Using the [`jwt` event](../../../references/events#jwt), Etcha will take the output of this as the JWT and assume it was verified successfully.  If we didn't output a JWT or exited non-zero, Etcha will reject the JWT.

Continue following the [Running Patterns guide](../../running-patterns).

#### Offline

We can use the certificate returned from `vault read -keys transit/export/public-key/etchaos`:

```bash
$ vault read -field keys transit/export/public-key/etchaos
map[1:gBykng9f71hnl54iBxadY6mUTEU058EhFJyT3C3RIjE=]
```

We can add the key to [`verifyKeys`](../../../references/config#verifykeys) by adding the prefix `ed25519public:`

```
{
  run: {
    verifyKeys: [
      'ed25519public:gBykng9f71hnl54iBxadY6mUTEU058EhFJyT3C3RIjE=',
    ],
  },
}
```

Continue following the [Running Patterns guide](../../running-patterns).
