function(chroot='', mode='0755', path)

  local vars = {
    chroot: if chroot == '' then '' else 'chroot %s ' % chroot,
    mode: if std.length(mode) == 3 then '0%s' % mode else mode,
    path: path,
  };

  {
    change: |||
      %(chroot)smkdir -p %(path)s
      %(chroot)schmod %(mode)s %(path)s
    ||| % vars,
    check: '%(chroot)s[[ -d %(path)s ]] && [[ $(stat -c "%%#a" %(path)s) == %(mode)s ]]' % vars,
    id: 'dir %s' % path,
    remove: '%srm -rf %s' % [chroot, path],
  }
