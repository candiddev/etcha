// Install package using apt, optionally ignoring recommends.

function(name, options='', recommends=false, skipRemove=false)
  {
    change: ('apt-get install %s ' % options) + (if recommends then '' else '--no-install-recommends ' + '-o DPkg::Options::="--force-confnew" -y %s' % name),
    check: 'dpkg -L %s' % name,
    id: 'apt ' + name,
    locks: [
      'apt',
    ],
    remove: if skipRemove then '' else 'apt-get remove -y --purge %s' % name,
  }
