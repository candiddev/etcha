---
categories:
  - feature
description: Etcha can infinitely scale your configuration delivery.
title: Infinite Scale
type: docs
---

```mermaid
flowchart TD
  s3[https://s3.example.com]
  user1[Security Engineer]
  user2[Developer]
  user3[Platform Engineer]
  server1[Linux Server #1...]
  server2[Linux Server #10,000...]
  prometheus[Prometheus]

  user1 -- Builds stig.jwt ---> s3
  user2 -- Builds myapp.jwt ---> s3
  user3 -- Builds kubelet.jwt ---> s3
  s3 -- Pulls stig.jwt, myapp.jwt ---- server1
  s3 -- Pulls stig.jwt, kubelet.jwt ---- server2
  server1 -- Pulls config metrics ---- prometheus
  server2 -- Pulls config metrics ---- prometheus
```

Etcha builds configurations into portable, signed files you can distribute from anywhere.

- Serve your configurations from a service like S3 to handle millions (or more) of clients.  It's just a text file!
- Etcha exposes metrics for services like Prometheus to monitor the distribution of configurations
- Configurations are rendered locally and rapidly applied
