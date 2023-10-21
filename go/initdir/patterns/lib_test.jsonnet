local apt = import '../lib/etcha/apt.libsonnet';
local aptKey = import '../lib/etcha/aptKey.libsonnet';
local copy = import '../lib/etcha/copy.libsonnet';
local dir = import '../lib/etcha/dir.libsonnet';
local file = import '../lib/etcha/file.libsonnet';
local mount = import '../lib/etcha/mount.libsonnet';
local password = import '../lib/etcha/password.libsonnet';
local symlink = import '../lib/etcha/symlink.libsonnet';
local systemdUnit = import '../lib/etcha/systemdUnit.libsonnet';

{
  run: [
    apt(package='fonts-fantasque-sans'),
    aptKey(name='bookworm', path='testdata', src='https://ftp-master.debian.org/keys/archive-key-12.asc'),
    copy(src='https://candid.dev/sitemap.xml', dst='testdata/sitemap.xml'),
    copy(src='testdata/sitemap.xml', dst='testdata/sitemap2.xml'),
    dir(mode='0644', path='testdata/test'),
    file(contents='root:*:19352:0:99999:7:::', group='daemon', ignoreContents=true, owner='daemon', path='testdata/shadow'),
    file(contents='hello', path='testdata/world'),
    file(path='testdata/touch'),
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
    command: '/usr/bin/sudo /usr/bin/bash -e -o pipefail -c',
  },
}
