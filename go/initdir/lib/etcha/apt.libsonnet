// Install package, optionally ignoring recommends.

function(package, recommends=false)
  {
    change: 'apt-get install ' + (if recommends then '' else '--no-install-recommends ' + '-o DPkg::Options::="--force-confnew" -y %s' % package),
    check: 'dpkg -l %s' % package,
    id: 'apt %s' % package,
    remove: 'apt-get remove -y --purge %s' % package,
  }
