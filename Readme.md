## Install
```
GOPATH=/tmp/gopath go get -u github.com/vadv/knife-sh
sudo install /tmp/gopath/bin/knife-sh /usr/local/bin/knife-sh
```

## Howto use
```
Help:
knife-sh HOSTS COMMAND (options)
    HOSTS is 'host1 host2' or /path/to/ip.txt or CHEF:QUERY or - for STDIN
    -C, --concurrency NUM   The number of concurrent connections, default: 100
    -x, --ssh-user USERNAME The ssh username, default: vadv
    -i, --identity-file IDENTITY_FILE,      default: /home/vadv/.ssh/id_rsa
    -t, --ssh-timeout SSH TIMEOUT(s)        The ssh connection timeout, default: 10
    -e, --execution-timeout EXECUTION TIMEOUT(s)    The command execution timeout, default: 0
    -s, --stop-on-first-error false|true    Exit on first error, default: false
    -c, --copy-file Copy file before execution, format: 'local-source:remote-destination'
    -c, --chef-client CHEF CLIENT   Chef client name, default: user
        --chef-certificate CERT FILE     Path to client certificate, default: /home/user/.chef/user.pem
    -a, --chef-attribute ATTRIBUTE  Chef attribute for connect, default: fqdn
    -u, --chef-url URL      Chef server url, default: https://chef/organizations/org/

You can also specify the long-attributes in the config file: ~/.knife-sh.rc in format like ~/.ssh/config ('key = value' or 'key value')
```

## Example config
```
chef-certificate /Users/username/.chef/admin.pem
chef-client admin
chef-attribute ipaddress
ssh-user root

# proxy option
jump-ssh-host 192.168.0.1:22
jump-ssh-user admin
jump-ssh-key <path to identity key >
```
