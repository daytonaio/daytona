// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package libvirt

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/proxy"
)

// IsRemoteURI returns true if the libvirt URI is a remote SSH connection
func IsRemoteURI(uri string) bool {
	return strings.Contains(uri, "qemu+ssh://")
}

// ShouldUseSSHTunnel checks if we should use SSH tunnel for proxy requests.
// Enabled automatically when LIBVIRT_URI is remote (qemu+ssh://), unless
// LIBVIRT_SSH_TUNNEL=false is set to explicitly disable it.
func ShouldUseSSHTunnel(libvirtURI string) bool {
	if !IsRemoteURI(libvirtURI) {
		return false
	}

	// Check for explicit disable
	if os.Getenv("LIBVIRT_SSH_TUNNEL") == "false" {
		return false
	}

	return true
}

// SSHKeyPath is the path to the SSH private key for dev environment tunneling
var SSHKeyPath = "/workspaces/daytona/.tmp/ssh/id_rsa"

// Persistent SSH SOCKS proxy configuration
var (
	// socksProxyPort is the local port for the SOCKS5 proxy
	socksProxyPort = 10800

	// sshProxyCmd holds the running SSH process for the SOCKS proxy
	sshProxyCmd   *exec.Cmd
	sshProxyMu    sync.Mutex
	sshProxyReady chan struct{}
	sshProxyOnce  sync.Once

	// Cached transports for reuse (keyed by SSH host)
	transportCache   = make(map[string]*http.Transport)
	transportCacheMu sync.RWMutex
)

func init() {
	// Initialize the ready channel
	sshProxyReady = make(chan struct{})
}

// ensureSOCKSProxy starts a persistent SSH SOCKS proxy if not already running.
// This creates a single SSH connection that handles all traffic, eliminating
// the overhead of spawning SSH processes for each request.
func ensureSOCKSProxy(sshHost string) error {
	var startErr error
	sshProxyOnce.Do(func() {
		sshProxyMu.Lock()
		defer sshProxyMu.Unlock()

		// Check if proxy port is already in use (maybe from previous run)
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", socksProxyPort), 100*time.Millisecond)
		if err == nil {
			conn.Close()
			log.Infof("SOCKS proxy already running on port %d", socksProxyPort)
			close(sshProxyReady)
			return
		}

		log.Infof("Starting persistent SSH SOCKS proxy to %s on port %d", sshHost, socksProxyPort)

		// Start SSH with SOCKS proxy (-D)
		sshProxyCmd = exec.Command("ssh",
			"-i", SSHKeyPath,
			"-o", "StrictHostKeyChecking=no",
			"-o", "BatchMode=yes",
			"-o", "ServerAliveInterval=30",
			"-o", "ServerAliveCountMax=3",
			"-o", "ExitOnForwardFailure=yes",
			"-N", // No remote command
			"-D", fmt.Sprintf("127.0.0.1:%d", socksProxyPort),
			sshHost)

		// Capture stderr for debugging
		stderr, err := sshProxyCmd.StderrPipe()
		if err != nil {
			startErr = fmt.Errorf("failed to get stderr pipe: %w", err)
			return
		}

		if err := sshProxyCmd.Start(); err != nil {
			startErr = fmt.Errorf("failed to start SSH SOCKS proxy: %w", err)
			return
		}

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
			err := sshProxyCmd.Wait()
			if err != nil {
				log.Warnf("SSH SOCKS proxy exited: %v", err)
			}
			sshProxyMu.Lock()
			sshProxyCmd = nil
			// Reset the once AND create a new ready channel so it can be restarted
			sshProxyOnce = sync.Once{}
			sshProxyReady = make(chan struct{})
			// Clear transport cache to force reconnection
			transportCacheMu.Lock()
			transportCache = make(map[string]*http.Transport)
			transportCacheMu.Unlock()
			sshProxyMu.Unlock()
			log.Info("SSH SOCKS proxy state reset, will restart on next request")
		}()

		// Wait for proxy to be ready
		maxRetries := 50 // 5 seconds total
		for i := 0; i < maxRetries; i++ {
			conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", socksProxyPort), 100*time.Millisecond)
			if err == nil {
				conn.Close()
				log.Infof("SSH SOCKS proxy ready on port %d", socksProxyPort)
				close(sshProxyReady)
				return
			}
			time.Sleep(100 * time.Millisecond)
		}

		startErr = fmt.Errorf("SSH SOCKS proxy failed to start within 5 seconds")
		if sshProxyCmd != nil && sshProxyCmd.Process != nil {
			sshProxyCmd.Process.Kill()
		}
	})

	if startErr != nil {
		return startErr
	}

	// Wait for proxy to be ready (non-blocking if already ready)
	select {
	case <-sshProxyReady:
		return nil
	case <-time.After(10 * time.Second):
		return fmt.Errorf("timeout waiting for SOCKS proxy to be ready")
	}
}

// GetSSHTunnelTransport returns an http.Transport that routes traffic through
// a persistent SSH SOCKS proxy. This is much faster than spawning SSH processes
// per request because the SSH connection is established once and reused.
func GetSSHTunnelTransport(sshHost string) *http.Transport {
	// Check cache first
	transportCacheMu.RLock()
	if transport, ok := transportCache[sshHost]; ok {
		transportCacheMu.RUnlock()
		return transport
	}
	transportCacheMu.RUnlock()

	// Ensure SOCKS proxy is running
	if err := ensureSOCKSProxy(sshHost); err != nil {
		log.Errorf("Failed to start SOCKS proxy, falling back to direct ssh -W: %v", err)
		return getFallbackTransport(sshHost)
	}

	// Create SOCKS5 dialer
	socksAddr := fmt.Sprintf("127.0.0.1:%d", socksProxyPort)
	dialer, err := proxy.SOCKS5("tcp", socksAddr, nil, proxy.Direct)
	if err != nil {
		log.Errorf("Failed to create SOCKS5 dialer: %v", err)
		return getFallbackTransport(sshHost)
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

// getFallbackTransport returns a transport that uses ssh -W per request.
// This is slower but serves as a fallback if SOCKS proxy fails.
func getFallbackTransport(sshHost string) *http.Transport {
	log.Warnf("Using fallback ssh -W transport for %s (slower)", sshHost)
	return &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			host, port, _ := net.SplitHostPort(addr)
			target := fmt.Sprintf("%s:%s", host, port)
			log.Debugf("SSH tunnel (fallback): dialing %s via %s", target, sshHost)

			cmd := exec.CommandContext(ctx, "ssh",
				"-i", SSHKeyPath,
				"-o", "StrictHostKeyChecking=no",
				"-o", "BatchMode=yes",
				sshHost, "-W", target)

			conn, err := newCmdConn(cmd)
			if err != nil {
				log.Errorf("SSH tunnel (fallback): failed to dial %s: %v", target, err)
				return nil, err
			}
			return conn, nil
		},
		ResponseHeaderTimeout: 30 * time.Second,
	}
}

// sshConn wraps SSH stdin/stdout as a net.Conn (used for fallback)
type sshConn struct {
	stdin  io.WriteCloser
	stdout io.ReadCloser
	cmd    *exec.Cmd
}

func (c *sshConn) Read(b []byte) (int, error) {
	return c.stdout.Read(b)
}

func (c *sshConn) Write(b []byte) (int, error) {
	return c.stdin.Write(b)
}

func (c *sshConn) Close() error {
	c.stdin.Close()
	c.stdout.Close()
	if c.cmd.Process != nil {
		return c.cmd.Process.Kill()
	}
	return nil
}

func (c *sshConn) LocalAddr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0}
}

func (c *sshConn) RemoteAddr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0}
}

func (c *sshConn) SetDeadline(t time.Time) error {
	return nil
}

func (c *sshConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *sshConn) SetWriteDeadline(t time.Time) error {
	return nil
}

// DialSSHTunnel creates a net.Conn that tunnels through SSH to reach a target address.
// Used by SSH gateway to connect to VMs in dev environments.
// Uses the persistent SOCKS proxy when available.
func DialSSHTunnel(sshHost, targetAddr string) (net.Conn, error) {
	log.Debugf("SSH tunnel: dialing %s via %s", targetAddr, sshHost)

	// Try to use SOCKS proxy first
	if err := ensureSOCKSProxy(sshHost); err == nil {
		socksAddr := fmt.Sprintf("127.0.0.1:%d", socksProxyPort)
		dialer, err := proxy.SOCKS5("tcp", socksAddr, nil, proxy.Direct)
		if err == nil {
			conn, err := dialer.Dial("tcp", targetAddr)
			if err == nil {
				log.Debugf("SOCKS proxy: connected to %s", targetAddr)
				return conn, nil
			}
			log.Warnf("SOCKS proxy dial failed, falling back: %v", err)
		}
	}

	// Fallback to ssh -W
	cmd := exec.Command("ssh",
		"-i", SSHKeyPath,
		"-o", "StrictHostKeyChecking=no",
		"-o", "BatchMode=yes",
		sshHost, "-W", targetAddr)

	conn, err := newCmdConn(cmd)
	if err != nil {
		log.Errorf("SSH tunnel (fallback): failed to dial %s: %v", targetAddr, err)
		return nil, err
	}
	log.Debugf("SSH tunnel (fallback): connected to %s", targetAddr)
	return conn, nil
}

// newCmdConn wraps an exec.Cmd's stdin/stdout as a net.Conn using ssh -W
func newCmdConn(cmd *exec.Cmd) (net.Conn, error) {
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdin pipe: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	// Capture stderr to log SSH errors
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stderr pipe: %w", err)
	}
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := stderr.Read(buf)
			if n > 0 {
				log.Warnf("SSH stderr: %s", string(buf[:n]))
			}
			if err != nil {
				return
			}
		}
	}()

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start ssh command: %w", err)
	}

	return &sshConn{
		stdin:  stdin,
		stdout: stdout,
		cmd:    cmd,
	}, nil
}

// StopSSHProxy stops the persistent SSH SOCKS proxy if running.
// Call this during graceful shutdown.
func StopSSHProxy() {
	sshProxyMu.Lock()
	defer sshProxyMu.Unlock()

	if sshProxyCmd != nil && sshProxyCmd.Process != nil {
		log.Info("Stopping SSH SOCKS proxy")
		sshProxyCmd.Process.Kill()
		sshProxyCmd = nil
	}
}
