// Manage a file at path.  Can set the contents, expand or not expand variables in the heredoc, set the owner and group, ignore content changes, and set the mode.

function(contents='', expand=false, group='root', ignoreContents=false, mode='0644', owner='root', path)
  local vars = {
    contents: contents,
    eof: if expand then 'EOF' else "'EOF'",
    group: group,
    mode: if std.length(mode) == 3 then '0%s' % mode else '%s' % mode,
    owner: owner,
    path: path,
  };

  local check = (
    if contents == '' || ignoreContents then '' else |||
      [[ $(sha1sum <(cat <<%(eof)s
      %(contents)s
      EOF
      ) | cut -d ' ' -f1) == $(sha1sum %(path)s | cut -d ' ' -f1) ]] &&
    ||| % vars
  ) + |||
    [[ -f %(path)s ]] && [[ $(stat -c "%%#a" %(path)s) == %(mode)s ]] && [[ $(stat -c "%%u" %(path)s) == $(getent passwd %(owner)s | cut -d: -f3 ) ]] && [[ $(stat -c "%%g" %(path)s) == $(getent group %(group)s | cut -d: -f3 ) ]]
  ||| % vars;

  local change = (
    if contents == '' then 'touch %(path)s\n' % vars else |||
      cat > %(path)s <<%(eof)s
      %(contents)s
      EOF
    ||| % vars
  ) + |||
    chmod %(mode)s %(path)s
    chown %(owner)s:%(group)s %(path)s
  ||| % vars;

  {
    id: 'file %s' % path,
    change: change,
    check: check,
    remove: 'rm %(path)s' % vars,
  }
