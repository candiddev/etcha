// Manage a systemd unit with name and optionally enable/restart it.

local file = import './file.libsonnet';

function(contents='', dir='/etc/systemd/system', enable=true, name, reload=true, restart=true)
  local systemd = if contents == '' then [] else [
    file(contents=contents, path='%s/%s' % [dir, name]) + {
      onChange: (
        if reload then [
          'systemctl daemon-reload %s' % name,
        ] else []
      ) + (
        if restart then [
          'systemctl restart %s' % name,
        ] else []
      ),
    },
  ];

  systemd + (
    if reload then [
      {
        id: 'systemctl daemon-reload %s' % name,
        change: 'systemctl daemon-reload',
      },
    ] else []
  ) + (
    if enable then [
      {
        change: 'systemctl enable --now %s' % name,
        check: 'systemctl is-enabled %s' % name,
        id: 'systemctl enable %s' % name,
        remove: 'systemctl disable --now %s' % name,
      },
    ] else []
  ) + (
    if restart then [
      {
        change: 'systemctl restart %s' % name,
        id: 'systemctl restart %s' % name,
        remove: 'systemctl stop %s' % name,
      },
    ] else []
  )
