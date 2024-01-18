// Manage a file at path.  Can set the contents, expand or not expand variables in the heredoc, set the owner and group, ignore content changes, and set the mode.

function(contents='', expand=false, group='""', ignoreContents=false, mode='0644', owner='""', path)
  local vars = {
    contents: contents,
    contentsChange: if contents == '' then '' else '-',
    contentsCheck: if ignoreContents then '' else '-',
    eof: if expand then 'EOF' else "'EOF'",
    group: group,
    mode: if std.length(mode) == 3 then '0%s' % mode else '%s' % mode,
    owner: owner,
    path: path,
  };

  {
    id: 'file %s' % path,
    check: 'etcha file check %(path)s %(mode)s %(owner)s %(group)s %(contentsCheck)s' % vars,
    change: 'etcha file change %(path)s %(mode)s %(owner)s %(group)s %(contentsChange)s' % vars,
    stdin: contents,
    remove: 'etcha file remove %(path)s %(mode)s %(owner)s %(group)s' % vars,
  }
