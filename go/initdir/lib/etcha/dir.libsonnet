// Manage a directory at path with a specific mode.

function(mode='0755', path)

  local vars = {
    mode: if std.length(mode) == 3 then '0%s' % mode else mode,
    path: path,
  };

  {
    change: |||
      mkdir -p %(path)s
      chmod %(mode)s %(path)s
    ||| % vars,
    check: '[[ -d %(path)s ]] && [[ $(stat -c "%%#a" %(path)s) == %(mode)s ]]' % vars,
    id: 'dir %s' % path,
    remove: 'rm -rf ' + path,
  }
