---
categories:
- explanations
description: An overview of Etcha's security
title: Security
---

## Bug Bounties

Etcha does not yet have an established bug bounty program.  
Please [contact us](mailto:info@candid.dev?subject=Etcha%20Bug) if you think you've found a bug or security issue with Etcha.

## CVEs

Etcha does not have any CVEs.  When a CVE is reported, it will be listed on this page.

## Code

Etcha is developed using Go.  Here are some of the methods we use to help keep Etcha's code free of vulnerabilities:

- **Auth Test Suite**: We use an extensive authentication and authorization test suite for every pull request and build.
- **Limit Third-Party Libraries**: We try and use as few third-party libraries as possible, and when we do select a third-party library, we review the codebase to ensure it's something we are comfortable maintaining.
- **Secure Software Supply Chain**: We require a clean `govulncheck` for every pull request and build.

## Cryptography

Etcha leverages boring, secure cryptographic keys for signing and verifying JWTs.  See [Cryptography]({{< ref "/docs/references/cryptography" >}}) for more information.

## Security Best Practices

Etcha can be a target for malicious usage.  See below for best practices on running Etcha in a secure manner:

### Don't Allow Push Access on the Internet

Etcha's Push mode is the only way for attackers to access an Etcha instance remotely.  The attacker would need to know a {{% config sources source %}} name, and correctly sign a verifiable JWT for that source.  They would also need to do all of this without being rate-limited.

While this is highly unlikely to occur, it's best to avoid exposing Etcha on the Internet.

{{% alert title="Candid Commentary" color="info" %}}
We have some of our Etcha instances exposed on the Internet, and we operate honeypot instances to check for automated attacks.
{{% /alert %}}

### Limit Source Execution

When running Etcha with multiple Sources, say for each application team, you should limit how the Sources are executed with some kind of sandbox technology, like containers, chroot, or cgroups.

### Protect Your Keys

Protect your signing keys to avoid leaking them.  Store them in a secure manner, and use separate signing keys wherever possible.

{{% alert title="Candid Commentary" color="info" %}}
[Rot](https://rotx.dev) works really well for this.
{{% /alert %}}


### Test Your Signing and Verify Commands

Delegating signing and verification is very useful for controlling the process, but it also puts all of the responsibility on you to implement it correctly!  Etcha cannot guarantee the JWT validity this way, and trusts you to understand the validation process.  Ensure you test these processes well to ensure you don't end up trusting every JWT.

### Use Trusted Pull Targets

Don't store your JWTs on web services you do not control.  While it's highly unlikely an attacker can bypass the cryptographic protections around JWTs, running Etcha in pull mode with only trusted targets is recommended.
