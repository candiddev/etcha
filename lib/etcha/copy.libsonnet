// Copy a file from src to dst.  Will use etcha to download the file or copy locally.  Can optionally provide a custom check function.

local n = import './native.libsonnet';

function(check='', dst, skipRemove=false, src)
  {
    id: 'copy %s' % dst,
    check: 'etcha copy check %s %s' % [src, dst],
    change: 'etcha copy change %s %s' % [src, dst],
    changeIgnore: true,
    locks: [
      'copy %s' % dst,
    ],
    remove: if !skipRemove then 'rm -rf ' + dst,
  }
