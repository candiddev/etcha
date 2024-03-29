---
categories:
- reference
description: Reference documentation for Etcha's Events
title: Events
---

Etcha's event system allows for users to trigger and chain [Patterns]({{< ref "/docs/references/patterns" >}}) dynamically.  Etcha also uses this event system to extract information from [Commands]({{< ref "/docs/references/commands" >}}).

## Triggering

All Events are triggered using Command [`onChange`]({{< ref "/docs/references/commands#onchange" >}}) values--Events are specified here by prefixing their name with `etcha:`:

```json
{
  "always": true,
  "change": "echo hello event handler!",
  "id": "event issuer",
  "onChange": {
    "etcha:my_event"
  }
}
```

## Handling

Events are sent and received from [Sources]({{< ref "/docs/references/config#sources" >}}).  A Source specifies the events it can receive and send.

## System Events

These are the event names Etcha already knows about and what they're used for:

### buildManifest

Firing this event during a `build` will add the output of the [Command's `change`]({{< ref "/docs/references/commands#change" >}}) to the [JWT's `etchaBuildManifest` property]({{< ref "/docs/references/jwt#etchabuildmanifest" >}}).

### jwt

#### signingCommands

Firing this event during [`signingCommands`]({{< ref "/docs/references/config#signingcommands" >}})) will set the output of the [Command's `change`]({{< ref "/docs/references/commands#change" >}}) to be the entire JWT generated by build.

#### verifyCommands

Firing this event during [`verifyCommands`]({{< ref "/docs/references/config#verifycommands" >}}) will have Etcha use the output of the [Command's `change`]({{< ref "/docs/references/commands#change" >}})) for the JWT Token.  Etcha will also **not verify** the token, as it assumes the verify commands have passed.

**DO NOT TRIGGER THIS EVENT IF THE TOKEN IS NOT VERIFIED**

### runVar_

Firing any event with this prefix during a `build` will add the output of the [Command's `change`]({{< ref "/docs/references/commands#change" >}}) to the [JWT's `etchaRunVars` property]({{< ref "/docs/references/jwt#etcharunvars" >}}).

### stderr

Firing this event will log the output of the [Command's `change`]({{< ref "/docs/references/commands#change" >}}) to stderr.

### stdout

Firing this event will log the output of the [Command's `change`]({{< ref "/docs/references/commands#change" >}}) to stdout.

### webhookBody

Firing this event during a [Webhook]({{< ref "/docs/guides/running-patterns#remote-run-via-webhooks" >}}) will have Etcha use the output of the [Command's `change`]({{< ref "/docs/references/commands#change" >}}) for the webhook response.

### webhookContentType

Firing this event during a [Webhook]({{< ref "/docs/guides/running-patterns#remote-run-via-webhooks" >}}) will have Etcha use the output of the [Command's `change`]({{< ref "/docs/references/commands#change" >}}) for the webhook content-type header.
