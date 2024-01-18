---
categories:
- guide
description: How to write Patterns which contain build and runtime configurations in Etcha.
title: Writing Patterns
weight: 20
---

In this guide, we'll go over initializing a directory, creating a Command library, and creating a Pattern.

## Initialize a Directory

In your code repository, run the command [`etcha init`]({{< ref "/docs/references/cli#init" >}}).  This command will do a few things:

- Scaffold a common development environment.
- Populate handy, validated, Jsonnet libraries to kickstart your development process.

After running this command, you'll have a directory structure like this:

- `lib`: This folder should contain subfolders with various [Jsonnet libraries]({{< ref "/docs/references/jsonnet" >}}).  You'll import these libraries into Patterns.  All of these files should have a `.libsonnet` extension.
- `lib/etcha`: This folder contains validated Jsonnet libraries.  You probably shouldn't change the files in here (instead, open a pull request!).  See [Libraries]({{< ref "/docs/references/libraries" >}}) for more information.  Etcha will add a comment at the top containing the Etcha version that generated these files.
- `patterns`: This folder should contain the main pattern files you end up building.  The files in this directory should have a `.jsonnet` extension.

Subsequent usage of this command will only overwrite the files in `lib/etcha`.  A Git diff should show you what changed, be sure to check this file into source control!

## Creating Commands

After the directory is initialized, you can start creating Command libraries and helpers, or move straight into creating Patterns.

You can create command libraries anywhere, but you should try and keep them under `lib`.  You may want to create folders based on business unit, function, or application--your folder structure will most likely follow conways law.  **It's highly recommended to centralize your Command and Patterns into one repository**, both from a sharing of code to testing/linting/signing.

A command library might look like this:

```
// lib/myorg/systemdRestart.libsonnet

function(name)
  {
    always: true,
    change: 'systemctl restart %s' % name,
    id: 'restart service %s' % name
  }
```

## Creating Patterns

Once you've built a few Command libraries (or skipped those for now), you'll need to build Patterns to push or pull to your Etcha instances.

It's recommended to create Patterns under the `patterns` folder, possibly even under subfolders.  Most Patterns should be tightly scoped to limit their length.  Similar to Commands, how you organize your Patterns will mirror your organizational structure.

Patterns contain build and run Command lists, as well as values for [JWTs]({{< ref "/docs/references/jwt" >}}).  A rendered Pattern may look like this:

```json
{
  "build": [
    {
      "change" "make myapp1",
      "check": "[[ -e myapp1 ]]",
      "id": "build myapp1"
    },
    {
      "change" "make myapp2",
      "check": "[[ -e myapp2 ]]",
      "id": "build myapp2"
    }
  ],
  "run": [
    {
      "change": "curl -L https://s3.example.com/myapp1_v2 -o /myapp1",
      "check": "myapp1 --version | grep v2",
      "id": "copy myapp1 v2",
      "onChange": [
        "restart myapp1"
      ]
    },
    {
      "change": "systemctl restart myapp1",
      "id": "restart myapp1"
    },
    {
      "change": "curl -L https://s3.example.com/myapp2_v2 -o /myapp2",
      "check": "myapp2 --version | grep v2",
      "id": "copy myapp2 v2",
      "onChange": [
        "restart myapp2"
      ]
    },
    {
      "change": "systemctl restart myapp2",
      "id": "restart myapp2"
    }
  ]
}
```

### Using Jsonnet

Patterns are written in Jsonnet though, so we'll include our Command libraries instead of writing everything twice:

```
local makeApp = import '../lib/myorg/makeApp.libsonnet';
local runApp = import '../lib/myorg/runApp.libsonnet';

{
  build: [
    makeApp(myapp1),
    makeApp(myapp2),
  ],
  run: [
    runApp(runApp1),
    runApp(runApp2),
  ]
}
```

### Reusing Patterns

Jsonnet lets us create reusable functions that we can compose into Patterns.  We can also reuse Patterns if we want, too:

```
local corePattern = import './core.jsonnet'
local makeApp = import '../lib/myorg/makeApp.libsonnet';
local runApp = import '../lib/myorg/runApp.libsonnet';

{
  build: [
    makeApp(myapp1),
    makeApp(myapp2),
  ],
  run: [
    corePattern.run,
    runApp(runApp1),
    runApp(runApp2),
  ]
}
```

### Nested Lists

You can also nest lists within `build` or `run`:

```
local makeApp = import '../lib/myorg/makeApp.libsonnet';
local runApp = import '../lib/myorg/runApp.libsonnet';

{
  build: [
    [
      [
        [
          makeApp(myapp1),
          makeApp(myapp2),
        ]
      ]
    ]
  ]
}
```

Etcha will flatten lists into a single, ordered list.  If you're curious to see how your Pattern might look once it's rendered, run [`etcha render`]({{< ref "/docs/references/cli#render" >}})).

### Build, Run, and Rendering

The `build` commands in a Pattern are ran during [`etcha build`]({{< ref "/docs/references/cli#build" >}})), most likely on your local instance or a CI/CD runner.

The `run` commands in a Pattern are ran on an Etcha instance after pulling or pushing the Pattern.  **All Patterns are rendered immediately before they are used**.  If you use [dynamic lookups]({{< ref "/docs/references/jsonnet#native-functions" >}})) in your Pattern, like `getEnv`, `getRecord`, or `getURL`, those functions will be called and rendered on the _instance performing the run_.

`build` and `run` lists are not required to have Commands.  A server configuration may not have any `build` Commands, just `run` Commands.
