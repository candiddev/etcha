// Install the latest version of Etcha to dst.  Will cache the file to cacheDir for subsequent runs, if specified

local copy = import './copy.libsonnet';

function(arch='amd64', cacheDir='', dst, onChange=[])
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
  ]
