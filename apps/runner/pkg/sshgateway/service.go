// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sshgateway

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"time"

	"github.com/daytonaio/runner/pkg/docker"
	"golang.org/x/crypto/ssh"
)

type Service struct {
	log          *slog.Logger
	dockerClient *docker.DockerClient
	port         int
}

func NewService(logger *slog.Logger, dockerClient *docker.DockerClient) *Service {
	port := GetSSHGatewayPort()

	service := &Service{
		log:          logger.With(slog.String("component", "ssh_gateway_service")),
		dockerClient: dockerClient,
		port:         port,
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

			s.log.WarnContext(ctx, "Public key authentication failed for sandbox", "sandboxID", sandboxId)
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

	s.log.InfoContext(ctx, "SSH Gateway listening on port", "port", s.port)

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			conn, err := listener.Accept()
			if err != nil {
				s.log.WarnContext(ctx, "Failed to accept incoming connection", "error", err)
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
		s.log.Warn("Failed to handshake", "error", err)
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
			s.log.Debug("Global request", "requestType", req.Type)
			// For now, just discard requests, but in a full implementation
			// these would be forwarded to the sandbox
			if req.WantReply {
				if err := req.Reply(false, []byte("not implemented")); err != nil {
					s.log.Warn("Failed to reply to global request", "error", err)
				}
			}
		}
	}()

	// connClosed signals per-channel watchers when the client connection dies.
	// A single Wait() goroutine per connection lets every channel select on it
	// instead of each blocking its own Wait() call.
	connClosed := make(chan struct{})
	go func() {
		serverConn.Wait() // nolint:errcheck
		close(connClosed)
	}()

	// Handle channels
	for newChannel := range chans {
		go s.handleChannel(newChannel, connClosed, sandboxId)
	}
}

// handleChannel handles an individual SSH channel
func (s *Service) handleChannel(newChannel ssh.NewChannel, connClosed <-chan struct{}, sandboxId string) {
	s.log.Debug("New channel", "channelType", newChannel.ChannelType(), "sandboxID", sandboxId)

	// Accept the channel from the client
	clientChannel, clientRequests, err := newChannel.Accept()
	if err != nil {
		s.log.Warn("Could not accept client channel", "error", err)
		return
	}
	defer clientChannel.Close()

	// Connect to the sandbox container via toolbox
	sandboxChannel, sandboxRequests, sandboxClient, err := s.connectToSandbox(sandboxId, newChannel.ChannelType(), newChannel.ExtraData())
	if err != nil {
		s.log.Warn("Could not connect to sandbox", "sandboxID", sandboxId, "error", err)
		clientChannel.Close()
		return
	}
	defer sandboxClient.Close()
	defer sandboxChannel.Close()

	// Forward requests from client to sandbox
	go func() {
		for req := range clientRequests {
			if req == nil {
				return
			}
			s.log.Debug("Client request", "requestType", req.Type, "sandboxID", sandboxId)

			ok, err := sandboxChannel.SendRequest(req.Type, req.WantReply, req.Payload)
			if req.WantReply {
				if err != nil {
					s.log.Warn("Failed to send request to sandbox", "requestType", req.Type, "sandboxID", sandboxId, "error", err)
					if replyErr := req.Reply(false, []byte(err.Error())); replyErr != nil {
						s.log.Warn("Failed to reply to client request", "error", replyErr)
					}
				} else {
					if replyErr := req.Reply(ok, nil); replyErr != nil {
						s.log.Warn("Failed to reply to client request", "error", replyErr)
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
			s.log.Debug("Sandbox request", "requestType", req.Type, "sandboxID", sandboxId)

			ok, err := clientChannel.SendRequest(req.Type, req.WantReply, req.Payload)
			if req.WantReply {
				if err != nil {
					s.log.Warn("Failed to send request to client", "requestType", req.Type, "sandboxID", sandboxId, "error", err)
					if replyErr := req.Reply(false, []byte(err.Error())); replyErr != nil {
						s.log.Warn("Failed to reply to sandbox request", "error", replyErr)
					}
				} else {
					if replyErr := req.Reply(ok, nil); replyErr != nil {
						s.log.Warn("Failed to reply to sandbox request", "error", replyErr)
					}
				}
			}
		}
	}()

	// Bidirectional data forwarding.
	// On client EOF, half-close the sandbox channel (MSG_CHANNEL_EOF via CloseWrite)
	// so the command in the sandbox can keep writing output and deliver its
	// exit-status (RFC 4254 half-close, e.g. `cat file | ssh host cmd`).
	go func() {
		_, err := io.Copy(sandboxChannel, clientChannel)
		if err != nil {
			s.log.Debug("Client to sandbox copy error", "error", err)
		}
		sandboxChannel.CloseWrite() // nolint:errcheck
	}()

	// The SSH library surfaces client transport failures as io.EOF on channel reads,
	// so the copy above cannot tell a clean stdin EOF from a SIGKILL/network drop, and
	// long-lived processes (VS Code Remote Server, JetBrains Gateway) silently ignore
	// MSG_CHANNEL_EOF. Watch the client connection itself instead: when it dies,
	// hard-close the sandbox channel — MSG_CHANNEL_CLOSE forces a reply, which unblocks
	// io.Copy(clientChannel, sandboxChannel) below and lets the sandbox shell receive
	// a hangup signal so orphaned processes exit cleanly. The watcher exits with the
	// session (sessionDone) so goroutines don't accumulate on long-lived connections
	// that open many channels.
	sessionDone := make(chan struct{})
	defer close(sessionDone)
	go func() {
		select {
		case <-connClosed:
			sandboxChannel.Close() // nolint:errcheck
		case <-sessionDone:
		}
	}()

	_, err = io.Copy(clientChannel, sandboxChannel)
	if err != nil {
		s.log.Debug("Sandbox to client copy error", "error", err)
	}

	s.log.Debug("Channel closed for sandbox", "sandboxID", sandboxId)
}

// connectToSandbox connects to the sandbox container via the toolbox
func (s *Service) connectToSandbox(sandboxId, channelType string, extraData []byte) (ssh.Channel, <-chan *ssh.Request, *ssh.Client, error) {
	// Get sandbox details via toolbox API
	sandboxDetails, err := s.getSandboxDetails(sandboxId)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get sandbox details: %w", err)
	}

	// Create SSH client config to connect to the sandbox
	clientConfig := &ssh.ClientConfig{
		User:            sandboxDetails.User,
		Auth:            []ssh.AuthMethod{ssh.Password("sandbox-ssh")}, // Use hardcoded password for sandbox auth
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         30 * time.Second,
	}

	// Connect to the sandbox container via toolbox
	sandboxClient, err := ssh.Dial("tcp", fmt.Sprintf("%s:22220", sandboxDetails.Hostname), clientConfig)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to connect to sandbox: %w", err)
	}

	// Open channel to the sandbox
	sandboxChannel, sandboxRequests, err := sandboxClient.OpenChannel(channelType, extraData)
	if err != nil {
		sandboxClient.Close()
		return nil, nil, nil, fmt.Errorf("failed to open channel to sandbox: %w", err)
	}

	// Return the real sandbox channel, requests, and client (caller must close client)
	return sandboxChannel, sandboxRequests, sandboxClient, nil
}

// getSandboxDetails gets sandbox information via docker client
func (s *Service) getSandboxDetails(sandboxId string) (*SandboxDetails, error) {
	// Get container details via docker client
	container, err := s.dockerClient.ContainerInspect(context.Background(), sandboxId)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect container %s: %w", sandboxId, err)
	}

	// Get container IP address
	containerIP := docker.GetContainerIpAddress(context.Background(), container)
	if containerIP == "" {
		return nil, fmt.Errorf("sandbox IP not found for %s", sandboxId)
	}

	return &SandboxDetails{
		User:     "daytona",
		Hostname: containerIP,
	}, nil
}

// SandboxDetails contains information about a sandbox
type SandboxDetails struct {
	User     string `json:"user"`
	Hostname string `json:"hostname"`
}
