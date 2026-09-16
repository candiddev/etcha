// Manage a symlink at dst pointing to src.

function(dst, skipRemove=false, src)
  {
    change: 'etcha link change %s %s' % [src, dst],
    changeIgnore: true,
    check: 'etcha link check %s %s' % [src, dst],
    id: 'symlink %s' % dst,
    locks: [
      'symlink %s' % dst,
    ],
    remove: if !skipRemove then 'etcha link remove %s %s' % [src, dst],
  }
