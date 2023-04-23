function(args='', dst, src)
  local vars = {
    args: args,
    dst: dst,
    src: src,
  };

  {
    change: |||
      mkdir -p %(src)s %(dst)s
      mount %(args)s %(src)s %(dst)s
    ||| % vars,
    check: 'mount | grep %s' % dst,
    id: 'mount %s' % dst,
    remove: |||
      umount %(dst)s
      rmdir %(src)s %(dst)s
    ||| % vars,
  }
