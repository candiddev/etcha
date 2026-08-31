// Manage systemd network files and services.  Requires a map of files, where the key is the file name (like 00-enp1s0.network) and the contents of the file as the value.  Can optionally provide a separate path for the files, and whether to enable or restart systemd-networkd on changes.

local file = import './file.libsonnet';
local systemdUnit = import './systemdUnit.libsonnet';

function(enable=true, expand=false, files, path='/etc/systemd/network', restart=true, skipRemove=false, start=true)
  {
    id: 'systemdNetwork',
    commands: [
      [
        file(contents=files[f], expand=expand, group='systemd-network', mode='0600', owner='systemd-network', path='%s/%s' % [path, f], skipRemove=skipRemove) + {
          onChange: if restart then [
            'systemctl restart systemd-networkd',
          ],
        }

        for f in std.objectFields(files)
      ],
      systemdUnit(enable=enable, name='systemd-networkd', restart=restart, skipRemove=skipRemove, start=start),
    ],
  }
