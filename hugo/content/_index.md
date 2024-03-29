---
description: Infinite Scale Configuration Management</br>For Distributed Platforms
title: Etcha
---

{{% blocks/section color="white" height=full %}}
<h1><b>Infinite Scale Configuration Management</b></h1>
<h2>Etcha creates serverless build and runtime configurations for your servers and apps.</h2>

```mermaid
flowchart TD
  style myapp fill:#d50000,fill-opacity:0.3,stroke:#d50000
  style storage fill:#d50000,fill-opacity:0.3,stroke:#d50000

  myapp{{patterns/myapp.jsonnet}}
  user([Users])

  subgraph Pull Mode&nbsp&nbsp&nbsp&nbsp&nbsp&nbsp&nbsp&nbsp&nbsp&nbsp&nbsp&nbsp&nbsp&nbsp&nbsp&nbsp&nbsp&nbsp
    storage(https://s3.example.com/myapp.jwt)
    servers[Servers and IoT]
    kubernetes["Kubernetes"]
  end

  subgraph &nbsp&nbsp&nbsp&nbsp&nbsp&nbsp&nbsp&nbsp&nbsp&nbsp&nbsp&nbsp&nbsp&nbsp&nbsp&nbsp&nbsp&nbsp&nbspPush Mode
    instance[Developer Instance]
  end

  user -- Write, Lint, and Test ----> myapp
  myapp -- Build, Sign, and Release ----> storage
  storage -- Pull, Verify, and Run ----> servers
  storage -- Pull, Verify, and Run ----> kubernetes
  myapp -- Build, Sign, Push, Verify, and Run -------> instance
```

{{% /blocks/section %}}
