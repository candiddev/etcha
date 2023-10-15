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

## Cryptography

Etcha leverages boring, secure cryptographic keys for signing and verifying JWTs.  See [Cryptography](../../references/cryptography) for more information.

## Security Best Practices

Etcha can be a target for malicious usage.  See below for best practices on running Etcha in a secure manner:

### Don't Allow Push Access on the Internet

Etcha's Push mode is the only way for attackers to access an Etcha instance remotely.  The attacker would need to know a [Source](../../references/configs#sources) name, as well correctly sign a verifiable JWT for that source.  They would also need to do do all of this without being rate-limited.  

While this is highly unlikely to occur, it's best to avoid exposing Etcha on the Internet.

### Limit Source Execution

When running Etcha with multiple Sources, say for each application team, you should limit how the Sources are executed with some kind of sandbox technology, like containers, chroot, or cgroups.

### Protect Your Keys

Protect your signing keys to avoid leaking them.  Store them in a secure manner, and use separate signing keys wherever possible.

### Test Your Signing and Verify Commands

Delegating signing and verification is very useful for controlling the process, but it also puts all of the responsibility on you to implement it correctly!  Etcha cannot guarantee the JWT validity this way, and trusts you to understand the validation process.  Ensure you test these processes well to ensure you don't end up trusting every JWT.

### Use Trusted Pull Targets

Don't store your JWTs on web services you do not control.  While it's highly unlikely an attacker can bypass the cryptographic protections around JWTs, running Etcha in pull mode with only trusted targets is recommended.
