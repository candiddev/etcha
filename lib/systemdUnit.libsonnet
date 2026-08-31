// Manage a systemd unit with name and optionally enable/restart it.

local file = import './file.libsonnet';

function(contents='', dir='/etc/systemd/system', enable=true, name, reload=true, restart=true, skipRemove=false, start=true)
  local systemd = if contents != '' then file(contents=contents, path='%s/%s' % [dir, name], skipRemove=skipRemove) + {
    onChange: [
      if reload == true then 'systemctl daemon-reload %s' % name,
      if restart == true then 'systemctl restart %s' % name,
    ],
  };

  {
    id: 'systemdUnit %s' % name,
    commands: [
      systemd,
      if reload == true then {
        id: 'systemctl daemon-reload %s' % name,
        change: 'systemctl daemon-reload',
      },
      if enable != null then {
        change: 'systemctl %s --now %s' % [if enable then 'enable' else 'disable', name],
        check: '%s systemctl is-enabled %s' % [if enable then '' else '!', name],
        id: 'systemctl %s %s' % [if enable then 'enable' else 'disable', name],
        remove: if !skipRemove then 'systemctl %s --now %s' % [if enable then 'disable' else 'enable', name],
      },
      if start != null then {
        change: 'systemctl %s %s' % [if start then 'start' else 'stop', name],
        check: '%s systemctl status %s' % [if start then '' else '!', name],
        id: 'systemctl %s %s' % [if start then 'start' else 'stop', name],
      },
      if restart == true then {
        change: 'systemctl restart %s' % name,
        id: 'systemctl restart %s' % name,
        remove: if !skipRemove then 'systemctl stop %s' % name,
      },
    ],
  }
