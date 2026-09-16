// Manage a directory at path with a specific mode.

// If force is true, non-empty directories will be removed, otherwise this may error during removal if the directory is not empty.
function(absent=false, force=true, group='""', mode='0755', owner='""', path, skipRemove=false)

  local vars = {
    flags: '%(force)s -g %(group)s -o %(owner)s -p %(mode)s' % self,
    force: if force then '-f' else '',
    group: group,
    mode: if std.length(mode) == 3 then '0%s' % mode else mode,
    owner: owner,
    path: path,
    remove: 'etcha dir %(flags)s remove %(path)s' % vars,
  };

  {
    change: if absent then vars.remove else 'etcha dir %(flags)s change %(path)s' % vars,
    check: if absent then '! [[ -d %(path)s ]]' % vars else 'etcha dir %(flags)s check %(path)s' % vars,
    changeIgnore: true,
    id: 'dir %s' % path,
    locks: [
      'dir %s' % path,
    ],
    remove: if absent || skipRemove then '' else vars.remove,
  }
