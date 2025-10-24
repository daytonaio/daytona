// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package mcp

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"

	"github.com/daytonaio/daytona/cli/internal/mcp/daytona"
	"github.com/daytonaio/daytona/cli/internal/mcp/fs"
	"github.com/daytonaio/daytona/cli/internal/mcp/git"
	"github.com/daytonaio/daytona/cli/internal/mcp/sandbox"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

var transportFlag string
var portFlag int
var tlsCertFilePathFlag string
var tlsKeyFilePathFlag string

var ServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve any Daytona MCP Server (empty string for daytona code execution MCP, 'sandbox' for Sandbox actions MCP, 'fs' for Filesystem operations MCP, 'git' for Git operations MCP)",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		interruptChan := make(chan os.Signal, 1)
		signal.Notify(interruptChan, os.Interrupt)

		errChan := make(chan error)

		var httpServer *http.Server = nil

		switch transportFlag {
		case "http":
			_, err := net.Dial("tcp", fmt.Sprintf(":%d", portFlag))
			if err == nil {
				return fmt.Errorf("cannot start MCP server, port %d is already in use", portFlag)
			}

			daytonaMcpHandler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
				return daytona.NewDaytonaMCPServer()
			}, &mcp.StreamableHTTPOptions{JSONResponse: true})

			sandboxMcpHandler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
				return sandbox.NewDaytonaSandboxMCPServer()
			}, &mcp.StreamableHTTPOptions{JSONResponse: true})

			fsMcpHandler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
				return fs.NewDaytonaFileSystemMCPServer()
			}, &mcp.StreamableHTTPOptions{JSONResponse: true})

			gitMcpHandler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
				return git.NewDaytonaGitMCPServer()
			}, &mcp.StreamableHTTPOptions{JSONResponse: true})

			httpMux := http.NewServeMux()
			httpMux.Handle("/mcp", daytonaMcpHandler)
			httpMux.Handle("/mcp/sandbox", sandboxMcpHandler)
			httpMux.Handle("/mcp/fs", fsMcpHandler)
			httpMux.Handle("/mcp/git", gitMcpHandler)

			httpServer = &http.Server{
				Addr:    fmt.Sprintf(":%d", portFlag),
				Handler: httpMux,
			}

			listener, err := net.Listen("tcp", httpServer.Addr)
			if err != nil {
				return err
			}

			go func() {
				if tlsCertFilePathFlag != "" && tlsKeyFilePathFlag != "" {
					errChan <- httpServer.ServeTLS(listener, tlsCertFilePathFlag, tlsKeyFilePathFlag)
				} else {
					errChan <- httpServer.Serve(listener)
				}
			}()

		case "stdio":
			mcpServerName := ""
			if len(args) == 1 {
				mcpServerName = args[0]
			}

			var server *mcp.Server
			switch mcpServerName {
			case "":
				server = daytona.NewDaytonaMCPServer()
			case "sandbox":
				server = sandbox.NewDaytonaSandboxMCPServer()
			case "fs":
				server = fs.NewDaytonaFileSystemMCPServer()
			case "git":
				server = git.NewDaytonaGitMCPServer()
			default:
				return fmt.Errorf("mcp server name %s is not supported", mcpServerName)
			}

			go func() {
				errChan <- server.Run(context.Background(), &mcp.StdioTransport{})
			}()

		default:
			return fmt.Errorf("invalid transport provided: %s (use 'stdio' or 'http')", transportFlag)
		}

		select {
		case err := <-errChan:
			return err
		case <-interruptChan:
			if httpServer != nil {
				err := httpServer.Shutdown(context.Background())
				if err != nil {
					log.Errorf("Error while shutting down HTTP server: %v", err)
				}
			}

			return nil
		}
	},
}

func init() {
	ServeCmd.Flags().StringVarP(&transportFlag, "transport", "t", "stdio", "Transport to use for the MCP server (stdio, http). Defaults to stdio.")
	ServeCmd.Flags().IntVarP(&portFlag, "port", "p", 8080, "Port to use for the MCP server. Defaults to 8080.")
	ServeCmd.Flags().StringVarP(&tlsCertFilePathFlag, "tls-cert-file", "c", "", "TLS certificate file to use for the MCP server. Defaults to empty.")
	ServeCmd.Flags().StringVarP(&tlsKeyFilePathFlag, "tls-key-file", "k", "", "TLS key file to use for the MCP server. Defaults to empty.")
}
