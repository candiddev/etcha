// Manage a file at path.  Can set the contents, optionally expand variables, the owner and group, ignore content changes, and set the mode.

function(absent=false, contents='', expand=false, group='""', ignoreContents=false, mode='0644', owner='""', path, skipRemove=false)
  local vars = {
    change: |||
      etcha file %(flags)s change %(path)s %(contentsChange)s << %(eof)s
      %(contents)s
    ||| % self,
    contents: contents + (if std.endsWith(contents, '\n') || contents == '' then '' else '\n') + 'EOF',
    contentsChange: if contents == '' then '' else '-',
    contentsCheck: if ignoreContents then '' else '-',
    eof: if expand then 'EOF' else "'EOF'",
    flags: '-g %(group)s -o %(owner)s -p %(mode)s' % self,
    group: group,
    mode: if std.length(mode) == 3 then '0%s' % mode else '%s' % mode,
    owner: owner,
    path: path,
    remove: 'etcha file %(flags)s remove %(path)s' % self,
  };

  {
    id: 'file %s%s' % [path, if absent then 'absent' else ''],
    check: if absent then '! [[ -f %(path)s ]]' % vars else |||
      etcha file %(flags)s check %(path)s %(contentsCheck)s << %(eof)s
      %(contents)s
    ||| % vars,
    change: if absent then vars.remove else vars.change,
    changeIgnore: true,
    locks: [
      'file %s' % path,
    ],
    remove: if !absent && !skipRemove then vars.remove,
  }
