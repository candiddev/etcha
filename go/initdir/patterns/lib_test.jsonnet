local apt = import '../lib/etcha/apt.libsonnet';
local aptKey = import '../lib/etcha/aptKey.libsonnet';
local copy = import '../lib/etcha/copy.libsonnet';
local dir = import '../lib/etcha/dir.libsonnet';
local etchaInstall = import '../lib/etcha/etchaInstall.libsonnet';
local file = import '../lib/etcha/file.libsonnet';
local line = import '../lib/etcha/line.libsonnet';
local mount = import '../lib/etcha/mount.libsonnet';
local password = import '../lib/etcha/password.libsonnet';
local symlink = import '../lib/etcha/symlink.libsonnet';
local systemdUnit = import '../lib/etcha/systemdUnit.libsonnet';

{
  run: [
    {
      id: 'build etcha',
      always: true,
      change: 'go build -o etcha .. && sudo mv etcha /usr/local/bin/etcha',
      remove: 'sudo rm /usr/local/bin/etcha',
      exec: {
        command: '/usr/bin/bash -e -c',
        envInherit: true,
      },
    },
    apt(package='fonts-rocknroll'),
    dir(path='testdata'),
    aptKey(name='bookworm', path='testdata', src='https://ftp-master.debian.org/keys/archive-key-12.asc'),
    copy(src='https://candid.dev/sitemap.xml', dst='testdata/sitemap.xml'),
    copy(src='testdata/sitemap.xml', dst='testdata/sitemap2.xml'),
    dir(group='daemon', mode='0700', owner='daemon', path='testdata/test'),
    etchaInstall(cacheDir='testdata/test', dst='testdata/etcha'),
    etchaInstall(dst='testdata/etcha1'),
    file(contents='root:*:19352:0:99999:7:::', group='daemon', owner='daemon', ignoreContents=true, path='testdata/shadow'),
    line(match='19352', path='testdata/shadow', replaceChange='19352!', replaceRemove='19352'),
    file(contents='hello', path='testdata/world'),
    file(path='testdata/touch'),
    dir(path='testdata/src'),
    mount(args='-o bind', dst='testdata/dst', src='testdata/src'),
    password(hash='notahash', path='testdata/shadow'),
    symlink(src='testdata/shadow', dst='testdata/shadowsym'),
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
    allowOverride: true,
    command: '/usr/bin/sudo /usr/bin/bash -e -o pipefail -c',
    envInherit: true,
  },
}
