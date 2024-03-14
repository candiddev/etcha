---
categories:
- guide
description: How to build and sign Patterns in Etcha.
title: Building Patterns
weight: 50
---

In this guide, we'll go over building and signing Patterns in Etcha.

## Build Process

Etcha builds Patterns by executing the `build` Command list on the local instance, and then generating a [JWT]({{< ref "/docs/references/jwt" >}}) containing the Pattern and other metadata.

The JWT is then cryptographically signed to prevent modification and prove to other Etcha instances that it was created by a trusted authority.

The output of the the build process is this JWT, a cryptographically secure file containing base64 encoded text.

## Signing Patterns

We need a way to sign Patterns before building them.  Out of the box, Etcha cannot build or sign Patterns without [cryptographic keys]({{< ref "/docs/references/cryptography" >}}).  Specifically, we need to provide values in our Etcha configuration for [`signingKey`]({{< ref "/docs/references/config#signingkey" >}}) or [`signingCommands`]({{< ref "/docs/references/config#signingcommands" >}}).

### `signingKey`

A `signingKey` will use a static, private key to sign the Pattern JWT.  This value can be hardcoded in your "builder", or you can retrieve it from environment variables, a remote URL, a DNS record, etc.  This text is like a password though and should be protected.

We can generate keys appropriate for signing using [`etcha generate-keys sign-verify`]({{< ref "/docs/references/cli#generate-keys" >}}).  We'll use the `privateKey` value during build for the `signingKey`, and then we'll save the `publicKey` and use that for the [`verifyKeys`]({{< ref "/docs/references/config#verifykeys" >}}) when we push/pull.

Here is an example Etcha configuration showing a static `signingKey`:

```json
{
  "build": {
    "signingKey": "ed25519private:MC4CAQAwBQYDK2VwBCIEIDZUP0nKuVhDJu5h0QKAQVrZBLrDa9CA09SPJKj/21eG:qsX98cmrLW".
  }
}
```

We can also leverage some of the [dynamic Jsonnet functions]({{< ref "/docs/references/jsonnet#native-functions" >}}) to pull the key from a separate tool, like HashiCorp Vault:

```
local getEnv(key) = std.native('getEnv')(key);
local getPath(path, fallback=null) = std.native('getPath')(path, fallback);

{
  build: {
    signingKey: std.parseJson(getPath('https://vault.mycompany.com/v1/kv/secret/etcha#x-vault-token:%s' % getEnv('VAULT_TOKEN'))).data.private_key,
  },
}
```

This example looks a little intense, lets walk through it:

- We declared two functions at the top to leverage [Jsonnet native functions]({{< ref "/docs/references/jsonnet#native-functions" >}}), `getEnv` which will retrieve an environment variable, and `getPath` which will retrieve the contents of a local file or URL.
- We defined our build object and signingKey using those function, working inside out:
  - We looked up the `VAULT_TOKEN` value using `getEnv`
  - We created a `path` for `getPath` using string formatting.  The path is the Vault API path to our signing key.  At the end, we added a syntax that `getPath` will use to set an HTTP header, `X-Vault-Token`, with the value from `getEnv`, which is used to authenticate to Vault with.
  - We parse the string returned by `getPath` to turn it into a Jsonnet object.
  - We retrieve the `data.private_key` value from the Jsonnet object.

### `signingCommands`

Some organizations may need to perform signing in a more secure, restrictive manner, like delegating signing to a HSM or another key provider.  We can use `signingCommands` for this.  `signingCommands` are a list of Commands that will use environment variables to sign a Token and return the signed Token to Etcha.

Etcha will set the [Environment Variable]({{< ref "/docs/references/commands#environment-variables" >}}) [`ETCHA_PAYLOAD`]({{< ref "/docs/references/commands#etcha_jwt" >}}) containing the base64 raw URL encoded string that needs to be signed.

Our signing Commands need to construct the rest of the [JWT]({{< ref "/docs/references/jwt" >}}), and then print the JWT during a Command `change` that triggers the event [`jwt`]({{< ref "/docs/references/events#jwt" >}}), like this:

```json
[
  {
    "always": true,
    "change": "<commands to build and print JWT>",
    "id": "build JWT",
    "onChange": [
      "etcha:jwt"
    ]
  }
]
```

## Building

Now that we've setup our signing keys, lets build this Pattern:

```
// patterns/myapp.jsonnet

local app_name = 'myapp';
local restart = function(name)
  {
    change: 'systemctl restart %s' % name,
    id: 'restart %s' % name,
  };

{
  build: [
    {
      always: true,
      change: 'make %s' % app_name,
      id: 'build %s' % app_name,
      onChange: [
        'etcha:buildManifest',
        'etcha:runEnv_myapp',
      ]
    },
  ],
  subject: app_name,
  run: [
    {
      change: 'curl -L https://s3.example.com/myapp_v2 -o /myapp',
      check: "myapp --version | grep v2",
      id: "copy %s v2' app_name,
      onChange: [
        'restart myapp',
      ],
    },
    restart(app_name),
  ]
}
```

We'll run [`etcha -c config.jsonnet build patterns/myapp.jsonnet myapp.jwt myapp`]({{< ref "/docs/references/cli#build" >}}).  This does a few things:

1. Import the Pattern Jsonnet file, `myapp.jsonnet`, and any other files it imports
2. Render the Pattern Jsonnet file
3. Run all of the [Commands]({{< ref "/docs/references/commands" >}}) in `build` in **Change Mode**.  Optionally using the [`Source`]({{< ref "/docs/references/config#sources" >}}) configuration [`exec`]({{< ref "/docs/references/config#exec" >}}), `myapp`, specified by the last parameter.
4. Collect metadata from these Commands to populate the [JWT]({{< ref "/docs/references/jwt" >}}), like [`buildManifest`]({{< ref "/docs/references/events#buildManifest" >}}) and [`runEnv_`]({{< ref "/docs/references/events#runEnv" >}}).  In our Pattern, the build command `build %s` will fire these events, and these values will contain the stdout of the `change` execution.
5. Create a JWT containing the _raw Pattern Jsonnet files_ collected in step 1, the metadata collected in step 4, and any JWT values set in our Pattern.
6. Sign the JWT using [`signingKey`](#signingkey) or [`signingCommands`](#signing-commands)
7. Save the JWT file to `myapp.jwt`, the parameter after the path to the Pattern.
