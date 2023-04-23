local file = import './file.libsonnet';

function(chroot='', contents)
  local nftables = file(contents=|||
    #!/usr/sbin/nft -f

    flush ruleset

    %s
  ||| % contents, mode='0400', path=chroot + '/etc/nftables.conf') + {
    onChange: ['systemctl restart nftables'],
  };

  [
    nftables,
    {
      change: 'systemctl restart nftables',
      id: 'systemctl restart nftables',
    },
  ]
