// Mount src to dst using args.

function(args='', dst, src)
  local vars = {
    args: args,
    dst: dst,
    src: src,
  };

  {
    change: |||
      mkdir -p %(dst)s
      mount %(args)s %(src)s %(dst)s
    ||| % vars,
    check: 'mount | grep %s' % dst,
    id: 'mount %s' % dst,
    remove: |||
      umount %(dst)s
      rmdir %(dst)s
    ||| % vars,
  }
