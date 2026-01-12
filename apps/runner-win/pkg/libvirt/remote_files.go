// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package libvirt

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

// Remote file operation helpers for handling both local and remote libvirt hosts.
// When the runner connects to a remote libvirt host via SSH (e.g., qemu+ssh://root@host/system),
// file operations need to be executed on the remote host, not locally.

// fileExists checks if a file exists (handles both local and remote)
func (l *LibVirt) fileExists(ctx context.Context, path string) (bool, error) {
	if l.isLocalURI() {
		_, err := os.Stat(path)
		if os.IsNotExist(err) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		return true, nil
	}

	host := l.extractHostFromURI()
	if host == "" {
		return false, fmt.Errorf("could not extract host from libvirt URI: %s", l.libvirtURI)
	}

	cmd := exec.CommandContext(ctx, "ssh", host, fmt.Sprintf("test -f %s && echo exists", path))
	output, _ := cmd.Output()
	return strings.TrimSpace(string(output)) == "exists", nil
}

// getFileSize returns the size of a file in bytes (handles both local and remote)
func (l *LibVirt) getFileSize(ctx context.Context, path string) (int64, error) {
	if l.isLocalURI() {
		info, err := os.Stat(path)
		if err != nil {
			return 0, err
		}
		return info.Size(), nil
	}

	host := l.extractHostFromURI()
	if host == "" {
		return 0, fmt.Errorf("could not extract host from libvirt URI: %s", l.libvirtURI)
	}

	cmd := exec.CommandContext(ctx, "ssh", host, fmt.Sprintf("stat -c %%s %s", path))
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to get file size: %w", err)
	}

	size, err := strconv.ParseInt(strings.TrimSpace(string(output)), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse file size: %w", err)
	}

	return size, nil
}

// ensureDir creates a directory if it doesn't exist (handles both local and remote)
func (l *LibVirt) ensureDir(ctx context.Context, path string) error {
	if l.isLocalURI() {
		return os.MkdirAll(path, 0755)
	}

	host := l.extractHostFromURI()
	if host == "" {
		return fmt.Errorf("could not extract host from libvirt URI: %s", l.libvirtURI)
	}

	cmd := exec.CommandContext(ctx, "ssh", host, fmt.Sprintf("mkdir -p %s", path))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create directory: %w (output: %s)", err, string(output))
	}

	return nil
}

// removeFile removes a file (handles both local and remote)
func (l *LibVirt) removeFile(ctx context.Context, path string) error {
	if l.isLocalURI() {
		return os.Remove(path)
	}

	host := l.extractHostFromURI()
	if host == "" {
		return fmt.Errorf("could not extract host from libvirt URI: %s", l.libvirtURI)
	}

	cmd := exec.CommandContext(ctx, "ssh", host, fmt.Sprintf("rm -f %s", path))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to remove file: %w (output: %s)", err, string(output))
	}

	return nil
}

// renameFile renames/moves a file (handles both local and remote)
func (l *LibVirt) renameFile(ctx context.Context, oldPath, newPath string) error {
	if l.isLocalURI() {
		return os.Rename(oldPath, newPath)
	}

	host := l.extractHostFromURI()
	if host == "" {
		return fmt.Errorf("could not extract host from libvirt URI: %s", l.libvirtURI)
	}

	cmd := exec.CommandContext(ctx, "ssh", host, fmt.Sprintf("mv %s %s", oldPath, newPath))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to rename file: %w (output: %s)", err, string(output))
	}

	return nil
}

// chmodFile sets file permissions (handles both local and remote)
func (l *LibVirt) chmodFile(ctx context.Context, path string, mode os.FileMode) error {
	if l.isLocalURI() {
		return os.Chmod(path, mode)
	}

	host := l.extractHostFromURI()
	if host == "" {
		return fmt.Errorf("could not extract host from libvirt URI: %s", l.libvirtURI)
	}

	cmd := exec.CommandContext(ctx, "ssh", host, fmt.Sprintf("chmod %o %s", mode, path))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to chmod file: %w (output: %s)", err, string(output))
	}

	return nil
}

// chownLibvirt sets file ownership to libvirt-qemu:kvm (handles both local and remote)
func (l *LibVirt) chownLibvirt(ctx context.Context, path string) error {
	var cmd *exec.Cmd

	if l.isLocalURI() {
		cmd = exec.CommandContext(ctx, "chown", "libvirt-qemu:kvm", path)
	} else {
		host := l.extractHostFromURI()
		if host == "" {
			return fmt.Errorf("could not extract host from libvirt URI: %s", l.libvirtURI)
		}
		cmd = exec.CommandContext(ctx, "ssh", host, fmt.Sprintf("chown libvirt-qemu:kvm %s", path))
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to chown file: %w (output: %s)", err, string(output))
	}

	return nil
}

// runQemuImgCheck validates a qcow2 image (handles both local and remote)
func (l *LibVirt) runQemuImgCheck(ctx context.Context, path string) error {
	var cmd *exec.Cmd

	if l.isLocalURI() {
		cmd = exec.CommandContext(ctx, "qemu-img", "check", path)
	} else {
		host := l.extractHostFromURI()
		if host == "" {
			return fmt.Errorf("could not extract host from libvirt URI: %s", l.libvirtURI)
		}
		cmd = exec.CommandContext(ctx, "ssh", host, fmt.Sprintf("qemu-img check %s", path))
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("qemu-img check failed: %w (output: %s)", err, string(output))
	}

	return nil
}

// sshReadCloser wraps an SSH command that streams file content via stdout
type sshReadCloser struct {
	cmd  *exec.Cmd
	pipe io.ReadCloser
}

func (s *sshReadCloser) Read(p []byte) (int, error) {
	return s.pipe.Read(p)
}

func (s *sshReadCloser) Close() error {
	s.pipe.Close()
	return s.cmd.Wait()
}

// openRemoteFileForRead opens a remote file for reading via SSH cat
func (l *LibVirt) openRemoteFileForRead(ctx context.Context, path string) (io.ReadCloser, error) {
	host := l.extractHostFromURI()
	if host == "" {
		return nil, fmt.Errorf("could not extract host from libvirt URI: %s", l.libvirtURI)
	}

	cmd := exec.CommandContext(ctx, "ssh", host, fmt.Sprintf("cat %s", path))
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start ssh cat: %w", err)
	}

	return &sshReadCloser{cmd: cmd, pipe: stdout}, nil
}

// sshWriteCloser wraps an SSH command that receives file content via stdin
type sshWriteCloser struct {
	cmd  *exec.Cmd
	pipe io.WriteCloser
}

func (s *sshWriteCloser) Write(p []byte) (int, error) {
	return s.pipe.Write(p)
}

func (s *sshWriteCloser) Close() error {
	s.pipe.Close()
	return s.cmd.Wait()
}

// openRemoteFileForWrite opens a remote file for writing via SSH cat >
func (l *LibVirt) openRemoteFileForWrite(ctx context.Context, path string) (io.WriteCloser, error) {
	host := l.extractHostFromURI()
	if host == "" {
		return nil, fmt.Errorf("could not extract host from libvirt URI: %s", l.libvirtURI)
	}

	cmd := exec.CommandContext(ctx, "ssh", host, fmt.Sprintf("cat > %s", path))
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdin pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start ssh cat: %w", err)
	}

	return &sshWriteCloser{cmd: cmd, pipe: stdin}, nil
}

// copyFileToRemote copies a local file to the remote host
func (l *LibVirt) copyFileToRemote(ctx context.Context, localPath, remotePath string) error {
	host := l.extractHostFromURI()
	if host == "" {
		return fmt.Errorf("could not extract host from libvirt URI: %s", l.libvirtURI)
	}

	log.Debugf("Copying %s to %s:%s", localPath, host, remotePath)

	// Use scp for file transfer
	cmd := exec.CommandContext(ctx, "scp", "-o", "StrictHostKeyChecking=no", localPath, fmt.Sprintf("%s:%s", host, remotePath))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("scp failed: %w (output: %s)", err, string(output))
	}

	return nil
}

// copyFileFromRemote copies a file from the remote host to local
func (l *LibVirt) copyFileFromRemote(ctx context.Context, remotePath, localPath string) error {
	host := l.extractHostFromURI()
	if host == "" {
		return fmt.Errorf("could not extract host from libvirt URI: %s", l.libvirtURI)
	}

	log.Debugf("Copying %s:%s to %s", host, remotePath, localPath)

	// Use scp for file transfer
	cmd := exec.CommandContext(ctx, "scp", "-o", "StrictHostKeyChecking=no", fmt.Sprintf("%s:%s", host, remotePath), localPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("scp failed: %w (output: %s)", err, string(output))
	}

	return nil
}

// streamFileToRemote streams data from a reader to a remote file
func (l *LibVirt) streamFileToRemote(ctx context.Context, reader io.Reader, remotePath string) (int64, error) {
	writer, err := l.openRemoteFileForWrite(ctx, remotePath)
	if err != nil {
		return 0, err
	}

	written, err := io.Copy(writer, reader)
	closeErr := writer.Close()

	if err != nil {
		return written, err
	}
	if closeErr != nil {
		return written, closeErr
	}

	return written, nil
}

// streamFileFromRemote streams data from a remote file to a writer
func (l *LibVirt) streamFileFromRemote(ctx context.Context, remotePath string, writer io.Writer) (int64, error) {
	reader, err := l.openRemoteFileForRead(ctx, remotePath)
	if err != nil {
		return 0, err
	}
	defer reader.Close()

	return io.Copy(writer, reader)
}

// readDir lists files in a directory (handles both local and remote)
func (l *LibVirt) readDir(ctx context.Context, path string) ([]string, error) {
	if l.isLocalURI() {
		entries, err := os.ReadDir(path)
		if err != nil {
			return nil, err
		}
		var names []string
		for _, entry := range entries {
			names = append(names, entry.Name())
		}
		return names, nil
	}

	host := l.extractHostFromURI()
	if host == "" {
		return nil, fmt.Errorf("could not extract host from libvirt URI: %s", l.libvirtURI)
	}

	cmd := exec.CommandContext(ctx, "ssh", host, fmt.Sprintf("ls -1 %s 2>/dev/null || true", path))
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list directory: %w", err)
	}

	var names []string
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		if line != "" {
			names = append(names, line)
		}
	}

	return names, nil
}

// progressWriter wraps an io.Writer to log write progress
type progressWriter struct {
	writer        io.Writer
	total         int64
	written       int64
	name          string
	lastLog       int64
	logInterval   int64
	isDownloading bool
}

func (pw *progressWriter) Write(p []byte) (int, error) {
	n, err := pw.writer.Write(p)
	pw.written += int64(n)

	if pw.written-pw.lastLog >= pw.logInterval {
		percent := float64(pw.written) / float64(pw.total) * 100
		action := "Uploading"
		if pw.isDownloading {
			action = "Downloading"
		}
		log.Infof("%s '%s': %.1f%% (%d / %d bytes)", action, pw.name, percent, pw.written, pw.total)
		pw.lastLog = pw.written
	}

	return n, err
}

// countingReader wraps an io.Reader to count bytes read
type countingReader struct {
	reader io.Reader
	count  int64
}

func (cr *countingReader) Read(p []byte) (int, error) {
	n, err := cr.reader.Read(p)
	cr.count += int64(n)
	return n, err
}

// execRemoteCommand executes a command on the remote host and returns output
func (l *LibVirt) execRemoteCommand(ctx context.Context, command string) (string, error) {
	host := l.extractHostFromURI()
	if host == "" {
		return "", fmt.Errorf("could not extract host from libvirt URI: %s", l.libvirtURI)
	}

	cmd := exec.CommandContext(ctx, "ssh", host, command)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("command failed: %w (stderr: %s)", err, stderr.String())
	}

	return stdout.String(), nil
}
