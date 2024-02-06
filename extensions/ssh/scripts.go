// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package ssh

var setupScript = `
#!/bin/bash

if [ ! -e "/usr/sbin/sshd" ]; then
	apt update -y && \
	apt install openssh-server -y
fi

mkdir -p /home/{{user}}/.ssh
touch /home/{{user}}/.ssh/authorized_keys
chmod 755 /home/{{user}}/.ssh
chmod 600 /home/{{user}}/.ssh/authorized_keys
echo "{{sshPublicKey}}" >> /home/{{user}}/.ssh/authorized_keys
chown -R {{user}}:{{user}} /home/{{user}}/.ssh

sudo mkdir -p /run/sshd
`

var startScript = `
#!/bin/bash
/usr/sbin/sshd -f /setup/ssh/sshd_config -o "SetEnv={{env}}"
`

var sshdConfig = `
Port 22
AddressFamily any
ListenAddress 0.0.0.0
ListenAddress ::

UsePAM yes

AllowAgentForwarding yes
AllowTcpForwarding yes
#GatewayPorts no
X11Forwarding yes
#X11DisplayOffset 10
#X11UseLocalhost yes
PermitTTY yes
PrintMotd no
PrintLastLog yes
TCPKeepAlive yes
PermitUserEnvironment yes
#Compression delayed
#ClientAliveInterval 0
#ClientAliveCountMax 3
UseDNS no
#PidFile /var/run/sshd.pid
#MaxStartups 10:30:100
#PermitTunnel no
#ChrootDirectory none
#VersionAddendum none

ChallengeResponseAuthentication no
KerberosAuthentication no
GSSAPIAuthentication no

# Don't read the user's ~/.rhosts and ~/.shosts files
IgnoreRhosts yes
# similar for protocol version 2
HostbasedAuthentication no

# Allow client to pass locale environment variables
AcceptEnv LANG LC_*

# override default of no subsystems
Subsystem       sftp    /usr/lib/openssh/sftp-server

# Example of overriding settings on a per-user basis
#Match User anoncvs
#       X11Forwarding no
#       AllowTcpForwarding no
#       PermitTTY no
#       ForceCommand cvs server

AllowUsers {{user}}

PasswordAuthentication yes
PermitEmptyPasswords yes

# PubkeyAcceptedAlgorithms does not work on older versions
PubkeyAcceptedKeyTypes +ssh-rsa
`
