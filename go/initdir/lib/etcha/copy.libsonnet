// Copy a file from src to dst.  Will use wget to download the file.  Can optionally provide a custom check function.

local n = import './native.libsonnet';

function(check='', dst, src)
  {
    id: 'copy %s' % dst,
    check: 'etcha copy check %s %s' % [src, dst],
    change: 'etcha copy change %s %s' % [src, dst],
    remove: 'rm -rf ' + dst,
  }
