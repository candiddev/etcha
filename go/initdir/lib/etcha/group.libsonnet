// Manage a local group.  Must specify id and name.  Can optionally specify members, paths to group and gshadow, and enable removal.

local line = import './line.libsonnet';

function(id, members='', name, pathGroup='/etc/group', pathGshadow='/etc/gshadow', remove=false)
  local vars = {
    id: id,
    members: members,
    name: name,
  };

  local replaceRemove = if remove then '""';

  [
    line(match='"(?m)^%s:.*"' % name, path=pathGroup, replaceChange='"%(name)s:x:%(id)s:%(members)s"' % vars, replaceRemove=replaceRemove),
    line(match='"(?m)^%s:.*"' % name, path=pathGshadow, replaceChange='"%(name)s:|::%(members)s"' % vars, replaceRemove=replaceRemove),
  ]
