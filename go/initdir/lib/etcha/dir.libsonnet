// Manage a directory at path with a specific mode.

function(group='""', owner='""', path, mode='0755')

  local vars = {
    flags: '-g %(group)s -o %(owner)s -p %(mode)s' % self,
    group: group,
    mode: if std.length(mode) == 3 then '0%s' % mode else mode,
    owner: owner,
    path: path,
  };

  {
    change: 'etcha dir %(flags)s change %(path)s' % vars,
    check: 'etcha dir %(flags)s check %(path)s' % vars,
    id: 'dir %s' % path,
    remove: 'etcha dir %(flags)s remove %(path)s' % vars,
  }
