---
categories:
- reference
description: Reference documentation for Etcha's cryptography
title: Cryptography
---

Etcha uses cryptographic signatures for signing [JWTs]({{< ref "/docs/references/jwt" >}}).  The cryptography for these signatures is provided using a cryptography wrapper to allow future proofing and swapping of cryptographic algorithms.

## Generating Keys

Etcha can generate well formatted, best choice cryptographic keys using {{% cli gen-keys %}}.

## Verify Keys

Verify keys are used for verifying [JWTs]({{< ref "/docs/references/jwt" >}}).  In Etcha, these are either RSA, EC, or Ed25519 public keys, in PKIX DER form, standard base64 encoded, with the algorithm type prepended at the beginning and an optional key ID at the end.

An example public key Etcha understands is `ed25519public:MCowBQYDK2VwAyEAwsXVnwaquPnF1J3oqhB1qFrBEAW+2FDkGYv7iqoPVHs=:lAASVe7woP`.

Etcha understands these public key algorithm types:

- `ed25519public`
- `ecp256public`
- `rsa2048public`

## Sign Keys

Sign keys are used for signing [JWTs]({{< ref "/docs/references/jwt" >}}).  In Etcha, these are either RSA, EC, or Ed25519 private keys, in PKCS #8 DER form, standard base64 encoded, with the algorithm type prepended at the beginning and an optional key ID at the end.

An example private key Etcha understands is `ed25519private:MC4CAQAwBQYDK2VwBCIEIH0r30uYYQVEFJJ7cG5fPLteuGUPb8qBH+vAOjZnnNGJ:lAASVe7woP`.

Etcha understands these private key algorithm types:

- `ed25519private`
- `ecp256private`
- `rsa2048private`
