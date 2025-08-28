/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/daytonaio/apiclient"
	"golang.org/x/crypto/ssh"
)

const (
	defaultPort = 2222
	runnerPort  = 2220
)

type SSHGateway struct {
	port       int
	apiClient  *apiclient.APIClient
	hostKey    ssh.Signer
	privateKey ssh.Signer
	publicKey  ssh.PublicKey
}

func main() {
	port := getEnvInt("SSH_GATEWAY_PORT", defaultPort)
	apiURL := getEnv("API_URL", "http://localhost:3000")
	apiKey := getEnv("API_KEY", "")
	sshPkPath := getEnv("SSH_PK_PATH", "")

	if apiKey == "" {
		log.Fatal("API_KEY environment variable is required")
	}

	if sshPkPath == "" {
		log.Fatal("SSH_PK_PATH environment variable is required")
	}

	// Check if the private key file exists and is readable
	if _, err := os.Stat(sshPkPath); os.IsNotExist(err) {
		log.Fatalf("SSH private key file does not exist: %s", sshPkPath)
	}

	clientConfig := apiclient.NewConfiguration()
	clientConfig.Servers = apiclient.ServerConfigurations{
		{
			URL: apiURL,
		},
	}

	clientConfig.AddDefaultHeader("Authorization", "Bearer "+apiKey)

	apiClient := apiclient.NewAPIClient(clientConfig)

	apiClient.GetConfig().HTTPClient = &http.Client{
		Transport: http.DefaultTransport,
	}

	// Generate a temporary host key for the gateway
	hostKey, err := generateHostKey()
	if err != nil {
		log.Fatalf("Failed to generate host key: %v", err)
	}

	// Load the private key from file
	privateKey, err := loadPrivateKeyFromFile(sshPkPath)
	if err != nil {
		log.Fatalf("Failed to load private key from %s: %v", sshPkPath, err)
	}

	// Generate public key from private key
	publicKey := privateKey.PublicKey()

	gateway := &SSHGateway{
		port:       port,
		apiClient:  apiClient,
		hostKey:    hostKey,
		privateKey: privateKey,
		publicKey:  publicKey,
	}

	log.Printf("Private key loaded from %s", sshPkPath)
	log.Printf("Public key generated: %s", string(ssh.MarshalAuthorizedKey(publicKey)))

	log.Printf("Starting SSH Gateway on port %d", port)
	if err := gateway.Start(); err != nil {
		log.Fatalf("Failed to start SSH Gateway: %v", err)
	}
}

func (g *SSHGateway) Start() error {
	serverConfig := &ssh.ServerConfig{
		// Allow no client auth initially, we'll handle it in the PublicKeyCallback
		NoClientAuth: false,
		// Custom authentication that treats username as token
		PublicKeyCallback: func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			// Extract token from username
			token := conn.User()
			if token == "" {
				return nil, fmt.Errorf("no token provided")
			}

			log.Printf("Validating token: %s", token)

			// Validate the token using the API
			validation, _, err := g.apiClient.SandboxAPI.ValidateSshAccess(context.Background()).Token(token).Execute()
			if err != nil {
				log.Printf("Failed to validate SSH access: %v", err)
				return nil, fmt.Errorf("authentication failed")
			}

			if !validation.Valid {
				log.Printf("Invalid token: %s", token)
				return nil, fmt.Errorf("invalid or expired token")
			}

			if validation.RunnerId == nil {
				log.Printf("No runner ID returned for token: %s", token)
				return nil, fmt.Errorf("no runner available")
			}

			runnerDomain := ""
			if validation.RunnerDomain != nil {
				runnerDomain = *validation.RunnerDomain
			}

			log.Printf("Token validated, returning permissions for runner: %s", validation.RunnerId)

			// Return permissions with runner information
			// Note: We don't check sandbox state here - we'll do that in the connection handler
			// to allow us to close the connection with a proper error message
			return &ssh.Permissions{
				Extensions: map[string]string{
					"runner-id":     *validation.RunnerId,
					"runner-domain": runnerDomain,
					"token":         token,
					"sandbox-id":    validation.SandboxId,
				},
			}, nil
		},
		// Disable password authentication completely
		PasswordCallback: func(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
			return nil, fmt.Errorf("password authentication not allowed")
		},
		// Custom authentication handler
		AuthLogCallback: func(conn ssh.ConnMetadata, method string, err error) {
			if err != nil {
				log.Printf("Authentication failed for user %s: %v", conn.User(), err)
			}
		},
	}

	// Add host key
	serverConfig.AddHostKey(g.hostKey)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", g.port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", g.port, err)
	}
	defer listener.Close()

	log.Printf("SSH Gateway listening on port %d", g.port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept incoming connection: %v", err)
			continue
		}

		go g.handleConnection(conn, serverConfig)
	}
}

func (g *SSHGateway) handleConnection(conn net.Conn, serverConfig *ssh.ServerConfig) {
	defer conn.Close()

	// Perform SSH handshake
	serverConn, chans, reqs, err := ssh.NewServerConn(conn, serverConfig)
	if err != nil {
		log.Printf("Failed to handshake: %v", err)
		return
	}
	defer serverConn.Close()

	// Extract runner information from permissions
	runnerID := serverConn.Permissions.Extensions["runner-id"]
	runnerDomain := serverConn.Permissions.Extensions["runner-domain"]
	token := serverConn.Permissions.Extensions["token"]
	sandboxId := serverConn.Permissions.Extensions["sandbox-id"]

	log.Printf("SSH connection established for runner: %s", runnerID)

	// Check if the sandbox is started before proceeding
	if sandboxId != "" {
		log.Printf("Checking sandbox state for sandbox: %s", sandboxId)
		sandbox, _, err := g.apiClient.SandboxAPI.GetSandbox(context.Background(), sandboxId).Execute()
		if err != nil {
			log.Printf("Failed to get sandbox state for %s: %v", sandboxId, err)
			// Send error message to client and close connection
			g.sendErrorAndClose(conn, fmt.Sprintf("Failed to verify sandbox state: %v", err))
			return
		}

		if sandbox.State == nil || *sandbox.State != apiclient.SANDBOXSTATE_STARTED {
			state := "unknown"
			if sandbox.State != nil {
				state = string(*sandbox.State)
			}

			log.Printf("Sandbox %s is not started (state: %s), closing connection", sandboxId, state)
			g.sendErrorAndClose(conn, fmt.Sprintf("Sandbox is not started (state: %s). Please start the sandbox before attempting to connect.", state))
			return
		}

		log.Printf("Sandbox %s is started, allowing SSH connection", sandboxId)
	} else {
		log.Printf("No sandbox ID provided, proceeding with connection")
	}

	// Handle global requests
	go func() {
		for req := range reqs {
			if req == nil {
				continue
			}
			log.Printf("Global request: %s", req.Type)
			// For now, just discard requests
			if req.WantReply {
				req.Reply(false, []byte("not implemented"))
			}
		}
	}()

	// Handle channels
	for newChannel := range chans {
		go g.handleChannel(newChannel, runnerID, runnerDomain, token, sandboxId)
	}
}

func (g *SSHGateway) handleChannel(newChannel ssh.NewChannel, runnerID string, runnerDomain string, token string, sandboxId string) {
	log.Printf("New channel: %s for runner: %s", newChannel.ChannelType(), runnerID)

	// Accept the channel from the client
	clientChannel, clientRequests, err := newChannel.Accept()
	if err != nil {
		log.Printf("Could not accept client channel: %v", err)
		return
	}
	defer clientChannel.Close()

	// Use the loaded private key instead of fetching from API
	signer := g.privateKey

	// Connect to the runner's SSH gateway
	runnerConn, err := g.connectToRunner(sandboxId, runnerDomain, signer)
	if err != nil {
		log.Printf("Failed to connect to runner: %v", err)
		clientChannel.Close()
		return
	}
	defer runnerConn.Close()

	// Open channel to the runner
	runnerChannel, runnerRequests, err := runnerConn.OpenChannel(newChannel.ChannelType(), newChannel.ExtraData())
	if err != nil {
		log.Printf("Failed to open channel to runner: %v", err)
		return
	}
	defer runnerChannel.Close()

	// Forward requests from client to runner
	go func() {
		for req := range clientRequests {
			if req == nil {
				return
			}
			log.Printf("Client request: %s for runner %s", req.Type, runnerID)

			ok, err := runnerChannel.SendRequest(req.Type, req.WantReply, req.Payload)
			if req.WantReply {
				if err != nil {
					log.Printf("Failed to send request to runner: %v", err)
					req.Reply(false, []byte(err.Error()))
				} else {
					req.Reply(ok, nil)
				}
			}
		}
	}()

	// Forward requests from runner to client
	go func() {
		for req := range runnerRequests {
			if req == nil {
				return
			}
			log.Printf("Runner request: %s for runner %s", req.Type, runnerID)

			ok, err := clientChannel.SendRequest(req.Type, req.WantReply, req.Payload)
			if req.WantReply {
				if err != nil {
					log.Printf("Failed to send request to client: %v", err)
					req.Reply(false, []byte(err.Error()))
				} else {
					req.Reply(ok, nil)
				}
			}
		}
	}()

	// Bidirectional data forwarding
	go func() {
		_, err := io.Copy(runnerChannel, clientChannel)
		if err != nil {
			log.Printf("Client to runner copy error: %v", err)
		}
	}()

	_, err = io.Copy(clientChannel, runnerChannel)
	if err != nil {
		log.Printf("Runner to client copy error: %v", err)
	}

	log.Printf("Channel closed for runner: %s", runnerID)
}

func (g *SSHGateway) connectToRunner(sandboxId string, runnerDomain string, signer ssh.Signer) (*ssh.Client, error) {
	// Use runner domain if available, otherwise use localhost
	host := runnerDomain
	if host == "" {
		host = "localhost"
	}

	// Handle localdev case: if runnerDomain contains a port, remove it
	// For example: "localtest.me:3003" -> "localtest.me"
	if strings.Contains(host, "localtest.me") && strings.Contains(host, ":") {
		if idx := strings.Index(host, ":"); idx != -1 {
			host = host[:idx]
		}
	}

	// Ensure host is not empty after processing
	if host == "" {
		return nil, fmt.Errorf("invalid host: empty host after processing runner domain")
	}

	config := &ssh.ClientConfig{
		User: sandboxId, // Default username for sandbox
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         30 * time.Second,
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", host, runnerPort), config)
	if err != nil {
		return nil, fmt.Errorf("failed to dial runner: %w", err)
	}

	return client, nil
}

// sendErrorAndClose sends an error message to the client and closes the connection
func (g *SSHGateway) sendErrorAndClose(conn net.Conn, errorMessage string) {
	log.Printf("Sending error to client: %s", errorMessage)

	// For now, just close the connection
	// The client will see "Connection closed by remote host"
	// In a more sophisticated implementation, we could send a proper SSH disconnect message
	// but this requires restructuring the connection handling
	conn.Close()
}

func generateHostKey() (ssh.Signer, error) {
	// Generate a new RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key: %w", err)
	}

	// Create SSH signer
	signer, err := ssh.NewSignerFromKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH signer: %w", err)
	}

	return signer, nil
}

func parsePrivateKey(privateKeyPEM string) (ssh.Signer, error) {
	// First try to parse as OpenSSH format (newer format)
	signer, err := ssh.ParsePrivateKey([]byte(privateKeyPEM))
	if err == nil {
		return signer, nil
	}

	// If OpenSSH parsing fails, try PKCS1 format (older format)
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key (tried OpenSSH and PKCS1 formats): %w", err)
	}

	signer, err = ssh.NewSignerFromKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH signer: %w", err)
	}

	return signer, nil
}

// GetPublicKeyString returns the public key in authorized_keys format
func (g *SSHGateway) GetPublicKeyString() string {
	return string(ssh.MarshalAuthorizedKey(g.publicKey))
}

// GetPublicKey returns the SSH public key
func (g *SSHGateway) GetPublicKey() ssh.PublicKey {
	return g.publicKey
}

func loadPrivateKeyFromFile(filePath string) (ssh.Signer, error) {
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file %s: %w", filePath, err)
	}

	signer, err := parsePrivateKey(string(bytes))
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key from file %s: %w", filePath, err)
	}

	return signer, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
