---
categories:
- reference
description: Reference documentation for Etcha's configuration
title: Config
---

{{% snippet config_format Etcha etcha %}}

## Configuration Values

{{% snippet config_key "build_pushDomainSuffix" %}}

String, a domain suffix to append to all [`targets`](#targets) for hostname resolution.

**Default:** `""`

{{% snippet config_key "build_pushMaxWorkers" %}}

Number, worker threads to use for pushing Patterns to targets.

**Default:** Number of CPUs

{{% snippet config_key "build_signingCommands" "(recommended)" %}}

{{% alert title="License Required" color="warning" %}}
This requires an [Unlimited License]({{< ref "/pricing" >}})
{{% /alert %}}

List of [Commands]({{< ref "/docs/references/commands" >}}) to run when signing a JWT instead of using a [`signingKey`](#build_signingKey).  See [Building Patterns]({{< ref "/docs/guides/building-patterns#signingcommands" >}}) for more information.

**Default:** `[]`

{{% snippet config_key "build_signingExec" %}}

[Exec](#exec) configuration for running [`signingCommands`](#build_signingCommands).

**Default:** `{}`

{{% snippet config_key "build_signingKey" "(recommended)" %}}

String, the [cryptographic signing key]({{< ref "/docs/references/cryptography" >}}) to use when signing JWTs.  See [Building Patterns]({{< ref "/docs/guides/building-patterns" >}}) for more information.

**Default:** `""`

{{% snippet config_key "build_signingKeyPath" "(recommended)" %}}

String, a path to a [cryptographic signing key]({{< ref "/docs/references/cryptography" >}}) to use when signing JWTs.  See [Building Patterns]({{< ref "/docs/guides/building-patterns" >}}) for more information.  If specified as just a filename, Etcha will ascend directories until it finds a matching filename.  Can be used to provide [keys used with Rot](https://rotx.dev).

**Default:** `""`

{{% snippet config_cli etcha %}}

{{% snippet config_key "exec" %}}

The main `exec` configuration.  Can be overridden by other configurations.  The format for all `exec` configurations is below.  The defaults for the main `exec` are:

```json
{
  "allowOverride": true,
  "command": "/usr/bin/bash -e -o pipefail -c",
}
```

{{% snippet config_key "exec_allowOverride" %}}

Boolean, enables this Exec environment to be overridden by other Exec configurations.  Exec overrides may be deeply nested, like this:

```mermaid
flowchart LR
  commandExec["command exec"]
  patternExec["pattern runExec"]
  sourceExec["source exec"]

  exec --> sourceExec
  sourceExec --> patternExec
  patternExec --> commandExec
```

Every exec in the path needs to allow overrides for command exec to be allowed.


**Default:** `false`

{{% snippet config_key "exec_command" %}}

String, the command to run before any [Commands]({{< ref "/docs/references/commands" >}}).  If this is specified, other commands will be added after it.  You'd typically put a shell interpreter in here, like `/usr/bin/bash -c`.

**Default:** `""`

{{% snippet config_key "exec_containerEntrypoint" %}}

String, override the container entrypoint for the [`containerImage`](#exec_containerImage).

**Default:** `""`

{{% snippet config_key "exec_containerImage" %}}

String, the container image to use.  If specified, [`command`](#exec_command), [`env`](#exec_env), and [`workDir`](#exec_workDir) will be passed/ran in a container.

**Default:** `""`

{{% snippet config_key "exec_containerNetwork" %}}

String, the container network to use.  Defaults to the default network for the container runtime.

**Default:** `""`

{{% snippet config_key "exec_containerPrivileged" %}}

Boolean, run the container as a privileged container.

**Default:** `false`

{{% snippet config_key "exec_containerPull" %}}

String, the container pull policy.

**Default:** `""`

{{% snippet config_key "exec_containerUser" %}}

String, the container user or UID.

**Default:** `""`

{{% snippet config_key "exec_containerVolumes" %}}

List of strings, the volumes to map into the container.

**Default:** `[]`

{{% snippet config_key "exec_containerWorkDir" %}}

String, override the WorkDir of a container.

**Default:** `""`

{{% snippet config_key "exec_env" %}}

Map of strings in the format `"ENVIRONMENT": "value"`, will set these as environment variables.  During override, keys will be merged with parent `exec` values, overwriting any existing ones that are defined.

**Default:** `{}`

{{% snippet config_key "exec_envInheirt" %}}

Boolean, whether to inherit the environment of the main Etcha process.

**Default:** `false`

{{% snippet config_key "exec_group" %}}

String, the group or GID to run the commands with.  Using this is typically privileged and requires root or extra capabilities.

**Default:** `""`

{{% snippet config_key "exec_sudo" %}}

Bool, will check if user is root and run the Commands using `sudo -E` if the user is not root.

**Default:** `false`

{{% snippet config_key "exec_user" %}}

String, the user or UID to run the commands with.  Using this is typically privileged and requires root or extra capabilities.

**Default:** `""`

{{% snippet config_key "exec_workDir" %}}

String, the working directory to execute commands within.

**Default:** `""`

{{% snippet config_httpClient Etcha %}}

{{% snippet config_httpServer ":4000" etcha %}}

{{% snippet config_jsonnet false %}}

{{% snippet config_licenseKey Etcha %}}

{{% snippet config_key "lint_exclude" %}}

String, a regexp of files to exclude from linting.

**Default:** `"etcha.jsonnet"`

{{% snippet config_key "exec_linters" "(recommended)" %}}

A map of strings and [Exec](#exec) configurations for linters.  These linters are ran using {{% cli lint %}}.  See [Linting Patterns]({{< ref "/docs/guides/linting-patterns" >}}) for more information.

**Default:**
```json
{
  "shellcheck": {
    "command": "-s bash -e 2154 -",
    "containerImage": "koalaman/shellcheck"
  }
}
```

{{% snippet config_key "run_jwtFilters" %}}

List of strings representing JWT property filters that must be present for a JWT to accepted.  This setting applies to all JWTs that Etcha parses.

The filter format is a list of comma separated filters.  Within the list, filters are OR, within a comma separated string, filters are AND:

```json
{
  "run": {
    "jwtFilters": [
      "aud=^a$,sub",
      "aud=^b",
    ],
  }
}
```

In this example, the JWT must match one of these conditions:

- Contain the property `aud` with a value that starts with `a`, and contain the property `sub` with any value.
- Contain the property `aud` with a value that starts with `b`.

**Default:** `[]`

{{% snippet config_key "run_randomizedStartDelaySec" %}}

Number, the maximum seconds to delay starting Etcha in listen mode.  A random number between 0 and this number will be chosen, and Etcha will wait to start listening for that amount of time.  Used to prevent thundering herds and accidental concurrent configuration applies.

**Default:** `0`

{{% snippet config_key "run_shellTimeout" %}}

String, the maximum duration until a shell session is ended.  {{% config_duration %}}

**Default:** `"10m"`

{{% snippet config_key "run_stateDir" "(recommended)" %}}

String, path to a writeable directory where Etcha can store patterns for future diffing.  Used during {{% cli run %}}.  Defaults to current working directory if unset.

**Default:** `""`

{{% snippet config_key "run_systemMetricsKey" "(recommended)" %}}

String, the secret to protect `/etcha/v1/system/metrics` endpoint during {{% cli run %}}.  Setting this value enables the metrics endpoint.  See [Monitoring]({{< ref "/docs/guides/monitoring" >}}) for more information.

**Default:** `""`

{{% snippet config_key "run_verifyCommands" "(recommended)" %}}

{{% alert title="License Required" color="warning" %}}
This requires an [Unlimited License]({{< ref "/pricing" >}})
{{% /alert %}}

List of [Commands]({{< ref "/docs/references/commands" >}}) to run when verifying a JWT instead of using [`verifyKeys`](#run_verifyKeys).  See [Running Patterns]({{< ref "/docs/guides/running-patterns#verifycommands" >}}) for more information.

**Default:** `[]`

{{% snippet config_key "run_verifyExec" %}}

[Exec](#exec) configuration for running [`verifyCommands`](#run_verifycommands).

**Default:** `{}`

{{% snippet config_key "run_verifyKeys" "(recommended)" %}}

List of [cryptographic verify keys]({{< ref "/docs/references/cryptography" >}}) to use when verifying JWTs.  See [Running Patterns]({{< ref "/docs/guides/running-patterns" >}}) for more information.

**Default:** `[]`

{{% snippet config_key "sources" %}}

Sources is a map of source names to source configurations.  See [Running Patterns]({{< ref "/docs/guides/running-patterns" >}}) for more information.

**Default:** `{}`

### `sources_[source]_allowPush` {#sources_allowPush}

Boolean, allow a source to receive patterns via push.

**Default:** `false`

### `sources_[source]_checkOnly` {#sources_checkOnly}

Boolean, prevents patterns received on this source from running change commands.

**Default:** `false`

### `sources_[source]_commands` {#sources_commands}

List of static [Commands]({{< ref "/docs/references/commands" >}}) to run for this source.  If allowed, Pattern pushes and pulls will override the list of Commands.  Commands will use the source's `exec` config, if allowed by the main `exec` config.  Commands will be ran at startup unless `triggerOnly` is set to true.  See [Running Commands]({{< ref "/docs/guides/running-commands#static-source-commands" >}}) for more information.

### `sources_[source]_eventsReceive` {#sources_eventsReceive}

List of event names that the source patterns can receive.

**Default:** `[]`

### `sources_[source]_eventsSend` {#sources_eventsSend}

String, a regular expression to match event names that the source patterns can send.  Can specify `".*"` to allow everything.

**Default:** `""`

### `sources_[source]_exec` {#sources_exec}

[Exec](#exec) configuration for the source.

**Default:** `{}`

### `sources_[source]_jwtFilters` {#sources_jwtFilters}

[run_jwtFilters](#run_jwtFIlters) configuration for the source.  Will replace global `run_jwtFilters` for this source.

**Default:** `[]`

### `sources_[source]_noRemove` {#sources_noRemove}

Boolean, never remove [Commands]({{< ref "/docs/references/commands" >}}) for a [Pattern]({{< ref "/docs/references/patterns" >}}) [source](#sources) when diffing.

**Default:** `false`

### `sources_[source]_noRestore` {#sources_noRestore}

Boolean, prevents Etcha from saving/restoring the JWTs for this `source`.  Useful for preventing `push-commands` from re-running at startup.

### `sources_[source]_pullIgnoreVersion` {#sources_pullIgnoreVersion}

Boolean, don't consider `etchaVersion` property differences in [JWTs]({{< ref "/docs/references/jwt" >}}) to require a new pull.

**Default:** `false`

### `sources_[source]_pullPaths` {#sources_pullPaths}

List of paths to pull JWTs from for this source.  Can be local disk paths or http/https paths.  For http/https paths, HTTP headers can be specified by appending `#header:value` and separating headers using `\r\n`, e.g. `#header1:value1\r\nheader2:value2`.  A special header, `skipVerify`, can also be added to ignore certificate verification errors.

See [Running Patterns]({{< ref "/docs/guides/running-patterns" >}}) for more information.

**Default:** `[]`

### `sources_[source]_runFrequency` {#sources_runFrequency}

String, duration between (optionally pulling) and running the source pattern.  Setting this to "" means the source will never be pulled/ran except at startup.  {{% config_duration %}}

**Default:** `""`

### `sources_[source]_runMulti` {#sources_runMulti}

Boolean, allows for multiple runs to of the source to happen at the same time.  By default, multiple runs will queue.  Some scenarios where this might occur include repeated pushes, or pulls with too low of a [`runFrequencySec`](#sources_runfrequencysec).  Use with caution.

**Default:** `false`

### `sources_[source]_shell` {#sources_shell}

[Exec](#exec) configuration for starting a shell via {{% cli shell %}}.  Setting the `command` value will enable shell access.

**Default:** `""`

### `sources_[source]_triggerOnly` {#sources_triggerOnly}

Boolean, when `true`, never run a Pattern unless it's triggered via [Events or Webhooks]({{< ref "/docs/guides/running-patterns" >}}).

**Default:** `false`

### `sources_[source]_verifyCommands` {#sources_verifyCommands}

{{% alert title="License Required" color="warning" %}}
This requires an [Unlimited License]({{< ref "/pricing" >}})
{{% /alert %}}

See [Run > verifyCommands](#run_verifyCommands).  Setting this value overrides `run.verifyCommands`.

**Default:** `[]`

### `sources_[source]_verifyExec` {#sources_verifyExec}

See [Run > verifyExec](#run_verifyExec).  Setting this value overrides `run.verifyExec` if `run.verifyExec` allows overrides.

**Default:** `{}`

### `sources_[source]_verifyKeys` {#sources_verifyKeys}

See [Run > verifyKeys](#run_verifyKeys).  Setting this value appends it to `run.verifyKeys`.

**Default:** `[]`

### `sources_[source]_webhookPaths` {#sources_webhookPaths}

List of HTTP paths to listen for webhooks.  See [Running Patterns]({{< ref "/docs/guides/running-patterns" >}}) for more information.

**Default:** `[]`

### `sources_[source]_vars` {#sources_vars}

A map of Source-specific [`vars`](#vars).

**Default:** `{}`

{{% snippet config_key "targets" %}}

A map of target names to target options for use by {{% cli push %}} and {{% cli shell %}}:

```json
{
  "targets": {
    "server1": {
      "hostname": "server1.example.com",
      "insecure": true,
      "port": 4001,
      "sources": [
        "core",
        "nginx"
      ],
      "vars": {
        "selinux": true
      }
    }
  }
}
```

See [Running Patterns]({{< ref "/docs/guides/running-patterns" >}}) for more information.

Targets are meant to be flexible **and proxyable**.  Etcha uses standard HTTP/HTTP2 functionality, including Server-Sent Events (SSE) for shell access, and it should work out of the box with most reverse proxies (NGINX, Traefik, HAProxy, etc).  You could have all of your Etcha devices behind a proxy and use separate paths (`/server1/etcha/v1/push`) or host-based routing on your proxy.  This would be the Etcha equivalent of a jumpbox.

### `targets_[target]_hostname` {#targets_hostname}

String, the hostname or IP address of the target (default: the target name).

**Default:** The target name

### `targets_[target]_insecure` {#targets_insecure}

Boolean, will use an insecure (not HTTPS) connection when connecting to the target.

**Default:** `false`

### `targets_[target]_pathPush` {#targets_pathPush}

String, the URL path for the push endpoint without any sources.

**Default:** `"/etcha/v1/push"`

### `targets_[target]_pathShell` {#targets_pathShell}

String, the URL path for the shell endpoint without any sources.

**Default:** `"/etcha/v1/shell"`

### `targets_[target]_port` {#targets_port}

String, the port number of the target.

**Default:** `"4000"`

### `targets_[target]_sourcePatterns` {#targets_sourcePatterns}

A map of source names to Pattern paths or Commands.  Etcha will push to this target if the source is specified with `etcha push`.  If the Pattern is an empty string, Etcha will allow any Pattern or Command to be pushed if the Source is matched.

**Default:** `{}`

### `targets_[target]_vars` {#targets_vars}

A map of Target-specific [`vars`](#vars).

**Default:** `false`

{{% snippet config_key "vars" %}}

A map of strings and any type of value.  Can be used during rendering to get/set values.  See [Patterns - Variables]({{< ref "/docs/references/patterns#variables" >}}), [Building Patterns]({{< ref "/docs/guides/building-patterns" >}}), and [Running Patterns]({{< ref "/docs/guides/running-patterns" >}}) for more information.

Vars are combined in this order:

- [Pattern `etchaRunVars`]({{< ref "/docs/references/patterns#runvars" >}})
- Top level `vars` (this value)
- [Source `vars`](#sourcevars)
- [Target `vars` (push only)](#targetvars)

Vars can be retrieved in Patterns using [`getConfig`]({{< ref "/docs/references/patterns#runvars" >}}).

**Default:** `{}`

The following `vars` are added to all Patterns:

#### `source`

String, the source name of the Pattern.

#### `sysinfo`

Map of values containing useful system information:

- `containerEngine`: String, the detected container engine (docker or podman)
- `cpuLogical`: Number, count of logical CPUs
- `defaultInterface`: String, the interface name used by the default route
- `fqdn`: String, the FQDN of the machine
- `hostname`: String, the short hostname of the machine
- `interfaces`: Map of interface details
- `ipv4`: List of IPv4 addresses in CIDR format (`1.1.1.1/24`)
- `ipv6`: List of IPv6 addresses in CIDR format (`::1/64`)
- `mac`: String, MAC address of interfaces
- `kernelRelease`: String, the kernel release name (`uname -r`)
- `kernelVersion`: String, the kernel version (`uname -v`)
- `machine`: String, the machine version, (like `x86_64`)
- `memoryTotal`: Number, total system memory in MB
- `osID`: String, the `ID` of the OS (`debian`)
- `osIDLike`: String, the `ID _LIKE` of the OS (`debian`)
- `osName`: String, the `NAME` of the OS (`Ubuntu`)
- `osType`: String, type of OS (`linux`)
- `osVersion`: String, the `VERSION` of the OS (`24.04 LTS (Noble Numbat)`)
- `osVersionCodename`: String, the `VERSION_CODENAME` of the OS (`bookworm`)
- `osVersionID`: String, the `VERSION_CODENAME` of the OS (`12`)
- `packageManager`: String, the package manager in use (`apt`)
- `runtimeArch`: String, architecture of the Etcha binary (`amd64`)

#### `test`

Boolean, will be `true` if a Pattern is in [test mode]({{< ref "/docs/guides/testing-patterns" >}}).


#### Build Vars

During a Pattern [`build`]({{< ref "/docs/guides/building-patterns" >}}), the following additional `vars` will be set:

- `srcDir`: String, the directory of the Pattern being built.
- `srcPath`: String, the path of the Pattern being built.

#### Push Vars

During a Pattern [`push`]({{< ref "/docs/guides/running-patterns" >}}), the following additional `vars` will be set:

- `source`: String, the source name that was pushed.
- `target`: String, the target name that is being pushed to.

#### Run Vars

During a Pattern [`run`]({{< ref "/docs/guides/running-patterns" >}}), the following additional `vars` will be set:

- `jwt`: String, the contents of the original JWT if the Pattern was run from a JWT.
