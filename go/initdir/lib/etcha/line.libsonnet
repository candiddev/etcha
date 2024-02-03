// Manage a line at path.  Can set the match regexp, the path to the file, the replacement strings for change, optionally for check (will use change if not specified) and remove (will not remove line if not specified).  Will append the line to the end of the file during a change if nothing matches.

function(match, path, replaceChange, replaceCheck=null, replaceRemove=null)
  local vars = {
    match: match,
    path: path,
    replaceChange: replaceChange,
    replaceCheck: if replaceCheck == null then replaceChange else replaceCheck,
    replaceRemove: replaceRemove,
  };

  {
    id: 'line %s %s' % [path, replaceChange],
    check: 'etcha line check %(path)s %(match)s %(replaceChange)s' % vars,
    change: 'etcha line change %(path)s %(match)s %(replaceChange)s' % vars,
  } + if replaceRemove == null then {} else {
    remove: 'etcha line change %(path)s %(match)s %(replaceRemove)s' % vars,
  }
