package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/daytonaio/mcp/internal/apiclient"
	"github.com/daytonaio/mcp/pkg/servers/daytona"
	"github.com/daytonaio/mcp/pkg/servers/fs"
	"github.com/daytonaio/mcp/pkg/servers/git"
	"github.com/daytonaio/mcp/pkg/servers/sandbox"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	log "github.com/sirupsen/logrus"
)

type MCPServerConfig struct {
	Port            int
	TLSCertFilePath string
	TLSKeyFilePath  string
	ApiUrl          string
}

type MCPServer struct {
	port            int
	tlsCertFilePath string
	tlsKeyFilePath  string
	apiUrl          string
	httpServer      *http.Server
}

func NewMCPServer(config MCPServerConfig) *MCPServer {
	return &MCPServer{
		port:            config.Port,
		tlsCertFilePath: config.TLSCertFilePath,
		tlsKeyFilePath:  config.TLSKeyFilePath,
		apiUrl:          config.ApiUrl,
	}
}

func (s *MCPServer) Start() error {
	_, err := net.Dial("tcp", fmt.Sprintf(":%d", s.port))
	if err == nil {
		return fmt.Errorf("cannot start MCP server, port %d is already in use", s.port)
	}

	daytonaMcpHandler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		authHeader := r.Header.Get("Authorization")
		headers := map[string]string{
			"Authorization":    authHeader,
			"X-Daytona-Source": "daytona-mcp",
		}
		return daytona.NewDaytonaMCPServer(apiclient.NewApiClient(s.apiUrl, headers)).Server
	}, &mcp.StreamableHTTPOptions{JSONResponse: true})

	sandboxMcpHandler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		authHeader := r.Header.Get("Authorization")
		headers := map[string]string{
			"Authorization":    authHeader,
			"X-Daytona-Source": "daytona-sandbox-mcp",
		}
		return sandbox.NewDaytonaSandboxMCPServer(apiclient.NewApiClient(s.apiUrl, headers)).Server
	}, &mcp.StreamableHTTPOptions{JSONResponse: true})

	fsMcpHandler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		authHeader := r.Header.Get("Authorization")
		headers := map[string]string{
			"Authorization":    authHeader,
			"X-Daytona-Source": "daytona-fs-mcp",
		}
		return fs.NewDaytonaFileSystemMCPServer(apiclient.NewApiClient(s.apiUrl, headers)).Server
	}, &mcp.StreamableHTTPOptions{JSONResponse: true})

	gitMcpHandler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		authHeader := r.Header.Get("Authorization")
		headers := map[string]string{
			"Authorization":    authHeader,
			"X-Daytona-Source": "daytona-git-mcp",
		}
		return git.NewDaytonaGitMCPServer(apiclient.NewApiClient(s.apiUrl, headers)).Server
	}, &mcp.StreamableHTTPOptions{JSONResponse: true})

	httpMux := http.NewServeMux()
	httpMux.Handle("/mcp", daytonaMcpHandler)
	httpMux.Handle("/mcp/sandbox", sandboxMcpHandler)
	httpMux.Handle("/mcp/fs", fsMcpHandler)
	httpMux.Handle("/mcp/git", gitMcpHandler)

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
		log.Error(err)
	}
}
