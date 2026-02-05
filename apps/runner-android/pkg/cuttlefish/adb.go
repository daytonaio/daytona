// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package cuttlefish

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

// ADBClient wraps ADB commands for interacting with Android devices
type ADBClient struct {
	adbPath    string
	sshHost    string
	sshKeyPath string
}

// ADBResult contains the result of an ADB command execution
type ADBResult struct {
	Output   string
	ExitCode int
	Error    error
}

// NewADBClient creates a new ADB client
func NewADBClient(adbPath, sshHost, sshKeyPath string) *ADBClient {
	if adbPath == "" {
		adbPath = "adb"
	}
	return &ADBClient{
		adbPath:    adbPath,
		sshHost:    sshHost,
		sshKeyPath: sshKeyPath,
	}
}

// isRemote returns true if operating in remote SSH mode
func (a *ADBClient) isRemote() bool {
	return a.sshHost != ""
}

// Shell executes a shell command on the Android device via ADB
func (a *ADBClient) Shell(ctx context.Context, serial, command string) (string, int, error) {
	args := []string{"-s", serial, "shell", command}
	output, err := a.runADB(ctx, args...)
	if err != nil {
		// Try to extract exit code from error
		if exitErr, ok := err.(*exec.ExitError); ok {
			return output, exitErr.ExitCode(), nil
		}
		return output, 1, err
	}
	return output, 0, nil
}

// ShellWithExitCode executes a shell command and captures the exit code properly
func (a *ADBClient) ShellWithExitCode(ctx context.Context, serial, command string) (string, int, error) {
	// Wrap command to capture exit code
	wrappedCmd := fmt.Sprintf("%s; echo \"\\nEXITCODE:$?\"", command)
	args := []string{"-s", serial, "shell", wrappedCmd}
	output, err := a.runADB(ctx, args...)
	if err != nil {
		return output, 1, err
	}

	// Parse exit code from output
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) > 0 {
		lastLine := lines[len(lines)-1]
		if strings.HasPrefix(lastLine, "EXITCODE:") {
			codeStr := strings.TrimPrefix(lastLine, "EXITCODE:")
			exitCode, _ := strconv.Atoi(strings.TrimSpace(codeStr))
			// Remove the exit code line from output
			output = strings.Join(lines[:len(lines)-1], "\n")
			return output, exitCode, nil
		}
	}

	return output, 0, nil
}

// Push uploads a file to the Android device
func (a *ADBClient) Push(ctx context.Context, serial, localPath, remotePath string) error {
	args := []string{"-s", serial, "push", localPath, remotePath}
	_, err := a.runADB(ctx, args...)
	return err
}

// PushFromReader uploads content from a reader to the Android device
// This is useful when the content is not on the local filesystem (e.g., in remote mode)
func (a *ADBClient) PushFromReader(ctx context.Context, serial string, reader io.Reader, remotePath string) error {
	// Read all content into memory
	content, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read content: %w", err)
	}

	if a.isRemote() {
		// In remote mode, we need to transfer the content via SSH
		// Use base64 encoding to safely transfer binary data
		encoded := base64.StdEncoding.EncodeToString(content)

		// Create a temp file on the remote host, push to device, then cleanup
		tmpPath := fmt.Sprintf("/tmp/adb_push_%d", os.Getpid())
		cmd := fmt.Sprintf("echo '%s' | base64 -d > %s && %s -s %s push %s %s && rm -f %s",
			encoded, tmpPath, a.adbPath, serial, tmpPath, remotePath, tmpPath)

		_, err := a.runSSHCommand(ctx, cmd)
		return err
	}

	// In local mode, create a temp file, push, then cleanup
	tmpFile, err := os.CreateTemp("", "adb_push_*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(content); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	tmpFile.Close()

	return a.Push(ctx, serial, tmpFile.Name(), remotePath)
}

// Pull downloads a file from the Android device
func (a *ADBClient) Pull(ctx context.Context, serial, remotePath, localPath string) error {
	args := []string{"-s", serial, "pull", remotePath, localPath}
	_, err := a.runADB(ctx, args...)
	return err
}

// PullToWriter downloads a file from the Android device to a writer
func (a *ADBClient) PullToWriter(ctx context.Context, serial, remotePath string, writer io.Writer) error {
	if a.isRemote() {
		// In remote mode, pull to temp file, read via SSH, then cleanup
		tmpPath := fmt.Sprintf("/tmp/adb_pull_%d", os.Getpid())
		cmd := fmt.Sprintf("%s -s %s pull %s %s && base64 %s && rm -f %s",
			a.adbPath, serial, remotePath, tmpPath, tmpPath, tmpPath)

		output, err := a.runSSHCommand(ctx, cmd)
		if err != nil {
			return err
		}

		// Decode base64 content
		decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(output))
		if err != nil {
			return fmt.Errorf("failed to decode content: %w", err)
		}

		_, err = writer.Write(decoded)
		return err
	}

	// In local mode, pull to temp file, read, then cleanup
	tmpFile, err := os.CreateTemp("", "adb_pull_*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	if err := a.Pull(ctx, serial, remotePath, tmpFile.Name()); err != nil {
		return err
	}

	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		return fmt.Errorf("failed to read pulled file: %w", err)
	}

	_, err = writer.Write(content)
	return err
}

// Screencap captures a screenshot from the Android device
func (a *ADBClient) Screencap(ctx context.Context, serial string) ([]byte, error) {
	// Use adb exec-out for binary output
	args := []string{"-s", serial, "exec-out", "screencap", "-p"}

	if a.isRemote() {
		// In remote mode, capture and base64 encode
		cmd := fmt.Sprintf("%s -s %s exec-out screencap -p | base64", a.adbPath, serial)
		output, err := a.runSSHCommand(ctx, cmd)
		if err != nil {
			return nil, err
		}
		return base64.StdEncoding.DecodeString(strings.TrimSpace(output))
	}

	// Local mode - capture binary output directly
	cmd := exec.CommandContext(ctx, a.adbPath, args...)
	return cmd.Output()
}

// Input sends input events to the Android device
// inputType can be: tap, swipe, text, keyevent
func (a *ADBClient) Input(ctx context.Context, serial, inputType string, args ...string) error {
	cmdArgs := []string{"-s", serial, "shell", "input", inputType}
	cmdArgs = append(cmdArgs, args...)
	_, err := a.runADB(ctx, cmdArgs...)
	return err
}

// Tap sends a tap event at the specified coordinates
func (a *ADBClient) Tap(ctx context.Context, serial string, x, y int) error {
	return a.Input(ctx, serial, "tap", strconv.Itoa(x), strconv.Itoa(y))
}

// Swipe sends a swipe event from (x1,y1) to (x2,y2) over duration milliseconds
func (a *ADBClient) Swipe(ctx context.Context, serial string, x1, y1, x2, y2, durationMs int) error {
	return a.Input(ctx, serial, "swipe",
		strconv.Itoa(x1), strconv.Itoa(y1),
		strconv.Itoa(x2), strconv.Itoa(y2),
		strconv.Itoa(durationMs))
}

// TypeText types text on the Android device
func (a *ADBClient) TypeText(ctx context.Context, serial, text string) error {
	// Escape special characters for shell
	escaped := strings.ReplaceAll(text, " ", "%s")
	escaped = strings.ReplaceAll(escaped, "'", "\\'")
	return a.Input(ctx, serial, "text", escaped)
}

// KeyEvent sends a key event (e.g., KEYCODE_HOME, KEYCODE_BACK)
func (a *ADBClient) KeyEvent(ctx context.Context, serial string, keycode string) error {
	return a.Input(ctx, serial, "keyevent", keycode)
}

// Install installs an APK on the Android device
func (a *ADBClient) Install(ctx context.Context, serial, apkPath string, flags ...string) error {
	args := []string{"-s", serial, "install"}
	args = append(args, flags...)
	args = append(args, apkPath)
	_, err := a.runADB(ctx, args...)
	return err
}

// InstallFromReader installs an APK from a reader
func (a *ADBClient) InstallFromReader(ctx context.Context, serial string, reader io.Reader, flags ...string) error {
	// Read APK content
	content, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read APK content: %w", err)
	}

	if a.isRemote() {
		// In remote mode, transfer APK via SSH then install
		encoded := base64.StdEncoding.EncodeToString(content)
		tmpPath := fmt.Sprintf("/tmp/install_%d.apk", os.Getpid())

		flagStr := strings.Join(flags, " ")
		cmd := fmt.Sprintf("echo '%s' | base64 -d > %s && %s -s %s install %s %s && rm -f %s",
			encoded, tmpPath, a.adbPath, serial, flagStr, tmpPath, tmpPath)

		_, err := a.runSSHCommand(ctx, cmd)
		return err
	}

	// Local mode - write to temp file and install
	tmpFile, err := os.CreateTemp("", "install_*.apk")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(content); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write APK: %w", err)
	}
	tmpFile.Close()

	return a.Install(ctx, serial, tmpFile.Name(), flags...)
}

// Uninstall removes an app from the Android device
func (a *ADBClient) Uninstall(ctx context.Context, serial, packageName string) error {
	args := []string{"-s", serial, "uninstall", packageName}
	_, err := a.runADB(ctx, args...)
	return err
}

// ListPackages lists installed packages on the Android device
func (a *ADBClient) ListPackages(ctx context.Context, serial string, flags ...string) ([]string, error) {
	cmdArgs := []string{"-s", serial, "shell", "pm", "list", "packages"}
	cmdArgs = append(cmdArgs, flags...)
	output, err := a.runADB(ctx, cmdArgs...)
	if err != nil {
		return nil, err
	}

	var packages []string
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "package:") {
			packages = append(packages, strings.TrimPrefix(line, "package:"))
		}
	}
	return packages, nil
}

// GetProp gets a system property from the Android device
func (a *ADBClient) GetProp(ctx context.Context, serial, prop string) (string, error) {
	output, _, err := a.Shell(ctx, serial, fmt.Sprintf("getprop %s", prop))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

// SetProp sets a system property on the Android device
func (a *ADBClient) SetProp(ctx context.Context, serial, prop, value string) error {
	_, _, err := a.Shell(ctx, serial, fmt.Sprintf("setprop %s %s", prop, value))
	return err
}

// ListFiles lists files in a directory on the Android device
func (a *ADBClient) ListFiles(ctx context.Context, serial, path string) ([]FileInfo, error) {
	// Use ls -la for detailed output
	output, exitCode, err := a.Shell(ctx, serial, fmt.Sprintf("ls -la %s 2>/dev/null || ls -l %s", path, path))
	if err != nil {
		return nil, err
	}
	if exitCode != 0 {
		return nil, fmt.Errorf("ls failed with exit code %d: %s", exitCode, output)
	}

	var files []FileInfo
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "total") {
			continue
		}

		info := parseFileInfo(line, path)
		if info != nil {
			files = append(files, *info)
		}
	}
	return files, nil
}

// FileInfo represents information about a file on the Android device
type FileInfo struct {
	Name       string `json:"name"`
	Path       string `json:"path"`
	Size       int64  `json:"size"`
	Mode       string `json:"mode"`
	IsDir      bool   `json:"isDir"`
	IsSymlink  bool   `json:"isSymlink"`
	Owner      string `json:"owner"`
	Group      string `json:"group"`
	ModTime    string `json:"modTime"`
	LinkTarget string `json:"linkTarget,omitempty"`
}

// parseFileInfo parses a line from ls -la output
func parseFileInfo(line, basePath string) *FileInfo {
	fields := strings.Fields(line)
	if len(fields) < 8 {
		return nil
	}

	mode := fields[0]
	owner := fields[2]
	group := fields[3]

	// Size might be in different positions depending on ls output format
	var size int64
	var name string
	var modTime string

	// Try to parse size - it's usually field 4
	sizeIdx := 4
	if len(fields) > 4 {
		size, _ = strconv.ParseInt(fields[sizeIdx], 10, 64)
	}

	// Name is the last field (or before -> for symlinks)
	nameIdx := len(fields) - 1
	name = fields[nameIdx]

	// Check for symlink
	var linkTarget string
	if strings.Contains(mode, "l") {
		for i, f := range fields {
			if f == "->" && i+1 < len(fields) {
				name = fields[i-1]
				linkTarget = fields[i+1]
				break
			}
		}
	}

	// Mod time is usually 3 fields before the name
	if nameIdx >= 3 {
		modTime = strings.Join(fields[nameIdx-3:nameIdx], " ")
	}

	return &FileInfo{
		Name:       name,
		Path:       filepath.Join(basePath, name),
		Size:       size,
		Mode:       mode,
		IsDir:      strings.HasPrefix(mode, "d"),
		IsSymlink:  strings.HasPrefix(mode, "l"),
		Owner:      owner,
		Group:      group,
		ModTime:    modTime,
		LinkTarget: linkTarget,
	}
}

// Stat gets file information for a specific path
func (a *ADBClient) Stat(ctx context.Context, serial, path string) (*FileInfo, error) {
	// Use stat command for detailed info
	output, exitCode, err := a.Shell(ctx, serial, fmt.Sprintf("stat -c '%%n|%%s|%%F|%%U|%%G|%%y' %s 2>/dev/null", path))
	if err != nil {
		return nil, err
	}

	if exitCode != 0 {
		// Fallback to ls -ld
		output, exitCode, err = a.Shell(ctx, serial, fmt.Sprintf("ls -ld %s", path))
		if err != nil {
			return nil, err
		}
		if exitCode != 0 {
			return nil, fmt.Errorf("file not found: %s", path)
		}
		return parseFileInfo(strings.TrimSpace(output), filepath.Dir(path)), nil
	}

	// Parse stat output
	parts := strings.Split(strings.TrimSpace(output), "|")
	if len(parts) < 6 {
		return nil, fmt.Errorf("unexpected stat output: %s", output)
	}

	size, _ := strconv.ParseInt(parts[1], 10, 64)
	fileType := parts[2]

	return &FileInfo{
		Name:    filepath.Base(parts[0]),
		Path:    parts[0],
		Size:    size,
		IsDir:   strings.Contains(fileType, "directory"),
		Owner:   parts[3],
		Group:   parts[4],
		ModTime: parts[5],
	}, nil
}

// Mkdir creates a directory on the Android device
func (a *ADBClient) Mkdir(ctx context.Context, serial, path string, parents bool) error {
	cmd := "mkdir"
	if parents {
		cmd = "mkdir -p"
	}
	_, exitCode, err := a.Shell(ctx, serial, fmt.Sprintf("%s %s", cmd, path))
	if err != nil {
		return err
	}
	if exitCode != 0 {
		return fmt.Errorf("mkdir failed with exit code %d", exitCode)
	}
	return nil
}

// Remove deletes a file or directory on the Android device
func (a *ADBClient) Remove(ctx context.Context, serial, path string, recursive bool) error {
	cmd := "rm -f"
	if recursive {
		cmd = "rm -rf"
	}
	_, exitCode, err := a.Shell(ctx, serial, fmt.Sprintf("%s %s", cmd, path))
	if err != nil {
		return err
	}
	if exitCode != 0 {
		return fmt.Errorf("rm failed with exit code %d", exitCode)
	}
	return nil
}

// Move moves/renames a file on the Android device
func (a *ADBClient) Move(ctx context.Context, serial, src, dst string) error {
	_, exitCode, err := a.Shell(ctx, serial, fmt.Sprintf("mv %s %s", src, dst))
	if err != nil {
		return err
	}
	if exitCode != 0 {
		return fmt.Errorf("mv failed with exit code %d", exitCode)
	}
	return nil
}

// Copy copies a file on the Android device
func (a *ADBClient) Copy(ctx context.Context, serial, src, dst string) error {
	_, exitCode, err := a.Shell(ctx, serial, fmt.Sprintf("cp %s %s", src, dst))
	if err != nil {
		return err
	}
	if exitCode != 0 {
		return fmt.Errorf("cp failed with exit code %d", exitCode)
	}
	return nil
}

// Logcat streams logcat output
func (a *ADBClient) Logcat(ctx context.Context, serial string, args ...string) (io.ReadCloser, error) {
	cmdArgs := []string{"-s", serial, "logcat"}
	cmdArgs = append(cmdArgs, args...)

	if a.isRemote() {
		// Build SSH command for logcat
		adbCmd := fmt.Sprintf("%s %s", a.adbPath, strings.Join(cmdArgs, " "))
		cmd := exec.CommandContext(ctx, "ssh",
			"-i", a.sshKeyPath,
			"-o", "StrictHostKeyChecking=accept-new",
			"-o", "BatchMode=yes",
			a.sshHost,
			adbCmd,
		)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return nil, err
		}
		if err := cmd.Start(); err != nil {
			return nil, err
		}
		return stdout, nil
	}

	cmd := exec.CommandContext(ctx, a.adbPath, cmdArgs...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return stdout, nil
}

// StartActivity starts an activity on the Android device
func (a *ADBClient) StartActivity(ctx context.Context, serial, component string, extras ...string) error {
	cmd := fmt.Sprintf("am start -n %s", component)
	for _, extra := range extras {
		cmd += " " + extra
	}
	_, exitCode, err := a.Shell(ctx, serial, cmd)
	if err != nil {
		return err
	}
	if exitCode != 0 {
		return fmt.Errorf("am start failed with exit code %d", exitCode)
	}
	return nil
}

// ForceStop force stops an app
func (a *ADBClient) ForceStop(ctx context.Context, serial, packageName string) error {
	_, exitCode, err := a.Shell(ctx, serial, fmt.Sprintf("am force-stop %s", packageName))
	if err != nil {
		return err
	}
	if exitCode != 0 {
		return fmt.Errorf("force-stop failed with exit code %d", exitCode)
	}
	return nil
}

// GetDeviceState returns the device state (device, offline, etc.)
func (a *ADBClient) GetDeviceState(ctx context.Context, serial string) (string, error) {
	args := []string{"-s", serial, "get-state"}
	output, err := a.runADB(ctx, args...)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

// WaitForDevice waits for a device to be available
func (a *ADBClient) WaitForDevice(ctx context.Context, serial string) error {
	args := []string{"-s", serial, "wait-for-device"}
	_, err := a.runADB(ctx, args...)
	return err
}

// runADB executes an ADB command
func (a *ADBClient) runADB(ctx context.Context, args ...string) (string, error) {
	if a.isRemote() {
		// Build command string for SSH
		cmdStr := a.adbPath
		for _, arg := range args {
			if needsQuoting(arg) {
				cmdStr += " '" + strings.ReplaceAll(arg, "'", "'\"'\"'") + "'"
			} else {
				cmdStr += " " + arg
			}
		}
		return a.runSSHCommand(ctx, cmdStr)
	}

	cmd := exec.CommandContext(ctx, a.adbPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	log.Debugf("Running ADB command: %s %v", a.adbPath, args)
	err := cmd.Run()
	output := stdout.String()
	if err != nil {
		output += stderr.String()
	}
	return output, err
}

// runSSHCommand executes a command on the remote host via SSH
func (a *ADBClient) runSSHCommand(ctx context.Context, command string) (string, error) {
	cmd := exec.CommandContext(ctx, "ssh",
		"-i", a.sshKeyPath,
		"-o", "StrictHostKeyChecking=accept-new",
		"-o", "ConnectTimeout=10",
		"-o", "BatchMode=yes",
		a.sshHost,
		command,
	)

	log.Debugf("Running SSH command: ssh -i %s %s '%s'", a.sshKeyPath, a.sshHost, command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Warnf("SSH ADB command failed: cmd=%s, err=%v, output=%s", command, err, string(output))
		return string(output), fmt.Errorf("ssh command failed: %w (output: %s)", err, string(output))
	}
	return string(output), nil
}
