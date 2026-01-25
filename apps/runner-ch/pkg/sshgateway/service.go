// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sshgateway

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/daytonaio/runner-ch/pkg/cloudhypervisor"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

type Service struct {
	chClient *cloudhypervisor.Client
	port     int
}

func NewService(chClient *cloudhypervisor.Client) *Service {
	port := GetSSHGatewayPort()

	service := &Service{
		chClient: chClient,
		port:     port,
	}

	return service
}

// GetPort returns the port the SSH gateway is configured to use
func (s *Service) GetPort() int {
	return s.port
}

// Start starts the SSH gateway server
func (s *Service) Start(ctx context.Context) error {
	// Get the public key from configuration
	publicKeyString, err := GetSSHPublicKey()
	if err != nil {
		return fmt.Errorf("failed to get SSH public key from config: %w", err)
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
		PublicKeyCallback: func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			// The username should be the sandbox ID
			sandboxId := conn.User()

			// Check if the provided key matches the configured public key
			if key.Type() == configPublicKey.Type() && bytes.Equal(key.Marshal(), configPublicKey.Marshal()) {
				return &ssh.Permissions{
					Extensions: map[string]string{
						"sandbox-id": sandboxId,
					},
				}, nil
			}

			log.Warnf("Public key authentication failed for sandbox %s", sandboxId)
			return nil, fmt.Errorf("authentication failed")
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
		log.Warnf("Failed to handshake: %v", err)
		return
	}
	defer serverConn.Close()

	sandboxId := serverConn.Permissions.Extensions["sandbox-id"]

	// Handle global requests
	go func() {
		for req := range reqs {
			if req == nil {
				continue
			}
			log.Debugf("Global request: %s", req.Type)
			// For now, just discard requests, but in a full implementation
			// these would be forwarded to the sandbox
			if req.WantReply {
				if err := req.Reply(false, []byte("not implemented")); err != nil {
					log.Warnf("Failed to reply to global request: %v", err)
				}
			}
		}
	}()

	// Handle channels
	for newChannel := range chans {
		go s.handleChannel(newChannel, sandboxId)
	}
}

// handleChannel handles an individual SSH channel
func (s *Service) handleChannel(newChannel ssh.NewChannel, sandboxId string) {
	log.Debugf("New channel: %s for sandbox: %s", newChannel.ChannelType(), sandboxId)

	// First, try to connect to the sandbox BEFORE accepting the channel
	// This allows us to reject with a proper error message if sandbox is unreachable
	sandboxChannel, sandboxRequests, err := s.connectToSandbox(sandboxId, newChannel.ChannelType(), newChannel.ExtraData())
	if err != nil {
		log.Warnf("Could not connect to sandbox %s: %v", sandboxId, err)
		newChannel.Reject(ssh.ConnectionFailed, fmt.Sprintf("could not connect to sandbox: %v", err))
		return
	}
	defer sandboxChannel.Close()

	// Now accept the client channel since we know sandbox is reachable
	clientChannel, clientRequests, err := newChannel.Accept()
	if err != nil {
		log.Warnf("Could not accept client channel: %v", err)
		return
	}
	defer clientChannel.Close()

	// Forward requests from client to sandbox
	go func() {
		for req := range clientRequests {
			if req == nil {
				return
			}
			log.Debugf("Client request: %s for sandbox %s", req.Type, sandboxId)

			ok, err := sandboxChannel.SendRequest(req.Type, req.WantReply, req.Payload)
			if req.WantReply {
				if err != nil {
					log.Warnf("Failed to send request to sandbox: %v", err)
					if replyErr := req.Reply(false, []byte(err.Error())); replyErr != nil {
						log.Warnf("Failed to reply to client request: %v", replyErr)
					}
				} else {
					if replyErr := req.Reply(ok, nil); replyErr != nil {
						log.Warnf("Failed to reply to client request: %v", replyErr)
					}
				}
			}
		}
	}()

	// Forward requests from sandbox to client
	go func() {
		for req := range sandboxRequests {
			if req == nil {
				return
			}
			log.Debugf("Sandbox request: %s for sandbox %s", req.Type, sandboxId)

			ok, err := clientChannel.SendRequest(req.Type, req.WantReply, req.Payload)
			if req.WantReply {
				if err != nil {
					log.Warnf("Failed to send request to client: %v", err)
					if replyErr := req.Reply(false, []byte(err.Error())); replyErr != nil {
						log.Warnf("Failed to reply to sandbox request: %v", replyErr)
					}
				} else {
					if replyErr := req.Reply(ok, nil); replyErr != nil {
						log.Warnf("Failed to reply to sandbox request: %v", replyErr)
					}
				}
			}
		}
	}()

	// Bidirectional data forwarding
	go func() {
		_, err := io.Copy(sandboxChannel, clientChannel)
		if err != nil {
			log.Debugf("Client to sandbox copy error: %v", err)
		}
	}()

	_, err = io.Copy(clientChannel, sandboxChannel)
	if err != nil {
		log.Debugf("Sandbox to client copy error: %v", err)
	}

	log.Debugf("Channel closed for sandbox: %s", sandboxId)
}

// connectToSandbox connects to the sandbox VM via the daemon SSH server
func (s *Service) connectToSandbox(sandboxId, channelType string, extraData []byte) (ssh.Channel, <-chan *ssh.Request, error) {
	// Get sandbox details
	sandboxDetails, err := s.getSandboxDetails(sandboxId)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get sandbox details: %w", err)
	}

	// Create SSH client config to connect to the sandbox
	clientConfig := &ssh.ClientConfig{
		User:            sandboxDetails.User,
		Auth:            []ssh.AuthMethod{ssh.Password("sandbox-ssh")}, // Use hardcoded password for sandbox auth
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         30 * time.Second,
	}

	sandboxAddr := fmt.Sprintf("%s:%d", sandboxDetails.Hostname, GetSandboxSSHPort())

	var sandboxClient *ssh.Client

	// Check if we need to tunnel through SSH (remote mode)
	if s.chClient.IsRemote() {
		sshHost := s.chClient.SSHHost
		log.Infof("SSH Gateway: tunneling to %s via %s", sandboxAddr, sshHost)

		// Dial through SSH tunnel
		tunnelConn, err := dialSSHTunnel(sshHost, s.chClient.SSHKeyPath, sandboxAddr)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create SSH tunnel: %w", err)
		}

		// Create SSH client over the tunnel connection
		sshConn, chans, reqs, err := ssh.NewClientConn(tunnelConn, sandboxAddr, clientConfig)
		if err != nil {
			tunnelConn.Close()
			return nil, nil, fmt.Errorf("failed to create SSH connection over tunnel: %w", err)
		}

		sandboxClient = ssh.NewClient(sshConn, chans, reqs)
	} else {
		// Direct connection (local/production mode)
		sandboxClient, err = ssh.Dial("tcp", sandboxAddr, clientConfig)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to connect to sandbox: %w", err)
		}
	}

	// Open channel to the sandbox
	sandboxChannel, sandboxRequests, err := sandboxClient.OpenChannel(channelType, extraData)
	if err != nil {
		sandboxClient.Close()
		return nil, nil, fmt.Errorf("failed to open channel to sandbox: %w", err)
	}

	// Return the real sandbox channel and requests
	return sandboxChannel, sandboxRequests, nil
}

// getSandboxDetails gets sandbox information via Cloud Hypervisor client
func (s *Service) getSandboxDetails(sandboxId string) (*SandboxDetails, error) {
	// Get sandbox info via CH client
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	sandboxInfo, err := s.chClient.GetSandboxInfo(ctx, sandboxId)
	if err != nil {
		return nil, fmt.Errorf("failed to get sandbox info for %s: %w", sandboxId, err)
	}

	// For remote mode, we need to use the namespace's external IP (10.0.{num}.1)
	// because the VM's internal IP (192.168.0.2) is only reachable from within the namespace.
	// Traffic to the namespace's external IP is DNATed to the VM.
	if s.chClient.IsRemote() {
		ns := s.chClient.GetNetNSPool().Get(sandboxId)
		if ns != nil && ns.ExternalIP != "" {
			log.Debugf("SSH Gateway: using namespace external IP %s for sandbox %s (remote mode)", ns.ExternalIP, sandboxId)
			return &SandboxDetails{
				User:     GetSandboxSSHUser(),
				Hostname: ns.ExternalIP,
			}, nil
		}
		log.Warnf("SSH Gateway: namespace not found for %s, falling back to VM IP", sandboxId)
	}

	// Get sandbox IP address (used for local mode or as fallback)
	sandboxIP := sandboxInfo.IpAddress
	if sandboxIP == "" {
		// Try to get from IP pool
		sandboxIP = s.chClient.GetIPPool().Get(sandboxId)
	}
	if sandboxIP == "" {
		// Last resort: try IP cache
		sandboxIP = cloudhypervisor.GetIPCache().Get(sandboxId)
	}
	if sandboxIP == "" {
		return nil, fmt.Errorf("sandbox IP not found for %s", sandboxId)
	}

	return &SandboxDetails{
		User:     GetSandboxSSHUser(),
		Hostname: sandboxIP,
	}, nil
}

// dialSSHTunnel creates a net.Conn that tunnels through SSH to reach a target address
func dialSSHTunnel(sshHost, sshKeyPath, targetAddr string) (net.Conn, error) {
	log.Debugf("SSH tunnel: dialing %s via %s", targetAddr, sshHost)

	// Use the SOCKS proxy transport to dial
	transport := cloudhypervisor.GetSSHTunnelTransport(sshHost, sshKeyPath)

	// Create a connection through the transport
	conn, err := transport.DialContext(context.Background(), "tcp", targetAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to dial through SSH tunnel: %w", err)
	}

	return conn, nil
}

// SandboxDetails contains information about a sandbox
type SandboxDetails struct {
	User     string `json:"user"`
	Hostname string `json:"hostname"`
}
