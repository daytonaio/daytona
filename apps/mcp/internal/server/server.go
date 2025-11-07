// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/daytonaio/mcp/internal/apiclient"
	"github.com/daytonaio/mcp/internal/auth"
	"github.com/daytonaio/mcp/internal/constants"
	"github.com/daytonaio/mcp/pkg/servers/daytona"
	"github.com/daytonaio/mcp/pkg/servers/fs"
	"github.com/daytonaio/mcp/pkg/servers/git"
	"github.com/daytonaio/mcp/pkg/servers/sandbox"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type MCPServerConfig struct {
	Port            int
	TLSCertFilePath string
	TLSKeyFilePath  string
	ApiUrl          string
	Auth0Domain     string
	Auth0ClientId   string
	Auth0Audience   string
}

type MCPServer struct {
	port            int
	tlsCertFilePath string
	tlsKeyFilePath  string
	apiUrl          string
	httpServer      *http.Server
	auth0Domain     string
	auth0ClientId   string
	auth0Audience   string
}

func NewMCPServer(config MCPServerConfig) *MCPServer {
	return &MCPServer{
		port:            config.Port,
		tlsCertFilePath: config.TLSCertFilePath,
		tlsKeyFilePath:  config.TLSKeyFilePath,
		apiUrl:          config.ApiUrl,
		auth0Domain:     config.Auth0Domain,
		auth0ClientId:   config.Auth0ClientId,
		auth0Audience:   config.Auth0Audience,
	}
}

func (s *MCPServer) Start() error {
	_, err := net.Dial("tcp", fmt.Sprintf(":%d", s.port))
	if err == nil {
		return fmt.Errorf("cannot start MCP server, port %d is already in use", s.port)
	}

	// Create auth middleware
	authMiddleware := auth.CreateAuthMiddleware(s.auth0Domain, s.auth0ClientId, s.auth0Audience)

	daytonaMcpHandler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		apiClient := apiclient.NewApiClient(constants.DAYTONA_MCP_SOURCE, s.apiUrl, r.Header)
		return daytona.NewDaytonaMCPServer(apiClient).Server
	}, &mcp.StreamableHTTPOptions{JSONResponse: true})

	sandboxMcpHandler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		apiClient := apiclient.NewApiClient(constants.DAYTONA_SANDBOX_MCP_SOURCE, s.apiUrl, r.Header)
		return sandbox.NewDaytonaSandboxMCPServer(apiClient).Server
	}, &mcp.StreamableHTTPOptions{JSONResponse: true})

	fsMcpHandler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		apiClient := apiclient.NewApiClient(constants.DAYTONA_FS_MCP_SOURCE, s.apiUrl, r.Header)
		return fs.NewDaytonaFileSystemMCPServer(apiClient).Server
	}, &mcp.StreamableHTTPOptions{JSONResponse: true})

	gitMcpHandler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		apiClient := apiclient.NewApiClient(constants.DAYTONA_GIT_MCP_SOURCE, s.apiUrl, r.Header)
		return git.NewDaytonaGitMCPServer(apiClient).Server
	}, &mcp.StreamableHTTPOptions{JSONResponse: true})

	httpMux := http.NewServeMux()

	// Healthcheck (no auth required)
	httpMux.Handle("/daytona", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	// OAuth protected resource metadata endpoints (no auth required - these are discovery endpoints)
	httpMux.Handle("/daytona/.well-known/oauth-protected-resource", auth.HandleOAuthProtectedResource(s.auth0Domain))
	httpMux.Handle("/daytona/sandbox/.well-known/oauth-protected-resource", auth.HandleOAuthProtectedResource(s.auth0Domain))
	httpMux.Handle("/daytona/fs/.well-known/oauth-protected-resource", auth.HandleOAuthProtectedResource(s.auth0Domain))
	httpMux.Handle("/daytona/git/.well-known/oauth-protected-resource", auth.HandleOAuthProtectedResource(s.auth0Domain))

	// MCP Servers (protected with auth middleware)
	httpMux.Handle("/daytona/mcp", authMiddleware(daytonaMcpHandler))
	httpMux.Handle("/daytona/sandbox/mcp", authMiddleware(sandboxMcpHandler))
	httpMux.Handle("/daytona/fs/mcp", authMiddleware(fsMcpHandler))
	httpMux.Handle("/daytona/git/mcp", authMiddleware(gitMcpHandler))

	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: httpMux,
	}

	listener, err := net.Listen("tcp", s.httpServer.Addr)
	if err != nil {
		return err
	}

	errChan := make(chan error)
	go func() {
		if s.tlsCertFilePath != "" && s.tlsKeyFilePath != "" {
			errChan <- s.httpServer.ServeTLS(listener, s.tlsCertFilePath, s.tlsKeyFilePath)
		} else {
			errChan <- s.httpServer.Serve(listener)
		}
	}()

	return <-errChan
}

func (s *MCPServer) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.httpServer.Shutdown(ctx); err != nil {
		slog.Error("Failed to shutdown MCP server", "error", err)
	}
}
