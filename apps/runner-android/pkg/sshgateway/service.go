// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sshgateway

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os/exec"
	"sync"
	"time"

	"github.com/daytonaio/runner-android/pkg/cuttlefish"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

type Service struct {
	cvdClient *cuttlefish.Client
	port      int
}

func NewService(cvdClient *cuttlefish.Client) *Service {
	port := GetSSHGatewayPort()

	service := &Service{
		cvdClient: cvdClient,
		port:      port,
	}

	return service
}

// GetPort returns the port the SSH gateway is configured to use
func (s *Service) GetPort() int {
	return s.port
}

// Start starts the SSH gateway server
// The SSH gateway enables users to create SSH tunnels to access ADB ports
// Handles connections from the main ssh-gateway for ADB port forwarding
// Main gateway connects with: User=sandboxId, Auth=PublicKey
func (s *Service) Start(ctx context.Context) error {
	// Get the public key from configuration (used to authenticate main gateway)
	publicKeyString, err := GetSSHPublicKey()
	if err != nil {
		log.Warnf("SSH Gateway: No public key configured, SSH gateway disabled: %v", err)
		<-ctx.Done()
		return nil
	}

	// Parse the public key from config
	configPublicKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(publicKeyString))
	if err != nil {
		return fmt.Errorf("failed to parse SSH public key from config: %w", err)
	}

	// Get the host key from configuration
	hostKey, err := GetSSHHostKey()
	if err != nil {
		return fmt.Errorf("failed to get SSH host key from config: %w", err)
	}

	serverConfig := &ssh.ServerConfig{
		// Public key authentication - main gateway connects with sandboxId as username
		PublicKeyCallback: func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			sandboxId := conn.User()

			// Validate the public key matches the configured key
			if key.Type() != configPublicKey.Type() || !bytes.Equal(key.Marshal(), configPublicKey.Marshal()) {
				log.Warnf("SSH Gateway: Public key authentication failed for %s", sandboxId)
				return nil, fmt.Errorf("authentication failed")
			}

			// Verify the sandbox exists
			if _, exists := s.cvdClient.GetInstance(sandboxId); !exists {
				log.Warnf("SSH Gateway: Sandbox %s not found", sandboxId)
				return nil, fmt.Errorf("sandbox not found")
			}

			log.Infof("SSH Gateway: Authenticated connection for sandbox %s", sandboxId)

			return &ssh.Permissions{
				Extensions: map[string]string{
					"sandbox-id": sandboxId,
				},
			}, nil
		},
		NoClientAuth: false,
	}

	serverConfig.AddHostKey(hostKey)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", s.port, err)
	}
	defer listener.Close()

	log.Infof("SSH Gateway listening on port %d", s.port)
	log.Info("SSH Gateway: Use SSH port forwarding to access ADB, e.g.:")
	log.Info("  ssh -L 5555:localhost:6520 -p %d <sandbox-id>@<runner-host>", s.port)

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			conn, err := listener.Accept()
			if err != nil {
				log.Warnf("Failed to accept incoming connection: %v", err)
				continue
			}

			go s.handleConnection(conn, serverConfig)
		}
	}
}

// handleConnection handles an individual SSH connection
func (s *Service) handleConnection(conn net.Conn, serverConfig *ssh.ServerConfig) {
	defer conn.Close()

	// Perform SSH handshake
	serverConn, chans, reqs, err := ssh.NewServerConn(conn, serverConfig)
	if err != nil {
		log.Warnf("SSH Gateway: Failed to handshake: %v", err)
		return
	}
	defer serverConn.Close()

	sandboxId := serverConn.Permissions.Extensions["sandbox-id"]
	log.Infof("SSH Gateway: New connection for sandbox %s from %s", sandboxId, conn.RemoteAddr())

	// Handle global requests (tcpip-forward, etc.)
	go s.handleGlobalRequests(reqs, sandboxId)

	// Handle channels
	for newChannel := range chans {
		go s.handleChannel(newChannel, sandboxId)
	}
}

// handleGlobalRequests handles global SSH requests like tcpip-forward
func (s *Service) handleGlobalRequests(reqs <-chan *ssh.Request, sandboxId string) {
	for req := range reqs {
		if req == nil {
			continue
		}
		log.Debugf("SSH Gateway: Global request type=%s for sandbox %s", req.Type, sandboxId)

		switch req.Type {
		case "keepalive@openssh.com":
			// Respond to keepalive requests
			if req.WantReply {
				req.Reply(true, nil)
			}
		default:
			// Reject unknown requests
			if req.WantReply {
				req.Reply(false, nil)
			}
		}
	}
}

// handleChannel handles an individual SSH channel
func (s *Service) handleChannel(newChannel ssh.NewChannel, sandboxId string) {
	channelType := newChannel.ChannelType()
	log.Debugf("SSH Gateway: New channel type=%s for sandbox %s", channelType, sandboxId)

	switch channelType {
	case "direct-tcpip":
		s.handleDirectTCPIP(newChannel, sandboxId)
	case "session":
		s.handleSessionChannel(newChannel, sandboxId)
	default:
		log.Warnf("SSH Gateway: Rejecting unsupported channel type: %s", channelType)
		newChannel.Reject(ssh.UnknownChannelType, fmt.Sprintf("unsupported channel type: %s", channelType))
	}
}

// directTCPIPPayload is the payload for direct-tcpip channel requests
// RFC 4254 Section 7.2
type directTCPIPPayload struct {
	DestHost string
	DestPort uint32
	OrigHost string
	OrigPort uint32
}

// handleDirectTCPIP handles direct-tcpip channel requests for port forwarding
// This is used when a user runs: ssh -L local:dest-host:dest-port sandbox@gateway
func (s *Service) handleDirectTCPIP(newChannel ssh.NewChannel, sandboxId string) {
	// Parse the destination from extra data
	var payload directTCPIPPayload
	if err := ssh.Unmarshal(newChannel.ExtraData(), &payload); err != nil {
		log.Warnf("SSH Gateway: Failed to parse direct-tcpip payload: %v", err)
		newChannel.Reject(ssh.ConnectionFailed, "failed to parse destination")
		return
	}

	destAddr := fmt.Sprintf("%s:%d", payload.DestHost, payload.DestPort)
	log.Infof("SSH Gateway: direct-tcpip request to %s for sandbox %s", destAddr, sandboxId)

	// Verify the sandbox exists
	_, exists := s.cvdClient.GetInstance(sandboxId)
	if !exists {
		log.Warnf("SSH Gateway: Sandbox %s not found", sandboxId)
		newChannel.Reject(ssh.ConnectionFailed, "sandbox not found")
		return
	}

	// Connect to the target address
	var targetConn net.Conn
	var err error

	if s.cvdClient.IsRemote() {
		// For remote mode, tunnel through SSH to the CVD host
		targetConn, err = s.dialThroughSSHTunnel(destAddr)
	} else {
		// For local mode, connect directly
		targetConn, err = net.DialTimeout("tcp", destAddr, 10*time.Second)
	}

	if err != nil {
		log.Warnf("SSH Gateway: Failed to connect to %s: %v", destAddr, err)
		newChannel.Reject(ssh.ConnectionFailed, fmt.Sprintf("failed to connect to %s: %v", destAddr, err))
		return
	}
	defer targetConn.Close()

	// Accept the channel
	channel, requests, err := newChannel.Accept()
	if err != nil {
		log.Warnf("SSH Gateway: Failed to accept channel: %v", err)
		return
	}
	defer channel.Close()

	// Discard channel requests
	go ssh.DiscardRequests(requests)

	// Bidirectional copy
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		io.Copy(targetConn, channel)
		// Signal EOF to target
		if tcpConn, ok := targetConn.(*net.TCPConn); ok {
			tcpConn.CloseWrite()
		}
	}()

	go func() {
		defer wg.Done()
		io.Copy(channel, targetConn)
		channel.CloseWrite()
	}()

	wg.Wait()
	log.Debugf("SSH Gateway: direct-tcpip connection to %s closed for sandbox %s", destAddr, sandboxId)
}

// dialThroughSSHTunnel creates a connection through the SSH tunnel to the CVD host
func (s *Service) dialThroughSSHTunnel(targetAddr string) (net.Conn, error) {
	sshHost := s.cvdClient.SSHHost
	sshKeyPath := s.cvdClient.SSHKeyPath

	// Parse host and port from targetAddr
	host, portStr, err := net.SplitHostPort(targetAddr)
	if err != nil {
		return nil, fmt.Errorf("invalid target address: %w", err)
	}

	// If connecting to localhost, we need to connect to the CVD host
	if host == "localhost" || host == "127.0.0.1" || host == "0.0.0.0" {
		// Connect through SSH port forwarding
		// Use ssh -W to proxy the connection through the SSH tunnel
		log.Debugf("SSH Gateway: Tunneling to %s via SSH host %s", targetAddr, sshHost)

		// Create a command that will proxy stdin/stdout through SSH
		cmd := exec.Command("ssh",
			"-i", sshKeyPath,
			"-o", "StrictHostKeyChecking=accept-new",
			"-o", "BatchMode=yes",
			"-o", "ConnectTimeout=10",
			"-W", fmt.Sprintf("localhost:%s", portStr),
			sshHost,
		)

		// Create pipes for communication
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
		}

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
		}

		if err := cmd.Start(); err != nil {
			return nil, fmt.Errorf("failed to start SSH tunnel: %w", err)
		}

		// Create a connection wrapper
		return &sshTunnelConn{
			cmd:    cmd,
			stdin:  stdin,
			stdout: stdout,
		}, nil
	}

	// For non-localhost addresses, connect directly (shouldn't happen for ADB)
	return net.DialTimeout("tcp", targetAddr, 10*time.Second)
}

// sshTunnelConn wraps an SSH -W command as a net.Conn
type sshTunnelConn struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
}

func (c *sshTunnelConn) Read(b []byte) (int, error) {
	return c.stdout.Read(b)
}

func (c *sshTunnelConn) Write(b []byte) (int, error) {
	return c.stdin.Write(b)
}

func (c *sshTunnelConn) Close() error {
	c.stdin.Close()
	c.stdout.Close()
	return c.cmd.Process.Kill()
}

func (c *sshTunnelConn) LocalAddr() net.Addr {
	return &net.TCPAddr{}
}

func (c *sshTunnelConn) RemoteAddr() net.Addr {
	return &net.TCPAddr{}
}

func (c *sshTunnelConn) SetDeadline(t time.Time) error {
	return nil
}

func (c *sshTunnelConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *sshTunnelConn) SetWriteDeadline(t time.Time) error {
	return nil
}

// handleSessionChannel handles session channel requests
// For Android, we provide an ADB shell wrapper
func (s *Service) handleSessionChannel(newChannel ssh.NewChannel, sandboxId string) {
	// Get the ADB serial for the sandbox
	serial, err := s.cvdClient.GetADBSerial(sandboxId)
	if err != nil {
		log.Warnf("SSH Gateway: Sandbox %s not found: %v", sandboxId, err)
		newChannel.Reject(ssh.ConnectionFailed, "sandbox not found")
		return
	}

	channel, requests, err := newChannel.Accept()
	if err != nil {
		log.Warnf("SSH Gateway: Failed to accept session channel: %v", err)
		return
	}
	defer channel.Close()

	// Process requests
	go func() {
		for req := range requests {
			if req == nil {
				continue
			}

			switch req.Type {
			case "shell", "exec":
				// Accept shell/exec requests
				if req.WantReply {
					req.Reply(true, nil)
				}
			case "pty-req":
				// Accept PTY request
				if req.WantReply {
					req.Reply(true, nil)
				}
			case "env":
				// Accept env settings
				if req.WantReply {
					req.Reply(true, nil)
				}
			default:
				if req.WantReply {
					req.Reply(false, nil)
				}
			}
		}
	}()

	// Start ADB shell
	var cmd *exec.Cmd
	adbPath := "adb" // Could be configurable

	if s.cvdClient.IsRemote() {
		// For remote mode, run ADB shell via SSH
		sshHost := s.cvdClient.SSHHost
		sshKeyPath := s.cvdClient.SSHKeyPath
		adbCmd := fmt.Sprintf("%s -s %s shell", adbPath, serial)
		cmd = exec.Command("ssh",
			"-i", sshKeyPath,
			"-o", "StrictHostKeyChecking=accept-new",
			"-o", "BatchMode=yes",
			"-t", "-t", // Force PTY allocation
			sshHost,
			adbCmd,
		)
	} else {
		cmd = exec.Command(adbPath, "-s", serial, "shell")
	}

	// Connect stdin/stdout/stderr
	cmd.Stdin = channel
	cmd.Stdout = channel
	cmd.Stderr = channel.Stderr()

	log.Infof("SSH Gateway: Starting ADB shell for sandbox %s (serial: %s)", sandboxId, serial)

	if err := cmd.Start(); err != nil {
		log.Errorf("SSH Gateway: Failed to start ADB shell: %v", err)
		channel.Write([]byte(fmt.Sprintf("Failed to start ADB shell: %v\r\n", err)))
		sendExitStatus(channel, 1)
		return
	}

	// Wait for command to complete
	err = cmd.Wait()
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			exitCode = 1
		}
	}

	sendExitStatus(channel, uint32(exitCode))
	log.Debugf("SSH Gateway: ADB shell for sandbox %s exited with code %d", sandboxId, exitCode)
}

// sendExitStatus sends the exit-status request to the channel
func sendExitStatus(channel ssh.Channel, exitCode uint32) {
	payload := make([]byte, 4)
	binary.BigEndian.PutUint32(payload, exitCode)
	channel.SendRequest("exit-status", false, payload)
}

// SandboxDetails contains information about a sandbox
type SandboxDetails struct {
	User      string `json:"user"`
	Hostname  string `json:"hostname"`
	ADBSerial string `json:"adbSerial"`
}

// getSandboxDetails gets sandbox information via Cuttlefish client
func (s *Service) getSandboxDetails(sandboxId string) (*SandboxDetails, error) {
	ctx := context.Background()

	sandboxInfo, err := s.cvdClient.GetSandboxInfo(ctx, sandboxId)
	if err != nil {
		return nil, fmt.Errorf("failed to get sandbox info for %s: %w", sandboxId, err)
	}

	return &SandboxDetails{
		User:      "shell", // ADB shell user
		ADBSerial: sandboxInfo.ADBSerial,
	}, nil
}

// GetADBConnectionInfo returns information needed to connect to ADB via SSH tunnel
func (s *Service) GetADBConnectionInfo(sandboxId string) (*ADBConnectionInfo, error) {
	info, exists := s.cvdClient.GetInstance(sandboxId)
	if !exists {
		return nil, fmt.Errorf("sandbox %s not found", sandboxId)
	}

	return &ADBConnectionInfo{
		SandboxId:   sandboxId,
		ADBPort:     info.ADBPort,
		ADBSerial:   info.ADBSerial,
		InstanceNum: info.InstanceNum,
		GatewayPort: s.port,
	}, nil
}

// ADBConnectionInfo contains information for ADB connection via SSH tunnel
type ADBConnectionInfo struct {
	SandboxId   string `json:"sandboxId"`
	ADBPort     int    `json:"adbPort"`
	ADBSerial   string `json:"adbSerial"`
	InstanceNum int    `json:"instanceNum"`
	GatewayPort int    `json:"gatewayPort"`
}
