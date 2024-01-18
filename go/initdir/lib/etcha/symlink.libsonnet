// Manage a symlink at dst pointing to src.

function(dst, src)
  {
    change: 'etcha link change %s %s' % [src, dst],
    check: 'etcha link check %s %s' % [src, dst],
    id: 'symlink %s' % dst,
    remove: 'etcha link remove %s %s' % [src, dst],
  }
