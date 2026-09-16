// Mount src to dst using args.

function(dst, options='', skipRemove=false, src, type)
  local vars = {
    dst: dst,
    options: if options == '' then '' else '-o ' + options,
    src: src,
    type: if type == 'bind' then '-o bind' else '-t ' + type,
  };

  {
    change: |||
      mkdir -p %(dst)s
      mount %(type)s %(options)s %(src)s %(dst)s
    ||| % vars,
    check: 'mount | grep "%s %s"' % [dst, if type == 'bind' then '' else 'type ' + type],
    id: 'mount %s' % dst,
    locks: [
      'mount %s' % dst,
    ],
    remove: if !skipRemove then |||
      umount %(dst)s
      rmdir %(dst)s || true
    ||| % vars,
  }
