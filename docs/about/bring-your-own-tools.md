---
categories:
  - feature
description: Etcha integrates with your existing processes and tooling.
title: Bring Your Own Tools
type: docs
---

```mermaid
flowchart LR
  style etchaBuild fill:#d50000,fill-opacity:0.3,stroke:#d50000
  style etchaLint fill:#d50000,fill-opacity:0.3,stroke:#d50000
  style etchaRun fill:#d50000,fill-opacity:0.3,stroke:#d50000
  style etchaTest fill:#d50000,fill-opacity:0.3,stroke:#d50000

  etchaBuild[etcha build]
  etchaLint[etcha lint]
  etchaRun[etcha run]
  etchaTest[etcha test]

  cicdServices[CI/CD Services - GitHub, GitLab, Jenkins]
  externalLinters[Linters - Open Policy Agent, Shellcheck]
  externalCrypto[External Cryptography - HSMs, HashiCorp Vault]
  externalStorage[Web Storage - S3 compatible, static file servers]

  etchaBuild --> cicdServices
  etchaBuild --> externalCrypto
  etchaBuild --> externalStorage
  etchaLint --> externalLinters
  etchaLint --> cicdServices
  etchaTest --> cicdServices
  etchaRun --> externalCrypto
  etchaRun --> externalStorage
```

Etcha is designed to integrate with your existing tooling and processes:

- Deploy Etcha builds from any HTTP/HTTPS endpoint, like S3 or an artifact registry
- Run Etcha build, lint, and test from your favorite CI/CD pipeline, like GitHub or GitLab
- Bring your own linters to check the commands Etcha will run, such as Open Policy Agent (OPA) or Shellcheck.
- Sign and verify Etcha builds using external key providers, like HashiCorp Vault or an HSM
