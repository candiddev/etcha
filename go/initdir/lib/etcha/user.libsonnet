// Manage a local user.  Must specify gid, id and name.  Can optionally specify comment, password hash, paths to passwd/shadow, whether to remove the user, and a shell.

local line = import './line.libsonnet';

function(comment='', gid, hash='*', home='/bin', id, name, pathPasswd='/etc/passwd', pathShadow='/etc/shadow', remove=false, shell='/usr/sbin/nologin')
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
    line(match="'(?m)^%s:.*\n'" % name, path=pathPasswd, replaceChange="'%(name)s:x:%(id)s:%(gid)s:%(comment)s:%(home)s:%(shell)s\n'" % vars, replaceRemove=replaceRemove),
    line(match="'(?m)^%s:.*\n'" % name, path=pathShadow, replaceChange="'%(name)s:%(hash)s:'\"$( (grep %(name)s: /etc/shadow || echo \"::$(( $(date +%%s)/60/60/24))\") | cut -d: -f3)\"':0:99999:7:::\n'" % vars, replaceRemove=replaceRemove),
  ]
