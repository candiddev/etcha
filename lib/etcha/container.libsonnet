// Manages a container using Docker or Podman.  Will automatically detect the engine from sysinfo, or one can be provided.

local file = import './file.libsonnet';

function(args='', cap_add=[], command='', deviceChown=false, devices=[], engine='', engineEnv={}, env=[], image, name, network='', options='', ports=[], privileged=false, pull='', restart='on-failure', skipRemove=false, user='', volumes=[], workdir='')
  local run = '%(args)s --name %(name)s %(cap_add)s %(devices)s %(env)s %(network)s %(options)s --restart %(restart)s %(ports)s %(privileged)s %(pull)s %(user)s %(volumes)s %(workdir)s %(image)s %(command)s' % {
    args: args,
    cap_add: std.join(' ', [
      '--cap-add %s' % cap

      for cap in cap_add
    ]),
    command: command,
    devices: std.join(' ', [
      '--device=%s:%s' % [device, device]

      for device in devices
    ]),
    env: std.join(' ', [
      '-e %s' % env

      for env in env
    ]),
    image: image,
    name: name,
    network: if network == '' then '' else '--network ' + network,
    options: options,
    ports: std.join(' ', [
      '-p %s' % port

      for port in ports
    ]),
    privileged: if privileged then '--privileged' else '',
    pull: if pull == '' then '' else '--pull always',
    restart: restart,
    user: if user == '' then '' else '--user ' + user,
    volumes: std.join(' ', [
      '--volume %s' % volume

      for volume in volumes
    ]),
    workdir: if workdir == '' then '' else '--workdir ' + workdir,
  };

  local e = if engine == '' then std.native('getConfig')().vars.sysinfo.containerEngine else engine;

  local rm = 'rm' + if e == 'podman' then ' -t0' else '';
  local userParts = std.split(user, ':');

  {
    commands: [
      [
        file(group=if std.length(userParts) == 2 then userParts[1] else '""', mode='""', owner=if userParts[0] == '' then '""' else userParts[0], path=device, skipRemove=true)

        for device in devices
        if deviceChown
      ] + [
        {
          id: 'start container %s' % name,
          check: "[[ $(%s inspect %s -f '{{.State.Running}}') = true ]]" % [e, name],
          change: '%(e)s %(rm)s -f %(name)s || true; %(e)s run -d %(run)s' % {
            e: e,
            name: name,
            rm: rm,
            run: run,
          },
          env: engineEnv,
          locks: [
            'start container %s' % name,
          ],
          remove: if !skipRemove then '%s %s -f %s' % [e, rm, name],
        },
      ],
    ],
    id: 'container %s' % name,
    serial: true,
  }
