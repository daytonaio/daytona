//go:build windows

// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package manager

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/daytonaio/daemon/pkg/toolbox/computeruse"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"golang.org/x/sys/windows"
)

const (
	consoleSessionPollInterval = 500 * time.Millisecond
	consoleSessionPollTimeout  = 60 * time.Second
)

// ErrNoActiveConsoleSession is returned when no interactive user is logged on
// within consoleSessionPollTimeout. In Daytona's Windows sandbox image this
// should never fire after AutoLogon completes.
var ErrNoActiveConsoleSession = errors.New("no active console session available; ensure a user is logged on (AutoLogon)")

// activeConsoleUserToken polls WTSGetActiveConsoleSessionId until a non-sentinel
// session id appears, then queries and duplicates the user's token to a primary
// token suitable for exec.Cmd.SysProcAttr.Token. Caller owns the returned handle
// and MUST ensure it lives until exec.Cmd.Start() returns; do NOT Close it
// immediately after attaching it to SysProcAttr, since Windows duplicates the
// handle during CreateProcessAsUser.
func activeConsoleUserToken() (windows.Token, error) {
	deadline := time.Now().Add(consoleSessionPollTimeout)
	for {
		sid := windows.WTSGetActiveConsoleSessionId()
		if sid != 0xFFFFFFFF {
			var raw windows.Token
			if err := windows.WTSQueryUserToken(sid, &raw); err == nil {
				var primary windows.Token
				err := windows.DuplicateTokenEx(
					raw,
					windows.MAXIMUM_ALLOWED,
					nil,
					windows.SecurityImpersonation,
					windows.TokenPrimary,
					&primary,
				)
				raw.Close()
				if err == nil {
					return primary, nil
				}
			}
		}
		if time.Now().After(deadline) {
			return 0, ErrNoActiveConsoleSession
		}
		time.Sleep(consoleSessionPollInterval)
	}
}

// GetComputerUse returns the cached IComputerUse client, or spawns the plugin
// binary into the active console session. Concurrent callers are serialized by
// the manager lock (getOrSpawn): exactly one spawn is ever in flight and every
// caller receives its result. KillComputerUse takes the same lock, so a stop
// racing an in-flight spawn waits for it to finish and then kills the fresh
// instance — nothing leaks.
func GetComputerUse(logger *slog.Logger, path string) (computeruse.IComputerUse, error) {
	return getOrSpawn(func() (*plugin.Client, computeruse.IComputerUse, string, error) {
		client, impl, err := spawnInConsoleSession(logger, path)
		if err != nil {
			return nil, nil, "", err
		}
		return client, impl, filepath.Dir(path), nil
	})
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

	token, err := activeConsoleUserToken()
	if err != nil {
		return nil, nil, err
	}
	// token ownership transfers to SysProcAttr; do NOT Close it here.

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
