local apt = import '../lib/etcha/apt.libsonnet';
local aptKey = import '../lib/etcha/aptKey.libsonnet';
local copy = import '../lib/etcha/copy.libsonnet';
local dir = import '../lib/etcha/dir.libsonnet';
local etchaInstall = import '../lib/etcha/etchaInstall.libsonnet';
local file = import '../lib/etcha/file.libsonnet';
local group = import '../lib/etcha/group.libsonnet';
local line = import '../lib/etcha/line.libsonnet';
local mount = import '../lib/etcha/mount.libsonnet';
local rotInstall = import '../lib/etcha/rotInstall.libsonnet';
local symlink = import '../lib/etcha/symlink.libsonnet';
local systemdNetwork = import '../lib/etcha/systemdNetwork.libsonnet';
local systemdUnit = import '../lib/etcha/systemdUnit.libsonnet';
local user = import '../lib/etcha/user.libsonnet';

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
    rotInstall(cacheDir='testdata/test', dst='testdata/rot'),
    rotInstall(dst='testdata/rot1'),
    file(contents=|||
      HOME=${HOMEDIR}
    |||, expand=true, path='testdata/home'),
    file(contents=|||
      HOME=/root
    |||, expand=false, path='testdata/home'),
    file(contents=|||
      root:x:0:0:root:/root:/bin/bash
      daemon:x:1:1:daemon:/usr/sbin:/usr/sbin/nologin
      bin:x:2:2:bin:/bin:/usr/sbin/nologin
      sys:x:3:3:sys:/dev:/usr/sbin/nologin
      sync:x:4:65534:sync:/bin:/bin/sync
      games:x:2010:60:samba:/home/games:/bin/false
    |||, group='daemon', owner='daemon', ignoreContents=true, path='testdata/passwd'),
    file(contents=|||
      root:*:18919:0:99999:7:::
      daemon:*:18907:0:99999:7:::
      bin:*:18907:0:99999:7:::
      sys:*:18907:0:99999:7:::
      sync:*:18907:0:99999:7:::
      games:*:18907:0:99999:7:::
    |||, group='daemon', owner='daemon', ignoreContents=true, path='testdata/shadow'),
    line(match='18919', path='testdata/shadow', replaceChange='18920', replaceRemove='18919'),
    user(comment='syncer', gid='444', hash='$y$j9T$hu21ZriPN8iCXixA/SRAI/$Wfw4stnUOjn3xULjD.7hhtn/mzvX/cwLePDPY9PBE6A', home='/sbin', id='5', name='sync', pathPasswd='testdata/passwd', pathShadow='testdata/shadow', remove=false, shell='/bin/bash'),
    user(comment='syncer', gid='444', hash='123', home='/sbin', id='5', name='syncer', pathPasswd='testdata/passwd', pathShadow='testdata/shadow', remove=false, shell='/bin/bash'),
    file(contents=|||
      root:x:0:
      daemon:x:1:
      bin:x:2:
      sys:x:3:
      adm:x:4:
      tty:x:5:
    |||, ignoreContents=true, path='testdata/group'),
    file(contents=|||
      root:*::
      daemon:*::
      bin:*::
      sys:*::
      adm:*::
      tty:*::
    |||, ignoreContents=true, path='testdata/gshadow'),
    group(id='4', members='user1,user2', name='adm', pathGroup='testdata/group', pathGshadow='testdata/gshadow'),
    group(id='1000', members='user3,user4', name='admins', pathGroup='testdata/group', pathGshadow='testdata/gshadow', remove=true),
    file(contents='hello', path='testdata/world'),
    file(path='testdata/touch'),
    dir(path='testdata/src'),
    mount(args='-o bind', dst='testdata/dst', src='testdata/src'),
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
    dir(path='testdata/network'),
    systemdNetwork(enable=false, files={
      'test.network': |||
        [Match]
        Name=!veth*
        Type=ether wlan

        [Network]
        DHCP=yes
      |||,
    }, path='testdata/network', restart=false),
  ],
  runExec: {
    allowOverride: true,
    command: std.native('getConfig')().exec.command,
    env: [
      'HOMEDIR=/root',
    ],
    envInherit: true,
    sudo: true,
  },
}
