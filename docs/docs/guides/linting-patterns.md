---
categories:
- guide
description: How to lint Patterns for syntax errors, security issues, and formatting mistakes in Etcha.
title: Linting Patterns
weight: 30
---

Etcha can lint Patterns, libraries, and more--basically if it's written in Jsonnet, Etcha can probably lint it.

## Performing Linting

You can lint an entire path or specific files using {{% cli lint %}}.  Lint will traverse directories and perform linting on all `.jsonnet` and `.libsonnet` files.  It will ensure they can be imported into Etcha correctly.  Any errors will be shown in the console, and the tool will exit with a non-zero status code.

You can also check the formatting of the files by adding the flag `-f` : `etcha -c lint mydir`.  Formatting errors will be reported, along with diffs on what the correct formatting should be.  The tool will exit with a non-zero status code on formatting errors, too.

**For Continuous Delivery/Continuous Integration Usage**, it's highly recommended to run linting across your entire Etcha codebase.

## Test Mode

Linting and Testing both set a flag within the config called {{% config test %}} to `true`.  You can retrieve this value within Jsonnet and adjust your Pattern files to render differently during test mode, i.e.:

```
// lib/mylib.libsonnet
local n = import '../etcha/native.libsonnet';
local config = n.getConfig();

{
  check: (if config.test then '' else '[[ -d "/mydir" ]]'),
  id: 'hello world',
}
```

In this example, `check` will be an empty string in test mode.

## External Linters

In addition to linting the Jsonnet syntax, Etcha can combine the `change`, `check`, and `remove` scripts and pass them through external linters via stdin.  These external linters are configured under {{% config lint %}}.

Out of the box, Etcha is configured to use Shellcheck as an external linter.  Any external linter failures will be reported in the console, and Etcha will exit with a non-zero code.

{{% alert title="Candid Commentary" color="info" %}}
Shellcheck is ran using a container.  If you have particularly long Patterns with lots of Commands, you may see a large speed improvement by installing shellcheck locally and changing the linter configuration to use the local executable instead of spinning up a container for every Command.
{{% /alert %}}
