---
categories:
- reference
description: Reference documentation for Etcha's CLI
title: CLI
---

## Arguments

Arguments must be entered before commands.

### `-c [paths]` {#c}

Path to the JSON/Jsonnet [configuration file]({{< ref "/docs/references/config" >}}).

### `-f [format]`

Set log format (human, kv, raw, default: human).

### `-l [level]`

Set minimum log level (none, debug, info, error, default: info).

### `-n`

Disable colored log output.

### `-x [key=value,]`

Set config key=value (can be provided multiple times)

## Commands

### `autocomplete`

{{< autocomplete name="Etcha" >}}

### `build`

Import [Pattern]({{< ref "/docs/references/patterns" >}}) [Jsonnet]({{< ref "/docs/references/jsonnet" >}}) files from path, execute build [Commands]({{< ref "/docs/references/commands" >}}), sign a [JWT]({{< ref "/docs/references/jwt" >}}), and output the JWT to the destination path.

### `compare`

Compare two [JWTs]({{< ref "/docs/references/jwt" >}}) to see if they have the same etchaBuildManfiest, etchaPattern, and etchaVersion (can optionally ignore version mismatch).

### `copy`

Copy a local file or HTTP path to a destination path.  Utilizes the same function as Jsonnet [getFile]({{< ref "/docs/references/jsonnet#getFile" >}}) and can set HTTP headers in the source path using `#`.

### `dir`

Manages a directory on the local machine using check/change/remove.  Can optionally set permissions, owner, or group, otherwise permissions will be `0755` and the user and group will be inherited from the current user.

### `eula`

View the Etcha End User License Agreement (EULA).

### `file`

Manages a file on the local machine using check/change/remove.  Can optionally provide contents directly or via stdin, or set permissions, owner, or group, otherwise permissions will be `0644` and the user and group will be inherited from the current user.

### `gen-keys`

Generate cryptographic keys for use with signing and encryption.  The keys will be output as JSON:

{{< highlight json >}}
{
  "privateKey": <private key>,
  "publicKey": <public key>,
}
{{< /highlight >}}

See [Cryptography]({{< ref "/docs/references/cryptography" >}}) for more details around key formats and usage guides.

### `init`

Create folders, files, and libraries for developing [Patterns]({{< ref "/docs/references/patterns" >}}).  Subsequent runs of init will only update the files under `lib/etcha`.  See [libraries]({{< ref "/docs/references/libraries" >}}) for documentation on the modules created by init.

### `jq`

Query JSON from stdin using jq.  Supports standard JQ queries.

### `jwt`

Show the contents of a JWT.  Will also report any verification errors.

### `line`

Manage a line in a file or in text on the local machine from stdin using check/change.  Match is the regexp of the line to match, and the replacement text that will be set for the line.  If the line does not exist, it will be appended to the end of the file.  Replacement text can use capture groups from within the regexp, such as `${1}`.

### `link`

Manage a symlink on the local machinge using check/change.

### `lint`

Lint all `.jsonnet` and `.libsonnet` files in the path, checking the syntax and optionally the formatting of the files.  Can also use external linters to provide more validation. See [Linting Patterns]({{< ref "/docs/guides/linting-patterns" >}}) for more information.

### `local`

Import [Pattern]({{< ref "/docs/references/patterns" >}}) [Jsonnet]({{< ref "/docs/references/jsonnet" >}}) files from path, execute all [Commands]({{< ref "/docs/references/commands" >}}) in the specified mode locally.  Can optionally specify a [Config Source]({{< ref "/docs/references/config#sources" >}}) and a Parent ID to filter Commands with.

### `push`

Push ad-hoc commands or a signed pattern to a remote Etcha instance.  See [Running Commands]({{< ref "/docs/guides/running-commands" >}}) for more information.

### `render`

Render a Pattern from JWT or Jsonnet and display the result.

### `run`

Run Etcha in listening mode, periodically pulling new patterns, receiving new patterns via push, and exposing metrics.  Can specify an additional flag to only run once and exit.

### `shell`

Start an interactive shell on the remote Etcha instance.  See [Shell Access]({{< ref "/docs/guides/shell-access" >}}) for more information.

### `show-config`

Show the rendered config from all sources (files, environment variables, and command line arguments).

### `show-pattern`

Show the rendered pattern of a JWT or pattern file.

### `test`

Test all patterns in path, optionally filtering for specific Command Parent IDs.  See [Testing Patterns]({{< ref "/docs/guides/testing-patterns" >}}) for more information.
