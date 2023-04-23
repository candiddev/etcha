function(hash, path='/etc/shadow')
  {
    change: |||
      sed -i 's/^\(root:\)[^:]*\(:.*\)$/\1%s\2/' %s
    ||| % [hash, path],
    check: 'grep "root:%s:" %s' % [hash, path],
    id: 'root password hash',
    remove: |||
      sed -i 's/^\(root:\)[^:]*\(:.*\)$/\1*\2/' %s
    ||| % path,
  }
