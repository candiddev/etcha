---
description: Etcha simplifies and scales your distributed application delivery. Define, lint, test, and build your infrasturcture as code in Jsonnet as Patterns.  Dynamically run Commands and Patterns with webhooks and event handlers. Securely access your devices over HTTP. Streamline your application infrastructure with Etcha.
title: Etcha | Full-Stack Configuration Management for Developers and Sysadmins
---

{{% blocks/section color="white" %}}
<h1 style="border-bottom: 2px solid var(--bs-red)"><b>Buld and Release Infrastructure as Code with Etcha</b></h1>
<h1>Full-Stack Configuration Management for Developers and Sysadmins</h1>
<div style="align-items: center; display: flex; justify-content: center; padding-top: 40px; width 100%">
  <a class="button button--red" href="/docs/guides/install-etcha">Download</a>
</div>
{{% /blocks/section %}}

{{< blocks/section color="white" type=row >}}
{{% blocks/feature icon="fa-wrench" title="Release Artifacts, Not Breaking Changes" %}}
Build, test, and release infrastructure as code using Patterns--a list of procedural Commands for Etcha to perform, written in [Jsonnet](https://jsonnet.org), and deployed like any other software release artifact.
{{% /blocks/feature %}}

{{% blocks/feature icon="fa-magnifying-glass" title="Lint and Test Your Infrastructure Changes" %}}
Lint, test, and validate your Patterns before deployment, ensuring they function as intended. Prevent costly downtime.
{{% /blocks/feature %}}

{{% blocks/feature icon="fa-arrows-turn-to-dots" title="Centralize or Decentralize Your Deployments" %}}
Push Patterns to Etcha clients for centralized control, or pull Patterns from artifact storage for decentralized environments.
{{% /blocks/feature %}}

{{% blocks/feature icon="fa-gear" title="Create Dynamic Infrastructure with Events and Webhooks" %}}
Trigger Commands and Patterns based on webhooks and events. Streamline workflows and respond rapidly to changes.
{{% /blocks/feature %}}

{{% blocks/feature icon="fa-terminal" title="No SSH Required, Shell Access Included" %}}
Securely connect to any device via HTTP-based shell. Manage your entire infrastructure without opening extra ports.
{{% /blocks/feature %}}
{{< /blocks/section >}}

{{< blocks/section color=white >}}
<h2 style="border-bottom: 2px solid var(--bs-red)"><b>Write Your Configurations Using Real Code</b></h2>
<h3>No more YAML, no more Jinja</h3>
{{< highlight bash >}}
local makeApp = import '../lib/myorg/makeApp.libsonnet';
local runApp = import '../lib/myorg/runApp.libsonnet';

{
  build: [
    makeApp(myapp1),
    makeApp(myapp2),
  ],
  run: [
    runApp(runApp1),
    runApp(runApp2),
  ]
}
{{< /highlight >}}

<h2 style="border-bottom: 2px solid var(--bs-red); padding-top: 50px"><b>Deploy Configurations Once, Read From Many</b></h2>
<h3>Manage thousands of devices quickly and efficiently</h3>
{{< highlight bash >}}
$ etcha build mypattern.jsonnet mypattern.jwt
$ s3 cp mypattern.jwt s3://mybucket/mypattern.jwt
$ etcha listen
INFO  Updating config for myapplication
INFO  Always changed write a file [100ms]
{{< /highlight >}}

<h2 style="border-bottom: 2px solid var(--bs-red); padding-top: 50px"><b>Event Driven Automation</b></h2>
<h3>Create bespoke webook integrations to quickly prototype or define new workflows</h3>
{{< highlight bash >}}
$ etcha listen
INFO  Always changed trigger event [1ms]
INFO  Changed trigger event [200ms]
INFO  Running source myapp for event trigger from ID trigger event
{{< /highlight >}}

{{< /blocks/section >}}
