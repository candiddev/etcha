---
categories:
  - feature
description: Etcha can run commands based on webhooks and events.
title: Webhook and Event Triggers
type: docs
---

```mermaid
flowchart LR
  etl[ETL Pattern]
  externalApp[External App]
  etcha1[Etcha Instance 1]
  etcha2[Etcha Instance 2]

  externalApp -- Sends Webhook --> etcha1
  etcha1 -- Runs --> etl
  etl -- Sends Webhook --> etcha2
```

Etcha can run configurations based on the result of a push or pull.  It can also run configurations when events fire from within Etcha as well as from customizable webhook paths.

- Rapidly build ETL pipelines triggered by events and webhooks
- Create simple event driven applications and scripts
- Expose system sockets and other local, non-HTTP services
