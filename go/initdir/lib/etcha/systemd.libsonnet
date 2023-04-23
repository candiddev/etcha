local file = import './file.libsonnet';

function(chroot='', contents, dir='/etc/systemd/system', enable=true, name, restart=true)
  local systemctl = if chroot == '' then 'systemctl' else 'chroot %s systemctl' % chroot;

  local systemd = file(contents=contents, path='%s%s/%s' % [chroot, dir, name]) + {
    onChange: (
      if chroot == '' then [
        'systemctl daemon-reload %s' % name,
      ] else []
    ) + (
      if chroot == '' && restart then [
        'systemctl restart %s' % name,
      ] else []
    ),
  };

  [
    systemd,
  ] + (
    if chroot == '' then [] else [
      {
        id: 'systemctl daemon-reload %s' % name,
        change: '%s daemon-reload' % systemctl,
      },
    ]
  ) + (
    if enable then [
      {
        change: '%s enable %s' % [systemctl, name],
        check: '%s is-enabled %s' % [systemctl, name],
        id: 'systemctl enable %s' % name,
        remove: '%s disable --now %s' % [systemctl, name],
      },
    ] else []
  ) + (
    if chroot == '' && restart then [
      {
        change: '%s restart %s' % [systemctl, name],
        id: 'systemctl restart %s' % name,
        remove: '%s stop %s' % [systemctl, name],
      },
    ] else []
  )
