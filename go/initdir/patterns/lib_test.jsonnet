local apt = import '../lib/etcha/apt.libsonnet';
local aptKey = import '../lib/etcha/aptKey.libsonnet';
local copy = import '../lib/etcha/copy.libsonnet';
local dir = import '../lib/etcha/dir.libsonnet';
local file = import '../lib/etcha/file.libsonnet';
local mount = import '../lib/etcha/mount.libsonnet';
local n = import '../lib/etcha/native.libsonnet';
local password = import '../lib/etcha/password.libsonnet';
local symlink = import '../lib/etcha/symlink.libsonnet';
local systemdUnit = import '../lib/etcha/systemdUnit.libsonnet';

local config = n.getConfig();

{
  run: [
    apt(package='fonts-fantasque-sans'),
    aptKey(name='bookworm', path=config.run.stateDir, src='https://ftp-master.debian.org/keys/archive-key-12.asc'),
    copy(src='https://candid.dev/sitemap.xml', dst=config.run.stateDir + '/sitemap.xml'),
    dir(mode='0644', path='%s/test' % config.run.stateDir),
    file(contents='root:*:19352:0:99999:7:::', group='daemon', ignoreContents=true, owner='daemon', path='%s/shadow' % config.run.stateDir),
    file(contents='hello', path='%s/world' % config.run.stateDir),
    file(path='%s/touch' % config.run.stateDir),
    mount(args='-o bind', dst='%s/dst' % config.run.stateDir, src='%s/src' % config.run.stateDir),
    password(hash='notahash', path='%s/shadow' % config.run.stateDir),
    symlink(src='%s/shadow' % config.run.stateDir, dst='%s/shadowsym' % config.run.stateDir),
    systemdUnit(contents=|||
      [Unit]
      Description=Test

      [Service]
      Type=oneshot
      ExecStart=/usr/bin/echo hello

      [Install]
      WantedBy=multi-user.target
    |||, enable=true, name='test.service', restart=true),
  ],
  runExec: {
    command: '/usr/bin/sudo /usr/bin/bash -e -o pipefail -c',
  },
}
