local copy = import './copy.libsonnet';
local file = import './file.libsonnet';
local systemdUnit = import './systemdUnit.libsonnet';

{
  // Install the latest version of Etcha to dst.  Will cache the file to cacheDir for subsequent runs, if specified
  install: function(arch='amd64', cacheDir='', dst, os='linux', skipRemove=false, src='https://etcha.dev/releases')
    local cache = if cacheDir == '' then dst else cacheDir + '/etcha';
    local s = if src == null || src == '' then 'https://etcha.dev/releases' else src;

    {
      commands: [
        {
          id: 'download Etcha to %s' % cache,
          check: 'sha256sum %s | grep "$(etcha copy change %s/etcha_%s_%s.sha256 -)" > /dev/null' % [cache, s, os, arch],
          change: 'etcha copy change %s/etcha_%s_%s.gz - | gzip -d > %s && chmod +x %s' % [s, os, arch, cache, cache],
          locks: [
            'etcha copy',
          ],
          remove: if !skipRemove then 'rm %s' % cache,
        },
        if cacheDir != '' then [
          copy(dst=dst, skipRemove=skipRemove, src='%s/etcha' % cacheDir),
        ],
      ],
      serial: true,
      id: 'etcha install',
    },
  // Create a systemd service for Etcha.
  service: function(binPath='/usr/bin/etcha', configContents=null, configPath='/etc/etcha.jsonnet', dir='/etc/systemd/system', enable=true, name='etcha.service', reload=true, restart=true, serviceArgs='', skipRemove=false, start=true)
    {
      commands: [
        if configContents != null then [
          file(contents=configContents, mode='0600', path=configPath) + {
            onChange: 'systemctl restart etcha.service',
          },
        ],
        systemdUnit(contents=|||
          [Unit]
          Description=Etcha - infinite scale configuration management for distributed platforms
          Documentation=https://etcha.dev
          After=network-online.target
          Wants=network-online.target

          [Service]
          ExecStart=%s -c %s run
          Restart=always
          RestartSec=5
          %s

          [Install]
          WantedBy=multi-user.target
        ||| % [binPath, configPath, serviceArgs], dir=dir, enable=enable, name=name, reload=reload, restart=restart, skipRemove=skipRemove, start=start),
      ],
      id: 'etcha service',
    },
}
