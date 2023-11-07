---
categories:
- reference
description: Reference documentation for Etcha's CLI
title: CLI
---

## Arguments

Arguments must be entered before commands.

### `-c [paths]`

Path to JSON/Jsonnet [configuration files](../config) separated by a comma.

### `-f [format]`

Set log format (human, kv, raw, default: human).

### `-l [level]`

Set minimum log level (none, debug, info, error, default: info).

### `-n`

Disable colored log output.

### `-x [key=value,]`

Set config key=value (can be provided multiple times)

## Commands

### `build [pattern path] [destination path] [config source, default: etcha]` {#build}

Import [Pattern](../patterns) [Jsonnet](../jsonnet) files from path, execute build [Commands](../commands), sign a [JWT](../jwt), and output the JWT to the destination path.  Can optionally specify a [Config Source](../config#sources).

### `compare [new jwt path or URL] [old jwt path or URL] [ignore version, default: no]` {#compare}

Compare two [JWTs](../jwt) to see if they have the same etchaBuildManfiest, etchaPattern, and etchaVersion (can optionally ignore version mismatch).

### `generate-keys [encrypt-asymmetric, encrypt-symmetric, sign-verify]` {#generate-keys}

Generate cryptographic keys for use with signing and encryption.  The keys will be output as JSON:

For encrypt-asymmetric and sign-verify, the format is:

{{< highlight json >}}
{
  "privateKey": <private key>,
  "publicKey": <public key>,
}
{{< /highlight >}}

For encrypt-symmetric, the format is:

{{< highlight json >}}
{
  "key": <key>
}
{{< /highlight >}}

See [Cryptography](../cryptography) for more details around key formats and usage guides.

### `init [directory, default: curret directory]` {#init}

Create folders, files, and libraries for developing [Patterns](../patterns).  Subsequent runs of init will only update the files under `lib/etcha`.  See [libraries](../libraries) for documentation on the modules created by init.

### `jq [jq query options]` {#jq}

Query JSON from stdin using jq.  Supports standard JQ queries, and the `-r` flag to render raw values.

### `lint [path] [check formatting, default: no]` {#lint}

Lint all `.jsonnet` and `.libsonnet` files in the path, checking the syntax and optionally the formatting of the files.  Can also use external linters to provide more validation. See 
[Linting Patterns](../../guides/linting-patterns) for more information.

### `local-change [pattern path] [config source, default: etcha]` {#run-change}

Import [Pattern](../patterns) [Jsonnet](../jsonnet) files from path, execute all [Commands](../commands) in change mode locally.  Can optionally specify a [Config Source](../config#sources).

### `local-remove [pattern path] [config source, default: etcha]` {#run-remove}

Import [Pattern](../patterns) [Jsonnet](../jsonnet) files from path, execute all remove [Commands](../commands) locally.  Can optionally specify a [Config Source](../config#sources).

### `push-command [command] [destination URL]` {#push-command}

Push an ad-hoc command to a destination URL.  See [Running Commands](../../guides/running-commands) for more information.

### `push-pattern [pattern path] [destination URL]` {#push-pattern}

Build and sign the [Pattern](../patterns) from path and push it to the destination URL.

### `run-listen` {#run-listen}

Run Etcha in listening mode, periodically pulling new patterns, receiving new patterns via push, and exposing metrics.

### `run-once` {#run-once}

Run Etcha patterns via [Config Sources](../config#sources), pull new patterns, and run the diff.

### `show-config` {#show-config}

Show the rendered config from all sources (files, environment variables, and command line arguments).

### `show-jwt [jwt path]` {#show-jwt}

Show the contents of a JWT.  Will also report any verification errors.

### `show-pattern [jwt or pattern path]` {#show-pattern}

Show the rendered pattern of a JWT or pattern file.

### `test [path] [test build commands, default: no]` {#test}

Test all patterns in path.  See [Testing Patterns](../../guides/testing-patterns) for more information.
