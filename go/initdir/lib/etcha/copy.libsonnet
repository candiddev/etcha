local n = import './native.libsonnet';

function(check='', dst, src)

  {
    id: 'copy %s' % dst,
    change: if n.regexMatch('^http(s)?://', src) then
      'curl -L %s -o %s' % [src, dst]
    else
      'cp -L --remove-estination %s %s' % [src, dst],
    check: if check != '' then
      check
    else if n.regexMatch('^http(s)?://', src) then
      '[[ $(curl -L %s | sha1sum | cut -d " " -f1) == $(sha1sum %s | cut -d " " -f1) ]]' % [src, dst]
    else
      '[[ -f %(dst)s ]] && [[ $(sha1sum %(src)s | cut -d " " -f1) == $(sha1sum %(dst)s | cut -d " " -f1 ]]' % { dst: dst, src: src },
    remove: 'rm -rf ' + dst,
  }
