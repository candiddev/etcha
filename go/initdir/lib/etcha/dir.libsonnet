// Manage a directory at path with a specific mode.

function(group='""', owner='""', path, mode='0755')

  local vars = {
    group: group,
    mode: if std.length(mode) == 3 then '0%s' % mode else mode,
    owner: owner,
    path: path,
  };

  {
    change: 'etcha dir change %(path)s %(mode)s %(owner)s %(group)s' % vars,
    check: 'etcha dir check %(path)s %(mode)s %(owner)s %(group)s' % vars,
    id: 'dir %s' % path,
    remove: 'etcha dir remove %(path)s %(mode)s %(owner)s %(group)s' % vars,
  }
