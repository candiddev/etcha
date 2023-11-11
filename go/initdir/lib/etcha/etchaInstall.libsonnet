// Install the latest version of Etcha to dst.  Will cache the file to cacheDir for subsequent runs, if specified

local copy = import '../etcha/copy.libsonnet';

function(arch='amd64', cacheDir='', dst)
  local cache = if cacheDir == '' then dst else cacheDir + '/etcha';

  [
    {
      id: 'download Etcha to %s' % cache,
      check: '%s version 2>&1 | grep $(curl -sL https://github.com/candiddev/etcha/releases/latest/download/version) > /dev/null' % cache,
      change: 'curl -sL https://github.com/candiddev/etcha/releases/latest/download/etcha_linux_%s.tar.gz | tar -xOz etcha > %s && chmod 0755 %s' % [arch, cache, cache],
      remove: 'rm %s' % cache,
    },
    if cacheDir == '' then [] else [
      copy(dst=dst, src='%s/etcha' % cacheDir),
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
