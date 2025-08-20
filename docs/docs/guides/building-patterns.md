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

{{% alert title="Candid Commentary" color="info" %}}
Depending on what you include in the build commands, the JWT may include sensitive values.  Keep in mind:

- Base64 encoding is not the same thing as encryption.
- Base64 encoded things are not encrypted.
{{% /alert %}}


## Signing Patterns

We need a way to sign Patterns before building them.  Out of the box, Etcha cannot build or sign Patterns without [cryptographic keys]({{< ref "/docs/references/cryptography" >}}).  Specifically, we need to provide values in our Etcha configuration for {{% config build_signingKey %}} or {{% config build_signingCommands %}}.

### `signingKey`

A `signingKey` will use a static, private key to sign the Pattern JWT.  This value can be hardcoded in your "builder", or you can retrieve it from environment variables, a remote URL, a DNS record, etc.  This text is like a password though and should be protected.

We can generate keys appropriate for signing using {{% cli gen-keys %}}.  We'll use the `privateKey` value during build for the `signingKey`, and then we'll save the `publicKey` and use that for the {{% config run_verifyKeys %}} when we push/pull.

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

{{% alert title="License Required" color="warning" %}}
This requires an [Unlimited License]({{< ref "/pricing" >}})
{{% /alert %}}

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

{{% alert title="Candid Commentary" color="info" %}}
This is behind a license purely from a footgun standpoint.  If this is something you need to do, we want to make sure you're doing it correctly.
{{% /alert %}}

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
3. Run all of the [Commands]({{< ref "/docs/references/commands" >}}) in `build` in **Change Mode**.  Optionally using {{% config sources %}} configuration {{% config exec %}}, `myapp`, specified by the last parameter.
4. Collect metadata from these Commands to populate the [JWT]({{< ref "/docs/references/jwt" >}}), like [`buildManifest`]({{< ref "/docs/references/events#buildManifest" >}}) and [`runEnv_`]({{< ref "/docs/references/events#runEnv" >}}).  In our Pattern, the build command `build %s` will fire these events, and these values will contain the stdout of the `change` execution.
5. Create a JWT containing the _raw Pattern Jsonnet files_ collected in step 1, the metadata collected in step 4, and any JWT values set in our Pattern.
6. Sign the JWT using [`signingKey`](#signingkey) or [`signingCommands`](#signing-commands)
7. Save the JWT file to `myapp.jwt`, the parameter after the path to the Pattern.

{{% alert color="primary" title="Passing Variables to Run" %}}
When building Patterns, you may have dynamic values in {{% config vars %}} or from a [Jsonnet native function]({{% ref "/docs/references/jsonnet" %}}).  Etcha can intelligently bundle these values using the [`get`]({{% ref "/docs/references/jsonnet#get" %}}) function:

```
local native = 'etcha/lib/etcha/native.libsonnet';
{
  run: [
    id: 'set a secret',
    always: true,
    change: 'echo %s > /tmp/mysecret' % native.get(field='secret', default=native.getEnv('SECRET')),
  ]
}
```

In this example, Etcha will get the environment variable `SECRET` during the build process and, assuming the variable was defined, cache the result to a field, `secret`, that will be stored within the JWT.

Etcha will not lookup the environment variable during the `run` process, instead using the cached result from the `build` process.
{{% /alert %}}
