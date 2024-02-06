package ssh_tunnel_util

import (
	"context"

	"github.com/daytonaio/daytona/config"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/ssh_tunnel"

	log "github.com/sirupsen/logrus"
)

func ForwardRemoteUnixSock(ctx context.Context, activeProfile config.Profile, localSock string, remoteSock string) (chan bool, chan error) {
	sshTun := ssh_tunnel.NewUnix(localSock, activeProfile.Hostname, remoteSock)

	sshTun.SetPort(activeProfile.Port)
	sshTun.SetUser(activeProfile.Auth.User)

	if activeProfile.Auth.Password != nil {
		sshTun.SetPassword(*activeProfile.Auth.Password)
	} else if activeProfile.Auth.PrivateKeyPath != nil {
		privateKeyPath, password, err := util.GetSshPrivateKeyPath(*activeProfile.Auth.PrivateKeyPath)
		if err != nil {
			log.Fatal(err)
		}
		if password != nil {
			sshTun.SetEncryptedKeyFile(privateKeyPath, *password)
		} else {
			sshTun.SetKeyFile(privateKeyPath)
		}
	}

	sshTun.SetTunneledConnState(func(tun *ssh_tunnel.SshTunnel, state *ssh_tunnel.TunneledConnectionState) {
		log.Debugf("%+v", state)
	})

	startedChann := make(chan bool, 1)

	sshTun.SetConnState(func(tun *ssh_tunnel.SshTunnel, state ssh_tunnel.ConnectionState) {
		switch state {
		case ssh_tunnel.StateStarting:
			log.Debugf("SSH Tunnel is Starting")
		case ssh_tunnel.StateStarted:
			log.Debugf("SSH Tunnel is Started")
			startedChann <- true
		case ssh_tunnel.StateStopped:
			log.Debugf("SSH Tunnel is Stopped")
		}
	})

	errChan := make(chan error)
	go func() {
		errChan <- sshTun.Start(ctx)
	}()

	return startedChann, errChan
}
