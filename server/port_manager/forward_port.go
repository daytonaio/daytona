package port_manager

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os/exec"

	"github.com/daytonaio/daytona/internal/util"

	log "github.com/sirupsen/logrus"
)

func ForwardPort(workspaceName string, projectContainerName string, port ContainerPort) (PortForward, error) {
	if isPortAlreadyForwarded(workspaceName, projectContainerName, port) {
		return PortForward{}, errors.New("port is already forwarded")
	}

	var hostPort HostPort
	address := "127.0.0.1"

	if util.IsPortAvailable(uint16(port)) {
		hostPort = HostPort(port)
	} else {
		ephemeralPort, err := util.GetAvailableEphemeralPort()
		if err != nil {
			return PortForward{}, err
		}
		log.Debugf("Port %d is not available, using %d instead.", port, ephemeralPort)
		hostPort = HostPort(ephemeralPort)
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Handle incoming TCP connections
	go func() {
		var lc net.ListenConfig

		listener, err := lc.Listen(ctx, "tcp", fmt.Sprintf("%s:%d", address, hostPort))
		if err != nil {
			log.Error("Error creating TCP listener:", err)
			return
		}

		log.Debugf("Port forward listener started on port %d", hostPort)

		defer listener.Close()

		go func() {
			<-ctx.Done()
			listener.Close()
		}()

		for {
			tcpConn, err := listener.Accept()
			if err != nil {
				select {
				case <-ctx.Done():
					log.Debug("Port forward listener stopped")
					return
				default:
					log.Error("Error accepting connection:", err)
					return
				}
			}

			go handleTCPClient(ctx, tcpConn, projectContainerName, port)
		}
	}()

	log.Infof("Port %d forwarded to %d", port, hostPort)

	portForward := PortForward{
		HostPort:      hostPort,
		ContainerPort: port,
		ctx:           ctx,
		cancelFunc:    cancel,
	}

	_, ok := workspacePortForwards[workspaceName]
	if !ok {
		workspacePortForwards[workspaceName] = WorkspacePortForward{
			WorkspaceName:       workspaceName,
			ProjectPortForwards: make(map[string]PortForwards),
		}
	}

	_, ok = workspacePortForwards[workspaceName].ProjectPortForwards[projectContainerName]
	if !ok {
		workspacePortForwards[workspaceName].ProjectPortForwards[projectContainerName] = make(map[ContainerPort]PortForward)
	}

	workspacePortForwards[workspaceName].ProjectPortForwards[projectContainerName][port] = portForward

	return portForward, nil
}

func isPortAlreadyForwarded(workspaceName string, projectContainerName string, port ContainerPort) bool {
	workspacePortForward, ok := workspacePortForwards[workspaceName]
	if !ok {
		return false
	}

	projectPortForwards, ok := workspacePortForward.ProjectPortForwards[projectContainerName]
	if !ok {
		return false
	}

	_, ok = projectPortForwards[port]

	return ok
}

func handleTCPClient(ctx context.Context, tcpConn net.Conn, containerName string, containerPort ContainerPort) {
	defer tcpConn.Close()

	cmd := exec.CommandContext(ctx, "docker", []string{"exec", "-i", containerName, "daytona", "expose-port", fmt.Sprint(containerPort)}...)
	cmd.Stdin = tcpConn
	cmd.Stdout = tcpConn
	cmd.Stderr = tcpConn

	err := cmd.Start()
	if err != nil {
		log.Error("Error while running 'docker exec' command that exposes a port from inside the project container: ", err)
		return
	}

	err = cmd.Wait()
	if err != nil {
		select {
		case <-ctx.Done():
			return
		default:
			log.Error("Waiting failed for 'docker exec' command that exposes a port from inside the project container:", err)
			return
		}
	}
}
