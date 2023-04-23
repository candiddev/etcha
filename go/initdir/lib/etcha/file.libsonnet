function(chroot='', contents='', expand=false, group='root', ignoreContents=false, mode='0644', owner='root', path)
  local vars = {
    chexec: if chroot == '' then '' else 'chroot %s' % chroot,
    chroot: chroot,
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
      ) | cut -d ' ' -f1) == $(sha1sum %(chroot)s%(path)s | cut -d ' ' -f1) ]] &&
    ||| % vars
  ) + |||
    [[ -f %(chroot)s%(path)s ]] && [[ $(stat -c "%%#a" %(chroot)s%(path)s) == %(mode)s ]] && [[ $(stat -c "%%u" %(chroot)s%(path)s) == $(%(chexec)sgetent passwd %(owner)s | cut -d: -f3 ) ]] && [[ $(stat -c "%%g" %(chroot)s%(path)s) == $(%(chexec)sgetent group %(group)s | cut -d: -f3 ) ]]
  ||| % vars;

  local change = (
    if contents == '' then '%(chexec)stouch %(path)s' % vars else |||
      cat > %(chroot)s%(path)s <<%(eof)s
      %(contents)s
      EOF
    ||| % vars
  ) + |||
    %(chexec)schmod %(mode)s %(path)s
    %(chexec)schown %(owner)s:%(group)s %(path)s
  ||| % vars;

  {
    id: 'file %s' % path,
    change: change,
    check: check,
    remove: '%(chexec)srm %(path)s' % vars,
  }
