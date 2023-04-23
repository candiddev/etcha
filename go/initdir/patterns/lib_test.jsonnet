local apt = import '../lib/etcha/apt.libsonnet';
local apt_key = import '../lib/etcha/apt_key.libsonnet';
local copy = import '../lib/etcha/copy.libsonnet';
local dir = import '../lib/etcha/dir.libsonnet';
local file = import '../lib/etcha/file.libsonnet';
local mount = import '../lib/etcha/mount.libsonnet';
local n = import '../lib/etcha/native.libsonnet';
local nftables = import '../lib/etcha/nftables.libsonnet';
local root = import '../lib/etcha/root.libsonnet';
local symlink = import '../lib/etcha/symlink.libsonnet';
local systemd = import '../lib/etcha/systemd.libsonnet';

local config = n.getConfig();

{
  exec: {
    command: '/usr/bin/sudo',
    flags: '/usr/bin/bash -e -o pipefail -c',
  },
  run: [
    apt(package='fonts-fantasque-sans'),
    apt_key(name='bookworm', path=config.run.stateDir, src='https://ftp-master.debian.org/keys/archive-key-12.asc'),
    copy(src='https://candid.dev/sitemap.xml', dst=config.run.stateDir + '/sitemap.xml'),
    dir(mode='0644', path='%s/etc/nftables' % config.run.stateDir),
    file(contents='root:*:19352:0:99999:7:::', group='daemon', ignoreContents=true, owner='daemon', path='%s/shadow' % config.run.stateDir),
    file(contents='hello', path='%s/world' % config.run.stateDir),
    mount(args='-o bind', dst='%s/dst' % config.run.stateDir, src='%s/src' % config.run.stateDir),
    nftables(chroot=config.run.stateDir, contents='nft list'),
    root(hash='notahash', path='%s/shadow' % config.run.stateDir),
    symlink(src='%s/shadow' % config.run.stateDir, dst='%s/shadowsym' % config.run.stateDir),
    systemd(chroot='/', contents=|||
      [Unit]
      Description=Test

      [Service]
      Type=oneshot
      ExecStart=/usr/bin/echo hello

      [Install]
      WantedBy=multi-user.target
    |||, enable=true, name='test.service', restart=true),
  ],
}
