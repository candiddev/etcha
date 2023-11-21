// Manage a symlink at dst pointing to src.

function(dst, src)
  {
    change: 'ln -nsf %s %s' % [src, dst],
    check: '[[ $(readlink %s) == %s ]]' % [dst, src],
    id: 'symlink %s' % dst,
    remove: 'rm -f %s' % dst,
  }
