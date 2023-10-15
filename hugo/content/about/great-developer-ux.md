---
categories:
  - feature
description: Etcha provides a great user experience for creating configurations.
title: Great Developer UX
type: docs
---

```bash
$ cat helloworld.jsonnet
local n = import '../lib/etcha/native.libsonnet';

{
  build: [
    {
      always: true,
      change: 'echo %s > /work/helloworld' % n.getEnv('HOSTNAME'),
      id: 'write a file',
      onChange: [
        'read file',
      ],
    },
    {
      change: 'cat /work/helloworld',
      id: 'read file',
      onChange: [
        'etcha:build_manifest',
      ],
    },
  ],
  buildExec: {
    command: '/bin/sh -c'
  },
}
$ etcha lint helloworld.jsonnet
ERROR shared/go/jsonnet/import.go:49
error importing jsonnet files: RUNTIME ERROR: couldn't open import "../lib/etcha/native.libsonnet": no match locally or in the Jsonnet library paths
        helloworld.jsonnet:1:11-49
```

Etcha makes writing configurations easier than ever before:

- Configurations are written using Jsonnet, a formal configuration language.  It's easy to understand, and even easier to create functions, libraries, and more abstractions for your users.
- Jsonnet has integrations with editors like VSCode for features like autocomplete and syntax checking.
- Etcha can lint and test commands to ensure they function exactly as you expect.
