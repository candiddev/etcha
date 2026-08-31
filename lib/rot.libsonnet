// Install the latest version of Rot to dst.  Will cache the file to cacheDir for subsequent runs, if specified

local copy = import './copy.libsonnet';

{
  install: function(arch='amd64', cacheDir='', dst, onChange=[], os='linux', skipRemove=false, src='https://rotx.dev/releases')
    local cache = if cacheDir == '' then dst else cacheDir + '/rot';
    local s = if src == null || src == '' then 'https://rotx.dev/releases' else src;

    {
      commands: [
        {
          id: 'download Rot to %s' % cache,
          check: 'sha256sum %s | grep "$(etcha copy change %s/rot_%s_%s.sha256 -)" > /dev/null' % [cache, s, os, arch],
          change: 'etcha copy change %s/rot_%s_%s.gz - | gzip -d > %s && chmod +x %s' % [s, os, arch, cache, cache],
          remove: if !skipRemove then 'rm %s' % cache,
          onChange: onChange,
        },
        if cacheDir != '' then [
          copy(dst=dst, skipRemove=skipRemove, src='%s/rot' % cacheDir) + {
            onChange: onChange,
          },
        ],
      ],
      serial: true,
      id: 'rot install',
    },
}
