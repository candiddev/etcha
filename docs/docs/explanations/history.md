---
categories:
- explanation
description: A brief history of Etcha
title: History
---

## Forge

Once a time, we created a really basic spaghetti script called Forge: https://github.com/candiddev/forge.

Forge was designed to script opinionated Debian installations, using local tools and bash only:

- Bootstrap a common partition layout including LUKS
- Build systemd-nspawn containers, VM images, and squashfs
- Only install the bare minimum of packaging
- Support multiple architectures
- Support SSH authentication using tools like Vault
- Use systemd for everything

Forge worked really well for us, and was a key component of our imaging process.  Our SaaS application, [Homechart](https://homechart.app), relied heavily on Forge.

Forge had a few problems though:

- Extremely difficult to test
- Painful to run using GitHub Actions
- Constant struggle to make it idempotent
- Too easy to mix build and runtime configurations

## Forge v2

We set out to rewrite Forge using Go to address a few of the problems:

- Maintainability
- Strict API
- Testing

What we ended up creating was more abstractions over common Linux tools.  At this point we could've doubled down, recreated more or less the common core-utils as Go functions (or used an existing [library](https://github.com/u-root/u-root)).  Instead, we looked at running scripts using Go.

We broke down a "Command" into three components:

- `check`: All of the `if`s we used in Forge for making a script idempotent
- `change`: The imperative work to be done
- `remove`: What to execute to remove the work

Right from the beginning we wanted our tool to be stateful-ish.  When you're developing images, being able to idempotently re-run and remove things speeds up your development time _considerably_. 

One key feature we added in Forge v2 was **signing the build**.  We created a rudimentary JWT that was to be included in the build.  We could use this JWT in conjunction with Secrets tooling like Vault or other federated identity providers to establish secret zero:  this server was built with this image, and here's the JWT to prove it.

## Dark Days of Go Templating

What's better than running scripts using Bash?  Not using Go templating to do it.  Unfortunately, we pursued this route for a few months, even going so far as creating a custom import syntax and linter.

Go templating presented a few problems:

- No one understands the syntax
- It's evaluated strictly top-down
- Whitespace chomping always bites you in the ass
- Bash scripts are already hard enough to read, adding Go templating just made them worse

This is what it looked like (warning: **it's bad**)
```
{{ range $path := list "/etc/apt/preferences.d" "/etc/apt/trusted.gpg.d" "/var/cache/apt/archives/partial" "/var/lib/apt/lists" "/var/lib/dpkg" }}
{{ template_file "../shared/dir.etc" decode_yaml `
path: %s%s
` $dirCache $path }}
{{ end }}

{{ template_file "../shared/file.etc" decode_yaml `
path: %s/var/lib/dpkg/status
` $dirCache }}

{{ template_file "../shared/apt.etc" decode_yaml `
packages:
  - ca-certificates
  - curl
  - gpg
` }}
```

We needed a real language, not just templating.

## Embracing Jsonnet

We evaluated three different configuration languages:

- CUE Lang
- Dhall
- Jsonnet

We ended up choosing Jsonnet because:

- Better Go support
- Battle tested (and developed by) Google
- More familiar syntax
- Smaller, more extensible standard library

We now could build images using Jsonnet, outputing a JSON list of Commands to run to build images.

We were using Ansible to deploy configurations on top of these images, and the next key feature would surface: how could we make ansible better?

- It's very slow
- Pushing over SSH is problematic and hard to limit/bootstrap

## JWTs and Pulls

We began prototyping using JWTs as a delivery method for configurations:

- Host them on object storage
- Version them like the rest of our release artifacts
- Ensure they were validated/verified using signing

This would give us a serverless, decentralized way to build and run our images and apps.

This is around the time we started to look at changing the name from Forge to Etcha too.

## Etcha

We released Etcha in October of 2023.  The initial version is a culmination of our work and usage of Etcha internally to run our SaaS platform.  It supports the common things we needed from Ansible, Bash, Packer, and even Terraform:

- Imperative scripting (and sometimes declarative, using a function to abstract the imperative bits)
- Stateful
- Lint-able and testable
- Scale beyond anything we would ever need

## EtchaOS

In June of 2024, we release EtchaOS: minimal, immutable, in-memory variants of popular Linux distributions like Debian, Fedora, and Ubuntu.  EtchaOS is powered by Etcha and is designed for containers, systemd, and IoT deployments.  Using EtchaOS, companies can rapidly deploy applications and services at scale using Etcha Patterns across cloud providers, on-premise, and edge.
