// Manage a line at path.  Can set the contents, owner and group, ignore content changes, and set the mode.  Will append the line to the end of the file during a change if nothing matches.

function(match, path, replaceChange, replaceRemove=null)
  local vars = {
    match: match,
    path: path,
    replaceChange: replaceChange,
    replaceRemove: replaceRemove,
  };

  {
    id: 'line %s %s' % [path, replaceChange],
    check: 'etcha line check %(path)s %(match)s %(replaceChange)s' % vars,
    change: 'etcha line change %(path)s %(match)s %(replaceChange)s' % vars,
  } + if replaceRemove == null then {} else {
    remove: 'etcha line change %(path)s %(match)s %(replaceRemove)s' % vars,
  }
