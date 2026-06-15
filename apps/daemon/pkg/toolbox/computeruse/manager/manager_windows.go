//go:build windows

// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package manager

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/daytonaio/daemon/pkg/toolbox/computeruse"
	"github.com/daytonaio/daemon/pkg/winsession"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
)

// consoleSessionPollTimeout bounds how long a plugin spawn waits for AutoLogon
// to produce an interactive console session before failing.
const consoleSessionPollTimeout = 60 * time.Second

// GetComputerUse returns the cached IComputerUse client, or spawns the plugin
// binary into the active console session, publishing the fresh impl via
// publish under the manager lock (see getOrSpawn). Concurrent callers are
// serialized by the manager lock: exactly one spawn is ever in flight and
// every caller receives its result. KillComputerUse takes the same lock, so a
// stop racing an in-flight spawn waits for it to finish and then kills the
// fresh instance — nothing leaks.
func GetComputerUse(logger *slog.Logger, path string, publish func(computeruse.IComputerUse)) (computeruse.IComputerUse, error) {
	return getOrSpawn(func() (pluginClient, computeruse.IComputerUse, error) {
		client, impl, err := spawnInConsoleSession(logger, path)
		if err != nil {
			// Return an untyped nil: wrapping a nil *plugin.Client in the
			// pluginClient interface would make it compare non-nil.
			return nil, nil, err
		}
		return client, impl, nil
	}, publish)
}

// spawnInConsoleSession spawns the plugin binary into the active console
// session and dispenses an IComputerUse client. Callers must hold the manager
// lock (via getOrSpawn).
//
// Single-tenant ephemeral VM assumptions:
//   - No AutoMTLS (only one user on the box; localhost TCP is fine).
//   - No svc.SessionChange handling (one user, one session, ever).
//   - No safe-path STARTUPINFOEX/AttributeList (token-driven session resolution
//     by Windows is sufficient — Tailscale and WireGuard both rely on this).
//
// On error the plugin client is killed in a defer so we don't leak the child.
func spawnInConsoleSession(logger *slog.Logger, path string) (*plugin.Client, computeruse.IComputerUse, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, nil, fmt.Errorf("computer-use plugin not found at path: %s", path)
	}

	token, err := winsession.ActiveConsoleUserToken(consoleSessionPollTimeout)
	if err != nil {
		return nil, nil, err
	}
	// CreateProcessAsUser references the token into the child during Start();
	// our duplicated handle stays ours and must outlive cmd.Start(), which
	// client.Client() below invokes synchronously — so closing at function
	// exit is safe on every path and prevents a per-spawn handle leak.
	defer token.Close()

	cmd := exec.Command(path)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Token:      syscall.Token(token),
		HideWindow: true,
	}

	pluginName := strings.TrimSuffix(filepath.Base(path), ".exe")
	hclogger := hclog.New(&hclog.LoggerOptions{
		Name:   pluginName,
		Output: os.Stdout,
		Level:  hclog.Debug,
	})

	pluginMap := map[string]plugin.Plugin{
		pluginName: &computeruse.ComputerUsePlugin{},
	}

	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: ComputerUseHandshakeConfig,
		Plugins:         pluginMap,
		Cmd:             cmd,
		Logger:          hclogger,
		Managed:         true,
		// AutoMTLS deliberately omitted: single-tenant ephemeral VM.
	})

	success := false
	defer func() {
		if !success {
			client.Kill()
		}
	}()

	logger.Info("Computer use plugin spawn requested", "pluginName", pluginName, "path", path)

	rpcClient, err := client.Client()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get RPC client for computer-use plugin: %w", err)
	}

	raw, err := rpcClient.Dispense(pluginName)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to dispense computer-use plugin: %w", err)
	}

	impl, ok := raw.(computeruse.IComputerUse)
	if !ok {
		return nil, nil, fmt.Errorf("unexpected type from computer-use plugin: %T", raw)
	}

	if _, err := impl.Initialize(); err != nil {
		return nil, nil, fmt.Errorf("failed to initialize computer-use plugin: %w", err)
	}

	success = true
	logger.Info("Computer-use plugin initialized successfully (Windows, session-spawned)")
	return client, impl, nil
}
