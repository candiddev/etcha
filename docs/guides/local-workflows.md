---
categories:
- guide
description: How to run Commands and functions from the Etcha CLI.
title: Local Workflows
weight: 70
---

In this guide, we'll go over turning Jsonnet Pattern lists into [Macros]({{% ref "/docs/references/macros" %}}), exposing them in the main Etcha CLI command list.

## Use Cases

With Macros, you can easily create CLI-driven workflows for developers that use your existing Jsonnet Pattern and Command definitions.  This could replace your existing CLI and CI/CD workflows powered by Makefiles or Bash scripts.

{{% alert title="Candid Commentary" color="info" %}}
We added Local Workflows to replace all of our build, deploy, lint, release, and test Bash scripts.  Combined with parallel Commands support, it dramatically shrunk our CI/CD times, made our workflows more discoverable, and easier to extend.
{{% /alert %}}

## Macro Files

By default, Etcha looks for Macros under the `etcha/macros` directory (or specified using {{% config macroDir %}}).  Etcha will walk this directory and find all files ending in `.jsonnet` and ignore other files.  For each file, Etcha will add it under {{% config cli_macros %}}:

- The macro name will be the relative path, replacing `/` with `-`.  For example, the file `etcha/macros/build/homechart/go.jsonnet` would be named `build-homechart-go`.
- If the Jsonnet file can be a single Command, a list of Commands, a Pattern, or a function that returns these.
- Any comments above the return value will be added to {{% config cli_macros_usage %}}.
- If the Jsonnet file contains a function, the function arguments will be converted to {{% config cli_macros_argumentsRequired %}} and {{% config cli_macros_argumentsOptional %}}:
```
function(target, source='etcha')
```

```json
{
  "argumentsOptional": [
    "source"
  ],
  "argumentsRequired": [
    "target"
  ]
}
```

## Usage

Once you've added Macro files, they'll display in the Etcha CLI automatically.  Etcha's CLI matching using abbreviations will also work, so you can run `etcha b-h-g` to run the Macro `etcha build-homechart-go`.

Here's a snippet of what our CLI usage like:

```bash
$ etcha
Usage: etcha [global flags] <command/macro>

    ________       __         
   / ____/ /______/ /_  ____ _
  / __/ / __/ ___/ __ \/ __ `/
 / /___/ /_/ /__/ / / / /_/ / 
/_____/\__/\___/_/ /_/\__,_/  
Configuration management that evolves with your infrastructure

By using this software you agree to the End-User License Agreement (EULA) for Etcha. To view the EULA, type `etcha eula`.

Macros:
containers-cleanup <app>
  Cleanup old containers on GitHub registry.

containers-release <app>
  Release a container.

etcha-deploy <target> <app>
  Deploy something using Etcha.

etcha-lint
  Lint etcha configs and patterns.

etcha-lint
  Lint etcha configs and patterns.

git-clean-ignored
  Clean ignored git files from repository.

git-prune
  Prune git branches and tags that don't exist on origin from repository.

go-build <app> [buildTags] [buildVars] [osArch] [osType]
  Build Go apps.
```
