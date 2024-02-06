package ssh_tunnel_util

import (
	"context"
	"dagent/config"
	"dagent/internal/util"
	"dagent/ssh_tunnel"

	log "github.com/sirupsen/logrus"
)

func ForwardRemoteTcpPort(activeProfile config.Profile, targetPort uint16) (uint16, chan error) {
	hostPort := targetPort

	if !util.IsPortAvailable(targetPort) {
		ephemeralPort, err := util.GetAvailableEphemeralPort()
		if err != nil {
			log.Fatal(err)
		}
		hostPort = ephemeralPort
	}

	sshTun := ssh_tunnel.New(int(hostPort), activeProfile.Hostname, int(targetPort))

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

	sshTun.SetConnState(func(tun *ssh_tunnel.SshTunnel, state ssh_tunnel.ConnectionState) {
		switch state {
		case ssh_tunnel.StateStarting:
			log.Debugf("SSH Tunnel is Starting")
		case ssh_tunnel.StateStarted:
			log.Debugf("SSH Tunnel is Started")
		case ssh_tunnel.StateStopped:
			log.Debugf("SSH Tunnel is Stopped")
		}
	})

	errChan := make(chan error)
	go func() {
		errChan <- sshTun.Start(context.Background())
	}()

	return hostPort, errChan
}
