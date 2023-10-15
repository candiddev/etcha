---
categories:
- reference
description: Reference documentation for Etcha JWTs
title: JWT
---

Etcha uses JSON Web Tokens (JWTs) to package and sign Patterns and other metadata into configurations that can be pushed or pulled from Etcha instances.

## Format

A JWT is three base64 raw URL encoded strings joined with periods.  It looks like this:

`eyJhbGciOiJFZERTQSIsImtpZCI6IlNOeGloVTFpQ2wiLCJ0eXAiOiJKV1QifQ.eyJMaWNlbnNlZCI6dHJ1ZSwibmFtZSI6IiIsIkxhc3RQdXJjaGFzZSI6IjAwMDEtMDEtMDFUMDA6MDA6MDBaIiwiYXVkIjoiYXVkaWVuY2UiLCJleHAiOjE2OTcwNzA2NTQsImp0aSI6ImlkIiwiaWF0IjoxNjk3MDcwNjQ0LCJpc3MiOiJpc3N1ZXIiLCJuYmYiOjE2OTcwNzA2NDQsInN1YiI6InN1YmplY3QifQ.u1tbySaWSZWvQmB0MY52Zh9YD7OcGPk7kTi6aGTsln7L-aSWH65iIliMuSls4DUTN-_bQDqgXSmj2nAPVZ0aCA`

The three parts are:

- [**Header**](#header), an object containing properties describing how the JWT was signed.
- [**Payload**](#payload), an object containing standard and custom properties.
- [**Signature**](#signature), a digital signature of the first two parts that can be cryptographically verified.

### Header

The header part of a JWT, also called a Javascript Object Signing and Encryption (JOSE) header, is typically made up of two properties:

- `alg`: specifies the cryptographic algorithm
- `typ`: specifies the type of token, for a JWT this will be `JWT`.

An additional property, `kid`, may be specified to provide a hint for the cryptographic key ID used to sign the token.

An example header looks like this:

```json
{
  "alg": "EdDSA",
  "typ": "JWT"
}
```

This JSON string is base64 raw URL encoded (no padding) into a header string:

`eyJhbGciOiJFZERTQSIsImtpZCI6InI5dzlOSXdoSlYiLCJ0eXAiOiJKV1QifQ.eyJldGNoYUJ1aWxkTWFuaWZlc3QiOiJidWlsZCIsImV0Y2hhUGF0dGVybiI6eyJlbnRyeXBvaW50IjoiL21haW4uanNvbm5ldCIsImZpbGVzIjp7Ii9tYWluLmpzb25uZXQiOiJ7XG5cdFx0XHRcdFx0cnVuOiBbXG5cdFx0XHRcdFx0XHRpZDogXCIxXCJcblx0XHRcdFx0XHRdXG5cdFx0XHRcdH0ifX0sImV0Y2hhUnVuRW52Ijp7ImhlbGxvIjoid29ybGQiLCJ3b3JsZCI6ImhlbGxvIn0sImV0Y2hhVmVyc2lvbiI6InYyMDIzLjEwLjAyIiwiYXVkIjoiYXVkaWVuY2UhIiwiZXhwIjoxNjk3MDcxMzIxLCJpYXQiOjE2OTcwNzEyNjIsImlzcyI6Imlzc3VlciEiLCJuYmYiOjE2OTcwNzEyNjIsInN1YiI6InN1YmplY3QhIn0.YD3rvh4BMEJUoBc0n7f4HW3moc-gV4mVaVrsgc1cHzKDEgSbVuAEj_ALEV3HSSUGx928gwrPk4AzmX1D4nibCQ`

### Payload

The payload part of a JWT is an object containing both standard ("Registered") and custom properties, called claims.

An example payload looks like this:

```json
{
  "etchaBuildManifest": "build",
  "etchaPattern": {
    "entrypoint": "/main.jsonnet",
    "files": {
      "/main.jsonnet": "{\n\t\t\t\t\trun: [\n\t\t\t\t\t\tid: \"1\"\n\t\t\t\t\t]\n\t\t\t\t}"
    }
  },
  "etchaRunEnv": {
    "hello": "world",
    "world": "hello"
  },
  "etchaVersion": "v2023.10.02",
  "aud": "audience!",
  "exp": 1697071321,
  "iat": 1697071262,
  "iss": "issuer!",
  "nbf": 1697071262,
  "sub": "subject!"
}
```

This JSON string is base64 raw URL encoded (no padding) into a header string:

`eyJMaWNlbnNlZCI6dHJ1ZSwibmFtZSI6IiIsIkxhc3RQdXJjaGFzZSI6IjAwMDEtMDEtMDFUMDA6MDA6MDBaIiwiYXVkIjoiYXVkaWVuY2UiLCJleHAiOjE2OTcwNzA2NTQsImp0aSI6ImlkIiwiaWF0IjoxNjk3MDcwNjQ0LCJpc3MiOiJpc3N1ZXIiLCJuYmYiOjE2OTcwNzA2NDQsInN1YiI6InN1YmplY3QifQ`

#### Registered JWT Claims

These are "standard" JWT claims present in most JWTs.

##### `aud`

List of strings, typically used to indicate who the JWT was meant for.  This is populated by the Pattern property [`audience`](../patterns#audience).

##### `exp`

Integer, a unix timestamp indicating when the JWT expires.  This is populated by the Pattern property [`expiresInSec`](../patterns#expiresinsec).

##### `jti`

String, typically used to indicate a unique ID for the JWT.  This is populated by the Pattern property [`id`](../patterns#id).

##### `iat`

Integer, a unix timestamp of when the JWT was created.  This is populated automatically during Pattern build and cannot be set.

#### `iss`

String, typically used to indicate a domain or organization that issued the JWT.  This is populated by the Pattern property [`issuer`](../patterns#issuer).

##### `nbf`

Integer, a unix timestamp of when the JWT becomes valid.  This is populated automatically during Pattern build to the same value as `iat`.  Future Etcha releases may make this configurable.

##### `sub`

String, typically used to indicate the subject or user for the JWT.  

String, typically used to indicate a domain or organization that issued the JWT.  This is populated by the Pattern property [`issuer`](../patterns#issuer).

#### Etcha Claims

Etcha adds the following claims to JWTs:

##### `etchaBuildManifest`

String, for [Commands](../commands) that produce the event `build_manifest`, the output of executing [`change`](../commands#change) is concated into this property.  This property is primarily used for cache busting/forcing a JWT to be downloaded during a run diff, but it could contain useful data around secure supply chain or patch versions.

##### `etchaPattern`

Object, containing the main [Pattern](../patterns) file and any related imports.  During build, Etcha will read in the Pattern path specified and gather all the imports.  It will set the `entrypoint` property to the main Pattern file, and include all files and imports under `files` as key, values:

```json
{
  "entrypoint": "main.jsonnet",
  "files": {
    "/func.libsonnet": "function(word)\n{check: \"echo %s\" % word, id: \"hello %s\" % word}",
    "/main.jsonnet": "local f = import \"./func.libsonnet\";\n{run: [f(\"world\")]}"
  }
}
```

When this JWT is pulled or pushed, Etcha will [render the Pattern](../patterns#rendering) specified in this object.  The above Pattern would look like this after rendering:

```json
{
  "run": [
    {
      "check": "echo world",
      "id": "hello world"
    }
  ]
}
```

##### `etchaRunEnv`

A map of string keys and values containing environment variables.  These environment variables will be added to the existing `runEnv` keys in [Patterns](../patterns#runenv) with the prefix `etcha_run_`, i.e. `{"key":"value"}` becomes `etcha_run_key=value`.  All values will be added to the [Commands Environment Variables](../commands#environment-variables) during a Pattern run.


##### `etchaVersion`

String, the Etcha version used to build/sign the JWT.  This can be ignored during diffs, otherwise a change in Etcha versioning will trigger a diff cycle.

### Signature

The final part of the JWT is the signature.  Given a string containing the base64 header and base64 payload, joined with a ".", a signature is generated using a private key or a HMAC and base64 raw URL encoded.  See [Cryptography](../cryptography) for more information on generating keys.

## Verification

Etcha verifies the JWTs in a few ways:

- It checks the signature using [`verifyKeys`](../config#verifykeys).  If a verify key cannot be found that matches the JWT, it does not accept the JWT (during either push or pull).
- It validates the [`exp`](#exp) and [`nbf`](#nbf) values to ensure the token has valid timestamps.
- 
This verification process ensures the JWT can be trusted.

## Custom Sign/Verify

Organizations wishing to provide more verification/signing capabilities can leverage custom verify and sign commands ([verifyCommands](../config#verifycommands) and [signCommands](../config#signingcommands)).  See [Building Patterns](../../guides/building-patterns) for more information.
