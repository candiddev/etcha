// Install package using the sysinfo.packageManager, optionally ignoring related packages.

local apt = import './apt.libsonnet';
local dnf = import './dnf.libsonnet';

local packageManager = std.native('getConfig')().vars.sysinfo.packageManager;

function(name, options='', related=false, skipRemove=false)
  if packageManager == 'apt' then apt(name, options, related, skipRemove) else if packageManager == 'dnf' then dnf(name, options, skipRemove, related)
