// Manage a password for a user.

function(hash, path='/etc/shadow', user='root')
  {
    change: |||
      # shellcheck disable=SC2016
      sed -i 's|^\(%s:\)[^:]*\(:.*\)$|\1%s\2|' %s
    ||| % [user, hash, path],
    check: 'grep "%s:%s:" %s' % [user, hash, path],
    id: '%s password hash' % user,
    remove: |||
      # shellcheck disable=SC2016
      sed -i 's|^\(%s:\)[^:]*\(:.*\)$|\1*\2|' %s
    ||| % [user, path],
  }
