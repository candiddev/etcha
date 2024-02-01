// Manage a local user.  Must specify id and name.  Can optionally specify members, paths to group and gshadow, and disable removal.

local line = import './line.libsonnet';

function(comment, gid, hash='*', home='/bin', id, name, pathPasswd='/etc/passwd', pathShadow='/etc/shadow', remove=false, shell='/usr/sbin/nologin')
  local vars = {
    comment: comment,
    gid: gid,
    hash: hash,
    home: home,
    id: id,
    name: name,
    shell: shell,
  };

  local replaceRemove = if remove then '""' else null;

  [
    {
      id: '%s password age' % name,
      check: '(grep %s: /etc/shadow || echo "::$(($(date +%%s)/60/60/24))") | cut -d: -f3' % name,
      envPrefix: 'age',
    },
    line(match='"(?m)^%s:.*"' % name, path=pathPasswd, replaceChange='"%(name)s:x:%(id)s:%(gid)s:%(comment)s:%(home)s:%(shell)s"' % vars, replaceRemove=replaceRemove),
    line(match='"(?m)^%s:.*"' % name, path=pathShadow, replaceChange='"%(name)s:%(hash)s:${age_CHECK_OUT}:0:99999:7:::"' % vars, replaceRemove=replaceRemove),
  ]
