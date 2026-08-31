// Install a package using dnf, optionally ignoring weak dependencies.

function(name, options='', skipRemove=false, weakdepends=false)
  {
    change: ('dnf install %s ' % options) + (if weakdepends then '' else '--setopt=install_weak_deps=False ' + '-y %s' % name),
    check: 'rpm -q %s' % name,
    id: 'dnf %s' % name,
    locks: [
      'dnf',
    ],
    remove: if !skipRemove then 'dnf remove -y %s' % name,
  }
