package ssh_gateway

import (
	"errors"
	"net"
	"os"
	"strings"
	"time"

	"github.com/daytonaio/daytona/agent/config"

	gateway "github.com/daytonaio/ssh-gateway"
	"golang.org/x/crypto/ssh"
)

var sshGateway *gateway.SshGateway

func Start() error {
	if sshGateway == nil {
		os.RemoveAll("/tmp/daytona/ssh_gateway.sock")

		workspaceKey, err := config.GetWorkspaceKey()
		if err != nil {
			return err
		}
		sshGateway = &gateway.SshGateway{
			HostKey:       *workspaceKey,
			ListenNetwork: "unix",
			ListenAddress: "/tmp/daytona/ssh_gateway.sock",
			NoClientAuth:  true,
			ValidateNoClientAuthCallback: func(projectContainerName string) (*gateway.DestSshServer, error) {
				splited := strings.Split(projectContainerName, "~")
				if len(splited) != 2 {
					return nil, errors.New("invalid project container name")
				}

				// TODO: Implement
				// workspaceName := splited[0]
				// projectName := splited[1]

				// w, err := workspace.FindWorkspace(workspaceName)
				// if err != nil {
				// 	return nil, err
				// }

				// project, err := w.GetProject(projectName)
				// if err != nil {
				// 	return nil, err
				// }

				// containerInfo, err := project.GetContainerInfo()
				// if err != nil {
				// 	return nil, err
				// }
				containerIp := "1.1.1.1"

				_, err = net.DialTimeout("tcp", net.JoinHostPort(containerIp, "22"), time.Second)
				if err != nil {
					return nil, errors.New("can not connect to project container")
				}

				return &gateway.DestSshServer{
					Network: "tcp",
					Address: net.JoinHostPort(containerIp, "22"),
					Config: &ssh.ClientConfig{
						User: "daytona",
						Auth: []ssh.AuthMethod{
							ssh.PublicKeys(*workspaceKey),
						},
						HostKeyCallback: ssh.InsecureIgnoreHostKey(),
					},
				}, nil
			},
		}
	}

	return sshGateway.Start()
}
