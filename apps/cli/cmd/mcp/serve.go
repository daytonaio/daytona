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
	Use:   "serve [MCP_SERVER_NAME]",
	Short: "Serve any Daytona MCP Server (empty string for daytona code execution MCP, 'sandbox' for Sandbox actions MCP, 'fs' for Filesystem operations MCP, 'git' for Git operations MCP)",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var server *mcp.Server
		switch args[0] {
		case "":
			server = daytona.NewDaytonaMCPServer()
		case "sandbox":
			server = sandbox.NewDaytonaSandboxMCPServer()
		case "fs":
			server = fs.NewDaytonaFileSystemMCPServer()
		case "git":
			server = git.NewDaytonaGitMCPServer()
		default:
			return fmt.Errorf("mcp server name %s is not supported", args[0])
		}

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

			handler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
				return server
			}, &mcp.StreamableHTTPOptions{JSONResponse: true})

			httpServer = &http.Server{
				Addr:    ":8080",
				Handler: handler,
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
			_, err := net.Dial("tcp", fmt.Sprintf(":%d", portFlag))
			if err == nil {
				return fmt.Errorf("cannot start MCP server, port %d is already in use", portFlag)
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
