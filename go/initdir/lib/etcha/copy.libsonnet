// Copy a file from src to dst.  Will use wget to download the file.  Can optionally provide a custom check function.

local n = import './native.libsonnet';

function(check='', dst, src)
  {
    id: 'copy %s' % dst,
    change: if n.regexMatch('^http(s)?://', src) then
      'wget -O %s %s ' % [dst, src]
    else
      'cp -L --remove-destination %s %s' % [src, dst],
    check: if check != '' then
      check
    else if n.regexMatch('^http(s)?://', src) then
      '[[ $(wget -O - %s | sha1sum | cut -d " " -f1) == $(sha1sum %s | cut -d " " -f1) ]]' % [src, dst]
    else
      '[[ -f %(dst)s ]] && [[ $(sha1sum %(src)s | cut -d " " -f1) == $(sha1sum %(dst)s | cut -d " " -f1) ]]' % { dst: dst, src: src },
    remove: 'rm -rf ' + dst,
  }
