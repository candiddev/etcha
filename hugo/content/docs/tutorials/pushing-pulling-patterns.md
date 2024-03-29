---
categories:
- tutorial
description: How to push and pull a simple Pattern.
title: Pushing and Pulling Patterns
---

In this tutorial, we'll push and pull a Pattern using Etcha.  Please follow all steps, even if you completed the last Tutorial.

## Requirements

- Docker or Podman (we'll use Docker here, but this should work with Podman, too)
- Access to pull down Etcha from GitHub's container registry (ghcr.io)
- A text editor

## Tutorial

### Prepare Our Environment

1. Open a local, empty directory in a shell like bash where we can read/write files and mount them into a container.
2. Create a temporary bash alias for Etcha so we can use it:

```bash
alias etcha='docker run --network etcha -u $(id -u):$(id -g) --rm -v $(pwd):/work -w /work ghcr.io/candiddev/etcha:latest'
```

3. Initialize a new container network and directory with Etcha:

```bash
$ docker network create etcha
$ mkdir etcha
$ etcha init .
```

We need to use a custom network so our Etcha containers can communicate with each other.

4. Lets write a new Pattern using a few of the Etcha libraries.  Create a new file under `patterns` called `run.jsonnet`.  Add in this content:

```
local n = import '../lib/etcha/native.libsonnet';

{
  run: [
    {
      always: true,
      change: 'echo %s > /work/hostname' % n.getEnv('HOSTNAME', 'fallback'),
      id: 'write a file',
      onChange: [
        'copy file',
      ],
      remove: 'rm /work/hostname',
    },
    {
      change: 'cp /work/hostname /work/hostname2',
      id: 'copy file',
      remove: 'rm /work/hostname2',
    },
  ],
  runExec: {
    command: '/bin/sh -c'
  },
}
```

## Running Etcha Listener

1. Lets configure Etcha with a [Source]({{< ref "/docs/guides/running-patterns#sources" >}}) that allows pushes and has `verifyKeys`, and run Etcha in listen mode:

```bash
$ docker run -d --name etcha_listen \
    --network etcha -p 4000:4000 \
    -u $(id -u):$(id -g) \
    -v $(pwd):/work -w /work \
    ghcr.io/candiddev/etcha:latest \
    -x run_systemMetricsSecret=secret \
    -x sources_listen='{
      "allowPush":true,
      "runExec": {
        "allowOverride":true
      },
      "verifyKeys": [
        "ed25519public:MCowBQYDK2VwAyEAw7eTEuEH0+TfgtX3zB+JZVnYD0eskY6qn3n7ZCA7wWM=:reqYEklgP4"
      ]
    }' run
```

The container should've started listening:

```
$ docker logs etcha_listen
level="INFO" function="etcha/go/run/run.go:59" status=200 success=true  message="Starting source runner..."
level="ERROR" function="etcha/go/pattern/jwt.go:56" status=500 success=false error="error reading JWT: error opening src: error opening src: open /work/etcha/listen.jwt: no such file or directory"
level="INFO" function="etcha/go/run/run.go:166" status=200 success=true  message="Generating self-signed certificate for listener..."
level="INFO" function="etcha/go/run/run.go:184" status=200 success=true  message="Starting listener..."
```

The errors at the beginning are normal, Etcha can't find an existing JWT for our source.  Etcha also generated a self-signed certificate for us to use.

2. Lets make sure we can access the web interface by pulling some metrics:

```bash
$ curl -sk https://localhost:4000/etcha/v1/system/metrics?key=secret
# HELP go_gc_duration_seconds A summary of the pause duration of garbage collection cycles.
# TYPE go_gc_duration_seconds summary
go_gc_duration_seconds{quantile="0"} 0.000104377
go_gc_duration_seconds{quantile="0.25"} 0.000104377
go_gc_duration_seconds{quantile="0.5"} 0.000104377
go_gc_duration_seconds{quantile="0.75"} 0.000104377
...
```

We should see a bunch of [metrics]({{< ref "/docs/guides/monitoring" >}}).  Nothing interesting yet--we haven't ran any commands.

## Pushing a Pattern

1. Lets push our run pattern to our instance:

```bash
$ etcha -x build_signingKey=ed25519private:MC4CAQAwBQYDK2VwBCIEIBq+BhDRYk8OJv1ksMwKtf0td5p3FGwypXq96gHKefGS:reqYEklgP4 \
    push -h localhost listen patterns/run.jsonnet
ERROR error performing request: Post "https://localhost:4000/etcha/v1/push/listen": tls: failed to verify certificate: x509: certificate is not valid for any names, but wanted to match localhost
localhost:
    ERROR: error performing request: Post "https://localhost:4000/etcha/v1/push/listen": tls: failed to verify certificate: x509: certificate is not valid for any names, but wanted to match localhost
```

Etcha couldn't verify the remote instance certificate.  That's OK, we can specify an additional argument to skip verification:

```bash
$ etcha -x build_signingKey=ed25519private:MC4CAQAwBQYDK2VwBCIEIE6dSkW4jnn3tx119BKw8+zOmhJyzTOsBlWcjqaHxMcX:ZcxoeWfSRt \
    -x build_pushTLSSkipVerify=true push -h localhost listen patterns/run.jsonnet
localhost:
    ERROR: push didn't match any sources
```

What happened?  Well, we used the wrong `signingKey`--the remote Etcha instance couldn't verify our push.  Lets use the right one and try again:

```bash
$ etcha -x build_signingKey=ed25519private:MC4CAQAwBQYDK2VwBCIEIBq+BhDRYk8OJv1ksMwKtf0td5p3FGwypXq96gHKefGS:reqYEklgP4 \
    -x build_pushTLSSkipVerify=true push -h localhost listen patterns/run.jsonnet
localhost:
    Changed 2: write a file, copy file
```

That's better.  We successfully pushed our config!  We can see the files `hostname` and `hostname2` exists:

```bash
$ cat hostname
871cf521e491
```

2. Lets push a new Pattern, only this time we'll remove the run commands and add a new one.  Create a new file called `pattern/newfile.jsonnet` with this content:

```
local n = import '../lib/etcha/native.libsonnet';

{
  run: [
    {
      change: 'echo %s > /work/newfile' % n.getEnv('HOSTNAME'),
      check: '[[ -e /work/newfile ]]',
      id: 'write a new file',
      remove: 'rm /work/newfile',
    },
  ],
  runExec: {
    command: '/bin/sh -c'
  },
}
```

Lets push the file:

```bash
 $ etcha -x build_signingKey=ed25519private:MC4CAQAwBQYDK2VwBCIEIBq+BhDRYk8OJv1ksMwKtf0td5p3FGwypXq96gHKefGS:reqYEklgP4 \
    -x build_pushTLSSkipVerify=true push -h localhost listen patterns/newfile.jsonnet
localhost:
    Changed 1: write a new file
    Removed 2: copy file, write a file
```

This time, Etcha ran the `remove` commands from the two commands we removed, and then ran the `change` for `write a new file`.  Sure enough, the old files `hostname` and `hostname2` are gone, and only `newfile` remains:

```bash
$ ls
etcha  lib  newfile  patterns  README.md
```

3. Lets push the same Pattern again.  Since it has a check, it should result in no changes:

```bash
$ etcha -x build_signingKey=ed25519private:MC4CAQAwBQYDK2VwBCIEIBq+BhDRYk8OJv1ksMwKtf0td5p3FGwypXq96gHKefGS:reqYEklgP4 \
    -x build_pushTLSSkipVerify=true push -h localhost listen patterns/newfile.jsonnet
localhost:
    No changes
```

## Pulling a Pattern

1. Lets remove newfile, all of the files under Etcha, and stop our existing container:

```bash
$ rm newfile
$ rm etcha/*
$ docker rm -f etcha_listen
```

2. Now we're going to start a new container in pull mode.  It will pull a JWT, `listen.jwt`, every 5 seconds:

```bash
$ docker run -d --name etcha_listen \
    --network etcha -p 4000:4000 \
    -u $(id -u):$(id -g) \
    -v $(pwd):/work -w /work \
    ghcr.io/candiddev/etcha:latest \
    -x run_systemMetricsSecret=secret \
    -x sources_listen='{
      "runExec": {
        "allowOverride":true
      },
      "pullPaths": [
        "/work/listen.jwt"
      ],
      "verifyKeys": [
        "ed25519public:MCowBQYDK2VwAyEAw7eTEuEH0+TfgtX3zB+JZVnYD0eskY6qn3n7ZCA7wWM=:reqYEklgP4"
      ]
    }' run
$ docker logs etcha_listen
level="INFO" function="etcha/go/run/run.go:59" status=200 success=true  message="Starting source runner..."
level="ERROR" function="etcha/go/pattern/jwt.go:56" status=500 success=false error="error reading JWT: error opening src: error opening src: open /work/etcha/listen.jwt: no such file or directory"
level="INFO" function="etcha/go/run/run.go:166" status=200 success=true  message="Generating self-signed certificate for listener..."
level="INFO" function="etcha/go/run/run.go:184" status=200 success=true  message="Starting listener..."
```

3. Lets build `listen.jwt` from `patterns/run.jsonnet`:

```bash
$ etcha -x build_signingKey=ed25519private:MC4CAQAwBQYDK2VwBCIEIE6dSkW4jnn3tx119BKw8+zOmhJyzTOsBlWcjqaHxMcX:ZcxoeWfSRt \
    build patterns/run.jsonnet listen.jwt
```

4. Lets see if it Etcha pulled the JWT:

```bash
$ docker logs -n 5 etcha_listen
level="ERROR" function="etcha/go/pattern/jwt.go:73" status=500 success=false error="error parsing JWT for source listen: error verifying signature against message"
level="ERROR" function="etcha/go/pattern/jwt.go:77" status=500 success=false error="no valid targets for source listen"
level="ERROR" function="etcha/go/pattern/jwt.go:43" status=500 success=false error="error verifying signature against message"
level="ERROR" function="etcha/go/pattern/jwt.go:73" status=500 success=false error="error parsing JWT for source listen: error verifying signature against message"
level="ERROR" function="etcha/go/pattern/jwt.go:77" status=500 success=false error="no valid targets for source listen"
```

Turns out, we built the JWT with the wrong signature.  Thankfully, Etcha couldn't verify it and didn't run it.

5. Lets rebuild with the right key and see what happens:

```bash
$ etcha -x build_signingKey=ed25519private:MC4CAQAwBQYDK2VwBCIEIBq+BhDRYk8OJv1ksMwKtf0td5p3FGwypXq96gHKefGS:reqYEklgP4 \
    build patterns/run.jsonnet listen.jwt
$ docker logs -n 5 etcha_listen
level="INFO" function="etcha/go/run/run.go:97" status=200 success=true sourceTrigger="pull" sourceName="listen" message="Updating config for listen..."
level="INFO" function="etcha/go/commands/command.go:95" status=200 success=true sourceTrigger="pull" sourceName="listen" commandID="write a file" commandMode="check" message="Always changing write a file..."
level="INFO" function="etcha/go/commands/command.go:129" status=200 success=true sourceTrigger="pull" sourceName="listen" commandID="write a file" commandMode="check" commandMode="change" message="Changing write a file..."
level="INFO" function="etcha/go/commands/command.go:97" status=200 success=true sourceTrigger="pull" sourceName="listen" commandID="copy file" commandMode="check" message="Triggering copy file via write a file..."
level="INFO" function="etcha/go/commands/command.go:129" status=200 success=true sourceTrigger="pull" sourceName="listen" commandID="copy file" commandMode="check" commandMode="change" message="Changing copy file..."
```

That's better.  And our hostname files reappeared:

```bash
$ ls
etcha  hostname  hostname2  lib  listen.jwt  patterns  README.md
```

6. Lets build the other Pattrn, save the JWT to the same location, and observe the changes:
```bash
$ etcha -x build_signingKey=ed25519private:MC4CAQAwBQYDK2VwBCIEIBq+BhDRYk8OJv1ksMwKtf0td5p3FGwypXq96gHKefGS:reqYEklgP4 \
    build patterns/newfile.jsonnet listen.jwt
$ docker logs -n 5 etcha_listen
level="INFO" function="etcha/go/commands/command.go:63" status=200 success=true sourceTrigger="pull" sourceName="listen" commandID="copy file" commandMode="remove" message="Removing copy file..."
level="INFO" function="etcha/go/commands/command.go:63" status=200 success=true sourceTrigger="pull" sourceName="listen" commandID="write a file" commandMode="remove" message="Removing write a file..."
level="INFO" function="etcha/go/commands/command.go:129" status=200 success=true sourceTrigger="pull" sourceName="listen" commandID="copy file" commandMode="check" commandMode="change" message="Changing copy file..."
level="INFO" function="etcha/go/run/run.go:97" status=200 success=true sourceTrigger="pull" sourceName="listen" message="Updating config for listen..."
level="INFO" function="etcha/go/commands/command.go:129" status=200 success=true sourceTrigger="pull" sourceName="listen" commandID="write a new file" commandMode="check" commandMode="change" message="Changing write a new file..."
```

Just like with the push, Etcha diff'd the new Pattern, created `newfile`, and removed `hostname` and `hostname2`:

```bash
$ ls
etcha  lib  newfile  newfile.jwt  patterns  README.md
```

7. Before we finish the tutorial, lets check out those metrics and see if anything interesting has shown up:

```bash
$ curl -sk https://localhost:4000/etcha/v1/system/metrics?key=secret | grep '^etcha'
etcha_commands_total{error="0",id="copy file",mode="remove",source="listen"} 1
etcha_commands_total{error="0",id="write a file",mode="remove",source="listen"} 1
etcha_commands_total{error="0",id="copy file",mode="change",source="listen"} 1
etcha_commands_total{error="1",id="copy file",mode="check",source="listen"} 1
etcha_commands_total{error="0",id="write a file",mode="change",source="listen"} 1
etcha_commands_total{error="1",id="write a file",mode="check",source="listen"} 1
etcha_commands_total{error="0",id="write a new file",mode="change",source="listen"} 1
etcha_commands_total{error="1",id="write a new file",mode="check",source="listen"} 1
etcha_sources_commands{mode="change",name="listen",trigger=""} 1
etcha_sources_commands{mode="remove",name="listen",trigger=""} 2
etcha_sources_total{error="0",name="listen",trigger=""} 2
```

Etcha surfaces metrics for all of the Pattern runs.  We can se the number of times a command was called, whether it errored (which for `check`, means run `change`), and how many times a source was triggered.

8. Remove the Etcha container and network:

```bash
$ docker rm etcha_listen
$ docker network rm etcha
```

## Summary

We've successfully pushed and pulled two different Patterns and saw the changes.  Next, we'll trigger Patterns using Events and Webhooks.
