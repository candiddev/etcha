// Manages a container using Docker or Podman.  Will automatically detect the engine from sysinfo, or one can be provided.

function(args='', cap_add=[], command='', devices=[], engine='', engineEnv={}, env=[], image, name, network='', options='', ports=[], privileged=false, pull='', restart='on-failure', skipRemove=false, user='', volumes=[], workdir='')
  local run = '%(args)s --name %(name)s %(cap_add)s %(devices)s %(env)s %(network)s %(options)s %(restart)s %(ports)s %(privileged)s %(pull)s %(user)s %(volumes)s %(workdir)s %(image)s %(command)s' % {
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
    restart: if restart == '' then '' else '--restart ' + restart,
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
    id: 'container %s' % name,
    check: "[[ $(%s inspect %s -f '{{.State.Running}}') = true ]]" % [e, name],
    change: '%(e)s %(rm)s -f %(name)s || true; %(e)s run -d %(run)s' % {
      e: e,
      name: name,
      rm: rm,
      run: run,
    },
    env: engineEnv,
    locks: [
      'container %s' % name,
    ],
    remove: if !skipRemove then '%s %s -f %s' % [e, rm, name],
  }
