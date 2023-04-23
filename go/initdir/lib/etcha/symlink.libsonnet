function(dst, src)
  {
    change: 'ln -sf %s %s' % [src, dst],
    check: '[[ $(readlink %s) == %s ]]' % [dst, src],
    id: 'symlink %s' % dst,
    remove: 'rm -f %s' % dst,
  }
