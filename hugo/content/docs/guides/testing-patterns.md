---
categories:
- guide
description: How to test Patterns in Etcha.
title: Testing Patterns
weight: 40
---

Etcha can test Patterns to ensure they actually work.  Because of the nature of `change`, `check`, and `remove`, testing is more or less free.

## Performing Testing

You can test an entire path or specific files using [`etcha test`]({{< ref "/docs/references/cli#test" >}})).  Test will traverse directories and perform testing on all `.jsonnet` files.  Testing is performed by executing the Commands:

1. Run all Commands in [**Change Mode**]({{< ref "/docs/references/commands#change-mode" >}}) (execute `check`, and if triggered, `change`).  Any `change` errors will cause testing to fail.
2. Run all Commands in [**Check Mode**]({{< ref "/docs/references/commands#check-mode" >}}) (execute `check` only).  Any `check` errors that would normally cause a `change` to occur will cause testing to fail (check/command is not idempotent).
3. Run all Commands in [**Remove Mode**]({{< ref "/docs/references/commands#remove-mode" >}}) (reverse order, execute `remove`).  Any `remove` errors will cause testing to fail.
4. Run all Commands in [**Check Mode**]({{< ref "/docs/references/commands#check-mode" >}}) (execute `check` only).  Any IDs that previously failed `check`/changed during step 1 that do not fail `check` again will cause testing to fail.

By default, Etcha will only test `run` Commands in a Pattern.  To test `build` Commands too, pass an additional argument after the path: `etcha test mydir yes`.

Testing is performed under the [`source`({{< ref "/docs/references/config#sources" >}}) "`test`".  It's highly recommended to configure an alternative `exec` configuration in here, such as a container, to avoid testing impacting a local system.

**For Continuous Delivery/Continuous Integration Usage**, it's highly recommended to run testing across your entire Etcha codebase.

## Test Mode

Linting and Testing both set a flag within [vars passed to the Pattern]({{< ref "/docs/references/patterns#vars" >}}), `test`, to `true`.  You can retrieve this value within Jsonnet and adjust your Pattern files to render differently during test mode, i.e.:

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
