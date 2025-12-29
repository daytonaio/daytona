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
	"time"

	log "github.com/sirupsen/logrus"
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

	log.Infof("SSH tunnel enabled for remote libvirt URI: %s", libvirtURI)
	return true
}

// SSHKeyPath is the path to the SSH private key for dev environment tunneling
var SSHKeyPath = "/workspaces/daytona/.tmp/ssh/id_rsa"

// GetSSHTunnelTransport returns an http.Transport that tunnels through SSH
// to reach VMs on the remote libvirt host. Dev environment only.
func GetSSHTunnelTransport(sshHost string) *http.Transport {
	return &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			host, port, _ := net.SplitHostPort(addr)
			target := fmt.Sprintf("%s:%s", host, port)
			log.Infof("SSH tunnel: dialing %s via %s", target, sshHost)
			cmd := exec.CommandContext(ctx, "ssh",
				"-i", SSHKeyPath,
				"-o", "StrictHostKeyChecking=no",
				"-o", "BatchMode=yes",
				sshHost, "-W", target)
			conn, err := newCmdConn(cmd)
			if err != nil {
				log.Errorf("SSH tunnel: failed to dial %s: %v", target, err)
				return nil, err
			}
			log.Infof("SSH tunnel: connected to %s", target)
			return conn, nil
		},
		ResponseHeaderTimeout: 30 * time.Second,
	}
}

// sshConn wraps SSH stdin/stdout as a net.Conn
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
	return c.cmd.Process.Kill()
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
