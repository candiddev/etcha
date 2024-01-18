// Create an apt PGP key with name from src under path.

function(name, path='/etc/apt/trusted.gpg.d', src)
  {
    change: 'etcha copy change %s - | gpg --dearmor > %s/%s.gpg' % [src, path, name],
    check: '[[ -f %s/%s.gpg ]]' % [path, name],
    id: 'apt_key %s' % name,
    remove: 'rm %s/%s.gpg' % [path, name],
  }
