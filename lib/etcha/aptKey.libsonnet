// Create an apt PGP key with name from src under path.

function(name, path='/etc/apt/trusted.gpg.d', skipRemove=false, src)
  {
    change: 'etcha copy change %s - | gpg --dearmor > %s/%s.gpg' % [src, path, name],
    check: '[[ -f %s/%s.gpg ]]' % [path, name],
    id: 'aptKey %s' % name,
    locks: [
      'aptKey %s' % name,
    ],
    remove: if !skipRemove then 'rm %s/%s.gpg' % [path, name],
  }
