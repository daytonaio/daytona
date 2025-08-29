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

	"github.com/daytonaio/runner/pkg/docker"
	"github.com/docker/docker/api/types"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

type Service struct {
	dockerClient *docker.DockerClient
	port         int
}

func NewService(dockerClient *docker.DockerClient) *Service {
	port := GetSSHGatewayPort()

	service := &Service{
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

	// Accept the channel from the client
	clientChannel, clientRequests, err := newChannel.Accept()
	if err != nil {
		log.Warnf("Could not accept client channel: %v", err)
		return
	}
	defer clientChannel.Close()

	// Connect to the sandbox container via toolbox
	sandboxChannel, sandboxRequests, err := s.connectToSandbox(sandboxId, newChannel.ChannelType(), newChannel.ExtraData())
	if err != nil {
		log.Warnf("Could not connect to sandbox %s: %v", sandboxId, err)
		clientChannel.Close()
		return
	}
	defer sandboxChannel.Close()

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

// connectToSandbox connects to the sandbox container via the toolbox
func (s *Service) connectToSandbox(sandboxId, channelType string, extraData []byte) (ssh.Channel, <-chan *ssh.Request, error) {
	// Get sandbox details via toolbox API
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

	// Connect to the sandbox container via toolbox
	sandboxClient, err := ssh.Dial("tcp", fmt.Sprintf("%s:22220", sandboxDetails.Hostname), clientConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to sandbox: %w", err)
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

// getSandboxDetails gets sandbox information via docker client
func (s *Service) getSandboxDetails(sandboxId string) (*SandboxDetails, error) {
	// Get container details via docker client
	container, err := s.dockerClient.ContainerInspect(context.Background(), sandboxId)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect container %s: %w", sandboxId, err)
	}

	// Get container IP address
	containerIP, err := s.getContainerIP(&container)
	if err != nil {
		return nil, fmt.Errorf("failed to get container IP for %s: %w", sandboxId, err)
	}

	return &SandboxDetails{
		User:     "daytona",
		Hostname: containerIP,
	}, nil
}

// getContainerIP extracts the IP address from a container
func (s *Service) getContainerIP(container *types.ContainerJSON) (string, error) {
	for _, network := range container.NetworkSettings.Networks {
		return network.IPAddress, nil
	}
	return "", fmt.Errorf("no IP address found. Is the Sandbox started?")
}

// SandboxDetails contains information about a sandbox
type SandboxDetails struct {
	User     string `json:"user"`
	Hostname string `json:"hostname"`
}
