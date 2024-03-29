// Install the latest version of Etcha to dst.  Will cache the file to cacheDir for subsequent runs, if specified

local copy = import './copy.libsonnet';
local systemdUnit = import './systemdUnit.libsonnet';

{
  install: function(arch='amd64', cacheDir='', dst, onChange=[])
    local cache = if cacheDir == '' then dst else cacheDir + '/etcha';

    [
      {
        id: 'download Etcha to %s' % cache,
        check: '%s version 2>&1 | grep "$(etcha copy change https://github.com/candiddev/etcha/releases/latest/download/version -)" > /dev/null' % cache,
        change: 'etcha copy change https://github.com/candiddev/etcha/releases/latest/download/etcha_linux_%s.tar.gz - | tar -xOz etcha > %s && chmod +x %s' % [arch, cache, cache],
        remove: 'rm %s' % cache,
        onChange: onChange,
      },
      if cacheDir == '' then [] else [
        copy(dst=dst, src='%s/etcha' % cacheDir) + {
          onChange: onChange,
        },
      ],
      {
        id: 'etcha version',
        always: true,
        change: '%s version' % dst,
        onChange: [
          'etcha:buildManifest',
        ],
      },
    ],
  service: function(config='/etc/etcha.jsonnet', dir='/etc/systemd/system', enable=true, name='etcha', reload=true, restart=true)
    {
      id: 'etcha service',
      commands: [
        systemdUnit(contents=|||
          [Unit]
          Description=Etcha - infinite scale configuration management for distributed platforms
          Documentation=https://etcha.dev
          After=network-online.target
          Wants=network-online.target

          [Service]
          ExecStart=/usr/bin/etcha -c %s run
          Restart=always
          RestartSec=5

          [Install]
          WantedBy=multi-user.target
        ||| % config, dir=dir, enable=enable, name=name, reload=reload, restart=restart),
      ],
    },
}
