// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package main

import (
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/daytonaio/mcp/internal/config"
	"github.com/daytonaio/mcp/internal/server"
	"github.com/daytonaio/mcp/internal/util"
	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
)

func main() {
	// Init slog logger
	logger := slog.New(tint.NewHandler(os.Stdout, &tint.Options{
		NoColor:    !isatty.IsTerminal(os.Stdout.Fd()),
		TimeFormat: time.RFC3339,
		Level:      util.ParseLogLevel(os.Getenv("LOG_LEVEL")),
	}))

	slog.SetDefault(logger)

	cfg, err := config.GetConfig()
	if err != nil {
		slog.Error("Failed to get config", "error", err)
		return
	}

	server := server.NewMCPServer(server.MCPServerConfig{
		Port:            cfg.Port,
		TLSCertFilePath: cfg.TLSCertFilePath,
		TLSKeyFilePath:  cfg.TLSKeyFilePath,
		ApiUrl:          cfg.ApiUrl,
		Auth0Domain:     cfg.Auth0Domain,
		Auth0ClientId:   cfg.Auth0ClientId,
		Auth0Audience:   cfg.Auth0Audience,
	})

	mcpServerErrChan := make(chan error)

	go func() {
		mcpServerErrChan <- server.Start()
	}()

	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt)

	select {
	case err := <-mcpServerErrChan:
		slog.Error("MCP server error", "error", err)
		return
	case <-interruptChan:
		server.Stop()
	}
}
