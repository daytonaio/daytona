// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package cloudhypervisor

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/proxy"
)

var (
	// Global SOCKS proxy state
	socksProxy     *exec.Cmd
	socksProxyMu   sync.Mutex
	socksProxyPort = 1080
	socksProxyHost = ""
	socksProxyOnce sync.Once
	socksProxyKey  = "" // sshKeyPath for current proxy

	// Transport cache for reusing connections
	transportCache   = make(map[string]*http.Transport)
	transportCacheMu sync.RWMutex
)

// ensureSOCKSProxy starts a SOCKS proxy via SSH if not already running
func ensureSOCKSProxy(sshHost, sshKeyPath string) error {
	var startErr error
	socksProxyOnce.Do(func() {
		socksProxyMu.Lock()
		defer socksProxyMu.Unlock()

		// Check if proxy port is already in use (maybe from previous run)
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", socksProxyPort), 100*time.Millisecond)
		if err == nil {
			conn.Close()
			log.Infof("SOCKS proxy already running on port %d", socksProxyPort)
			return
		}

		// Find an available port
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			startErr = fmt.Errorf("failed to find available port: %w", err)
			return
		}
		socksProxyPort = listener.Addr().(*net.TCPAddr).Port
		listener.Close()

		log.Infof("Starting persistent SSH SOCKS proxy to %s on port %d", sshHost, socksProxyPort)

		// Start SSH with SOCKS proxy (-D flag)
		args := []string{
			"-o", "StrictHostKeyChecking=no",
			"-o", "UserKnownHostsFile=/dev/null",
			"-o", "BatchMode=yes",
			"-o", "ServerAliveInterval=30",
			"-o", "ServerAliveCountMax=3",
			"-o", "ExitOnForwardFailure=yes",
			"-N", // No remote command
			"-D", fmt.Sprintf("127.0.0.1:%d", socksProxyPort),
		}

		if sshKeyPath != "" {
			args = append(args, "-i", sshKeyPath)
		}

		args = append(args, sshHost)

		socksProxy = exec.Command("ssh", args...)

		// Capture stderr for debugging
		stderr, err := socksProxy.StderrPipe()
		if err != nil {
			startErr = fmt.Errorf("failed to get stderr pipe: %w", err)
			return
		}

		if err := socksProxy.Start(); err != nil {
			startErr = fmt.Errorf("failed to start SOCKS proxy: %w", err)
			return
		}

		socksProxyHost = sshHost
		socksProxyKey = sshKeyPath

		// Log stderr in background
		go func() {
			buf := make([]byte, 1024)
			for {
				n, err := stderr.Read(buf)
				if n > 0 {
					log.Debugf("SSH SOCKS proxy stderr: %s", string(buf[:n]))
				}
				if err != nil {
					return
				}
			}
		}()

		// Monitor SSH process and restart if it dies
		go func() {
			err := socksProxy.Wait()
			if err != nil {
				log.Warnf("SSH SOCKS proxy exited: %v", err)
			}
			socksProxyMu.Lock()
			socksProxy = nil
			socksProxyMu.Unlock()
			// Reset the once so it can be restarted
			socksProxyOnce = sync.Once{}
		}()

		log.Infof("Started SOCKS proxy on port %d for %s (PID: %d)", socksProxyPort, sshHost, socksProxy.Process.Pid)

		// Wait for proxy to be ready
		startErr = waitForSOCKSProxy()
	})

	return startErr
}

// waitForSOCKSProxy waits for the SOCKS proxy to be ready
func waitForSOCKSProxy() error {
	addr := fmt.Sprintf("127.0.0.1:%d", socksProxyPort)
	for i := 0; i < 50; i++ {
		conn, err := net.DialTimeout("tcp", addr, 100*time.Millisecond)
		if err == nil {
			conn.Close()
			log.Debugf("SOCKS proxy ready on %s", addr)
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for SOCKS proxy on %s", addr)
}

// GetSSHTunnelTransport returns an http.Transport that routes traffic through
// a persistent SSH SOCKS proxy. This is much faster than spawning SSH processes
// per request because the SSH connection is established once and reused.
func GetSSHTunnelTransport(sshHost, sshKeyPath string) *http.Transport {
	// Check cache first
	transportCacheMu.RLock()
	if transport, ok := transportCache[sshHost]; ok {
		transportCacheMu.RUnlock()
		return transport
	}
	transportCacheMu.RUnlock()

	// Ensure SOCKS proxy is running
	if err := ensureSOCKSProxy(sshHost, sshKeyPath); err != nil {
		log.Errorf("Failed to start SOCKS proxy, falling back to direct ssh -W: %v", err)
		return getFallbackTransport(sshHost, sshKeyPath)
	}

	// Create SOCKS5 dialer
	socksAddr := fmt.Sprintf("127.0.0.1:%d", socksProxyPort)
	dialer, err := proxy.SOCKS5("tcp", socksAddr, nil, proxy.Direct)
	if err != nil {
		log.Errorf("Failed to create SOCKS5 dialer: %v", err)
		return getFallbackTransport(sshHost, sshKeyPath)
	}

	// Create transport using SOCKS proxy
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			log.Debugf("SOCKS proxy: dialing %s", addr)
			conn, err := dialer.Dial(network, addr)
			if err != nil {
				log.Errorf("SOCKS proxy: failed to dial %s: %v", addr, err)
				return nil, err
			}
			log.Debugf("SOCKS proxy: connected to %s", addr)
			return conn, nil
		},
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       90 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
	}

	// Cache the transport
	transportCacheMu.Lock()
	transportCache[sshHost] = transport
	transportCacheMu.Unlock()

	log.Infof("Created SOCKS proxy transport for %s", sshHost)
	return transport
}

// getFallbackTransport returns a transport that uses ssh -W per request
// This is slower but works as a fallback
func getFallbackTransport(sshHost, sshKeyPath string) *http.Transport {
	return &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			args := []string{
				"-o", "StrictHostKeyChecking=no",
				"-o", "UserKnownHostsFile=/dev/null",
			}
			if sshKeyPath != "" {
				args = append(args, "-i", sshKeyPath)
			}
			args = append(args, sshHost, "-W", addr)

			cmd := exec.CommandContext(ctx, "ssh", args...)

			stdin, err := cmd.StdinPipe()
			if err != nil {
				return nil, err
			}

			stdout, err := cmd.StdoutPipe()
			if err != nil {
				stdin.Close()
				return nil, err
			}

			if err := cmd.Start(); err != nil {
				stdin.Close()
				stdout.Close()
				return nil, err
			}

			return &sshTunnelConn{
				stdin:  stdin,
				stdout: stdout,
				cmd:    cmd,
			}, nil
		},
		MaxIdleConns:          10,
		MaxIdleConnsPerHost:   2,
		IdleConnTimeout:       30 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
	}
}

// sshTunnelConn wraps stdin/stdout of an ssh -W process as a net.Conn
type sshTunnelConn struct {
	stdin interface {
		Write([]byte) (int, error)
		Close() error
	}
	stdout interface {
		Read([]byte) (int, error)
		Close() error
	}
	cmd *exec.Cmd
}

func (c *sshTunnelConn) Read(b []byte) (int, error) {
	return c.stdout.Read(b)
}

func (c *sshTunnelConn) Write(b []byte) (int, error) {
	return c.stdin.Write(b)
}

func (c *sshTunnelConn) Close() error {
	c.stdin.Close()
	c.stdout.Close()
	if c.cmd.Process != nil {
		c.cmd.Process.Kill()
	}
	return c.cmd.Wait()
}

func (c *sshTunnelConn) LocalAddr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0}
}

func (c *sshTunnelConn) RemoteAddr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0}
}

func (c *sshTunnelConn) SetDeadline(t time.Time) error {
	return nil
}

func (c *sshTunnelConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *sshTunnelConn) SetWriteDeadline(t time.Time) error {
	return nil
}

// StopSOCKSProxy stops the global SOCKS proxy if running
func StopSOCKSProxy() {
	socksProxyMu.Lock()
	defer socksProxyMu.Unlock()

	if socksProxy != nil && socksProxy.Process != nil {
		log.Info("Stopping SOCKS proxy")
		socksProxy.Process.Kill()
		socksProxy.Wait()
		socksProxy = nil
		socksProxyHost = ""
	}

	// Clear transport cache
	transportCacheMu.Lock()
	transportCache = make(map[string]*http.Transport)
	transportCacheMu.Unlock()
}
