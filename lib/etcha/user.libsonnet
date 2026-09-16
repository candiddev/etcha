// Manage a local user.  Must specify gid, id and name.  Can optionally specify comment, password hash, paths to passwd/shadow, whether to remove the user, and a shell.

local line = import './line.libsonnet';

function(comment='', gid, hash='*', home='/bin', id, name, pathPasswd='/etc/passwd', pathShadow='/etc/shadow', shell='/usr/sbin/nologin', skipRemove=false)
  local vars = {
    comment: comment,
    gid: gid,
    hash: hash,
    home: home,
    id: id,
    name: name,
    shadow: pathShadow,
    shell: shell,
  };

  local replaceRemove = if !skipRemove then '""';

  {
    commands: [
      line(match="'^%s:.*'" % name, path=pathPasswd, replaceChange="'%(name)s:x:%(id)s:%(gid)s:%(comment)s:%(home)s:%(shell)s'" % vars, replaceRemove=replaceRemove),
      line(match="'^%s:.*'" % name, path=pathShadow, replaceChange="'%(name)s:%(hash)s:'\"$( (grep %(name)s: %(shadow)s || echo \"::$(( $(date +%%s)/60/60/24))\") | cut -d: -f3)\"':0:99999:7:::'" % vars, replaceRemove=replaceRemove),
    ],
    id: 'user %s' % name,
  }
