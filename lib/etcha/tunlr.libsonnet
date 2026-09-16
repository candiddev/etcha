local copy = import './copy.libsonnet';
local file = import './file.libsonnet';
local systemdUnit = import './systemdUnit.libsonnet';

{
  // Install the latest version of Tunlr to dst.  Will cache the file to cacheDir for subsequent runs, if specified
  install: function(arch='amd64', cacheDir='', dst, os='linux', service='', skipRemove=false, src='https://tunlr.dev/releases')
    local cache = if cacheDir == '' then dst else cacheDir + '/tunlr';
    local s = if src == null || src == '' then 'https://tunlr.dev/releases' else src;
    local stop = if service == '' then '' else 'systemctl stop %s; ' % service;

    {
      commands: [
        {
          id: 'download Tunlr to %s' % cache,
          check: 'sha256sum %s | grep "$(etcha copy change %s/tunlr_%s_%s.sha256 -)" > /dev/null' % [cache, s, os, arch],
          change: '%setcha copy change %s/tunlr_%s_%s.gz - | gzip -d > %s && chmod +x %s' % [stop, s, os, arch, cache, cache],
          remove: if !skipRemove then 'rm %s' % cache,
        },
        if cacheDir != '' then [
          copy(dst=dst, skipRemove=skipRemove, src='%s/tunlr' % cacheDir),
        ],
      ],
      serial: true,
      id: 'tunlr install',
    },
  // Create a systemd service for Tunlr.
  service: function(binPath='/usr/bin/tunlr', client=true, configContents=null, configPath='/etc/tunlr.jsonnet', dir='/etc/systemd/system', enable=true, name='tunlr.service', reload=true, restart=true, serviceArgs='', skipRemove=false, start=true, user='tunlr')
    [
      if configContents != null then [
        file(contents=configContents, mode='0600', owner=user, path=configPath) + {
          onChange: 'systemctl restart tunlr.service',
        },
      ],
      systemdUnit(contents=|||
        [Unit]
        Description=Tunlr - Secure, simple, and scalable reverse proxy for dynamic environments
        Documentation=https://tunlr.dev
        After=network-online.target
        Wants=network-online.target

        [Service]
        ExecStart=%s -c %s %s
        Restart=always
        RestartSec=5
        User=%s
        %s

        [Install]
        WantedBy=multi-user.target
      ||| % [binPath, configPath, if client then 'client' else 'server', user, serviceArgs], dir=dir, enable=enable, name=name, reload=reload, restart=restart, skipRemove=skipRemove, start=start),
    ],
}
