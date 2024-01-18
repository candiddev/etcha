// Manage a line at path.  Can set the contents, expand or not expand variables in the heredoc, set the owner and group, ignore content changes, and set the mode.

function(match, path, replaceChange, replaceRemove='""')
  local vars = {
    match: match,
    path: path,
    replaceChange: replaceChange,
    replaceRemove: replaceRemove,
  };

  {
    id: 'line %s' % path,
    check: 'etcha line check %(path)s %(match)s %(replaceChange)s' % vars,
    change: 'etcha line change %(path)s %(match)s %(replaceChange)s' % vars,
    remove: 'etcha line change %(path)s %(match)s %(replaceRemove)s' % vars,
  }
