function(chroot='', package, recommends=false)
  {
    change: (if chroot != '' then 'chroot ' else '' + 'apt-get install ') + (if recommends then '' else '--no-install-recommends ' + '-o DPkg::Options::="--force-confnew" -y %s' % package),
    check: (if chroot != '' then 'chroot ' else '') + 'dpkg -l %s' % package,
    id: 'apt %s' % package,
    remove: (if chroot != '' then 'chroot ' else '') + 'apt-get remove -y --purge %s' % package,
  }
