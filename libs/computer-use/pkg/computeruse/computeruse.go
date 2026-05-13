// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/daytonaio/daemon/pkg/toolbox/computeruse"
	"github.com/godbus/dbus/v5"
	log "github.com/sirupsen/logrus"
)

type Process struct {
	Name         string
	Command      string
	Args         []string
	User         string
	Priority     int
	Required     bool
	Env          map[string]string
	LogFile      string
	ErrFile      string
	AutoRestart  bool
	Ready        func(context.Context, *Process) error
	ReadyName    string
	ReadyTimeout time.Duration
	childPID     int
	cancel       context.CancelFunc
	supervisorID uint64
	mu           sync.Mutex
	supervising  bool
	stopping     bool
}

type ComputerUse struct {
	processes map[string]*Process
	mu        sync.RWMutex
	configDir string

	// AT-SPI accessibility bus connection. Lazily established on first call to
	// connectA11y(); protected by atspiMu. Implementation lives in accessibility.go.
	atspiMu   sync.Mutex
	atspiConn *dbus.Conn

	a11yHealth func() bool
	waitDBus   func(string, time.Duration) error

	a11yStatusMu        sync.Mutex
	a11yStatusRunning   bool
	a11yStatusCheckedAt time.Time
	restartDelay        time.Duration
}

var _ computeruse.IComputerUse = &ComputerUse{}

type processFile interface {
	io.Writer
	Close() error
}

var openProcessFile = func(name string) (processFile, error) {
	return os.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
}

func (c *ComputerUse) Initialize() (*computeruse.Empty, error) {
	c.processes = make(map[string]*Process)
	// Create config directory for logs
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return new(computeruse.Empty), fmt.Errorf("failed to get home directory: %v", err)
	}
	c.configDir = filepath.Join(homeDir, ".daytona", "computeruse")
	err = os.MkdirAll(c.configDir, 0755)
	if err != nil {
		return new(computeruse.Empty), fmt.Errorf("failed to create config directory: %v", err)
	}

	// Start a D-Bus session and set env vars globally
	cmd := exec.Command("dbus-launch")
	output, err := cmd.Output()
	if err != nil {
		log.Errorf("Failed to start dbus-launch: %v", err)
	} else {
		for _, line := range strings.Split(string(output), "\n") {
			if strings.HasPrefix(line, "DBUS_SESSION_BUS_ADDRESS=") || strings.HasPrefix(line, "DBUS_SESSION_BUS_PID=") {
				parts := strings.SplitN(line, ";", 2)
				for _, part := range parts {
					kv := strings.SplitN(part, "=", 2)
					if len(kv) == 2 {
						os.Setenv(strings.TrimSpace(kv[0]), strings.TrimSpace(kv[1]))
					}
				}
			}
		}
	}

	if err := c.initializeProcesses(homeDir); err != nil {
		return new(computeruse.Empty), err
	}

	return new(computeruse.Empty), nil
}

func (c *ComputerUse) Start() (*computeruse.Empty, error) {
	// Set DISPLAY environment variable in the main process
	display := os.Getenv("DISPLAY")
	if display == "" {
		display = ":0"
	}
	os.Setenv("DISPLAY", display)
	log.Infof("Set DISPLAY environment variable to: %s", display)

	// Start all processes in order of priority
	if err := c.startAllProcesses(context.Background()); err != nil {
		return nil, err
	}
	c.invalidateA11yStatus()

	// Check process status after starting
	status, err := c.GetProcessStatus()
	if err != nil {
		return nil, fmt.Errorf("failed to get process status after start: %v", err)
	}

	// Check if all required processes are running. atspi is deliberately
	// excluded — a11y failures surface as 503 A11Y_UNAVAILABLE on the /a11y
	// endpoints, and should not block the rest of computer-use from starting.
	required := []string{"xvfb", "xfce4", "x11vnc", "novnc"}
	var failed []string
	for _, name := range required {
		if s, ok := status[name]; !ok || !s.Running {
			failed = append(failed, name)
		}
	}

	if len(failed) > 0 {
		return nil, fmt.Errorf("failed to start: %v", failed)
	}

	return new(computeruse.Empty), nil
}

func (c *ComputerUse) initializeProcesses(homeDir string) error {
	// Get environment variables from Dockerfile or use defaults
	vncResolution := os.Getenv("VNC_RESOLUTION")
	if vncResolution == "" {
		vncResolution = "1024x768"
	}

	vncPort := os.Getenv("VNC_PORT")
	if vncPort == "" {
		vncPort = "5901"
	}

	noVncPort := os.Getenv("NO_VNC_PORT")
	if noVncPort == "" {
		noVncPort = "6080"
	}

	display := os.Getenv("DISPLAY")
	if display == "" {
		display = ":0"
	}

	// Get user from environment, fallback to DAYTONA_SANDBOX_USER or default to "root" (just in case, but should not happen)
	user := os.Getenv("VNC_USER")
	if user == "" {
		user = os.Getenv("DAYTONA_SANDBOX_USER")
		if user == "" {
			user = "root"
		}
	}

	// Get D-Bus session address from environment
	dbusAddress := os.Getenv("DBUS_SESSION_BUS_ADDRESS")

	novncCommand, novncArgs, err := selectNoVNCCommand(vncPort, noVncPort)
	if err != nil {
		return fmt.Errorf("failed to initialize novnc process: %w", err)
	}

	// Process 1: Xvfb (X Virtual Framebuffer)
	c.processes["xvfb"] = &Process{
		Name:         "xvfb",
		Command:      "/usr/bin/Xvfb",
		Args:         []string{display, "-screen", "0", vncResolution + "x24"},
		User:         user,
		Priority:     100,
		Required:     true,
		AutoRestart:  true,
		Ready:        readyXvfb(display),
		ReadyName:    "xvfb-display",
		ReadyTimeout: 10 * time.Second,
		Env: map[string]string{
			"DISPLAY": display,
		},
		LogFile: filepath.Join(c.configDir, "xvfb.log"),
		ErrFile: filepath.Join(c.configDir, "xvfb.err"),
	}

	// Process 2: xfce4 (Desktop Environment)
	c.processes["xfce4"] = &Process{
		Name:         "xfce4",
		Command:      "/usr/bin/startxfce4",
		Args:         []string{},
		User:         user,
		Priority:     200,
		Required:     true,
		AutoRestart:  true,
		Ready:        readyXfce(display, user),
		ReadyName:    "xfce4-desktop",
		ReadyTimeout: 20 * time.Second,
		Env: map[string]string{
			"DISPLAY":                  display,
			"HOME":                     homeDir,
			"USER":                     user,
			"DBUS_SESSION_BUS_ADDRESS": dbusAddress,
		},
		LogFile: filepath.Join(c.configDir, "xfce4.log"),
		ErrFile: filepath.Join(c.configDir, "xfce4.err"),
	}

	// Process 2.5: at-spi-bus-launcher (AT-SPI daemon for the accessibility API)
	// Launches the org.a11y.Bus service so GTK/Qt/Electron apps can publish
	// their widget trees. It uses the session D-Bus created during Initialize.
	// Path lives in /usr/libexec on Debian/Ubuntu; some distros ship it under
	// /usr/lib/at-spi2-core/; as a last resort fall back to $PATH so images
	// that put the binary somewhere unusual still work. If we can't find the
	// binary anywhere, we skip registering atspi entirely — the a11y endpoints
	// will return 503 cleanly instead of thrashing on a nonexistent command.
	atspiCommand := ""
	if _, err := os.Stat("/usr/libexec/at-spi-bus-launcher"); err == nil {
		atspiCommand = "/usr/libexec/at-spi-bus-launcher"
	} else if _, err := os.Stat("/usr/lib/at-spi2-core/at-spi-bus-launcher"); err == nil {
		atspiCommand = "/usr/lib/at-spi2-core/at-spi-bus-launcher"
	} else if p, err := exec.LookPath("at-spi-bus-launcher"); err == nil {
		atspiCommand = p
	}
	if atspiCommand == "" {
		log.Warnf("at-spi-bus-launcher not found in any known location; accessibility API will return 503 until the binary is installed")
	} else if dbusAddress == "" {
		log.Warnf("DBUS_SESSION_BUS_ADDRESS is empty; accessibility API will return 503 until session D-Bus is available")
	} else {
		waitDBus := waitForSessionBus
		if c.waitDBus != nil {
			waitDBus = c.waitDBus
		}
		if err := waitDBus(dbusAddress, 5*time.Second); err != nil {
			log.Warnf("session D-Bus check failed for at-spi-bus-launcher; launching anyway so the a11y bus can start if D-Bus becomes usable: %v", err)
		}
		c.processes["atspi"] = &Process{
			Name:        "atspi",
			Command:     atspiCommand,
			Args:        []string{"--launch-immediately"},
			User:        user,
			Priority:    250,
			AutoRestart: false,
			Ready: func(context.Context, *Process) error {
				if c.isA11yAvailable() {
					return nil
				}
				return fmt.Errorf("AT-SPI bus is unavailable")
			},
			ReadyName:    "atspi-bus",
			ReadyTimeout: 2 * time.Second,
			Env: map[string]string{
				"DISPLAY":                  display,
				"HOME":                     homeDir,
				"USER":                     user,
				"DBUS_SESSION_BUS_ADDRESS": dbusAddress,
			},
			LogFile: filepath.Join(c.configDir, "atspi.log"),
			ErrFile: filepath.Join(c.configDir, "atspi.err"),
		}
	}

	// Process 3: x11vnc (VNC Server)
	c.processes["x11vnc"] = &Process{
		Name:         "x11vnc",
		Command:      "/usr/bin/x11vnc",
		Args:         []string{"-display", display, "-forever", "-shared", "-rfbport", vncPort},
		User:         user,
		Priority:     300,
		Required:     true,
		AutoRestart:  true,
		Ready:        readyTCP("127.0.0.1", vncPort),
		ReadyName:    "x11vnc-tcp",
		ReadyTimeout: 10 * time.Second,
		Env: map[string]string{
			"DISPLAY": display,
		},
		LogFile: filepath.Join(c.configDir, "x11vnc.log"),
		ErrFile: filepath.Join(c.configDir, "x11vnc.err"),
	}

	c.processes["novnc"] = &Process{
		Name:         "novnc",
		Command:      novncCommand,
		Args:         novncArgs,
		User:         user,
		Priority:     400,
		Required:     true,
		AutoRestart:  true,
		Ready:        readyHTTP("127.0.0.1", noVncPort),
		ReadyName:    "novnc-http",
		ReadyTimeout: 10 * time.Second,
		Env: map[string]string{
			"DISPLAY": display,
		},
		LogFile: filepath.Join(c.configDir, "novnc.log"),
		ErrFile: filepath.Join(c.configDir, "novnc.err"),
	}
	return nil
}

func selectNoVNCCommand(vncPort, noVncPort string) (string, []string, error) {
	command, err := exec.LookPath("websockify")
	if err != nil {
		return "", nil, fmt.Errorf("websockify not found in PATH: %w", err)
	}
	log.Infof("Using direct websockify")
	return command, []string{"--web=/usr/share/novnc/", noVncPort, "localhost:" + vncPort}, nil
}

func waitForSessionBus(address string, timeout time.Duration) error {
	if address == "" {
		return fmt.Errorf("DBUS_SESSION_BUS_ADDRESS is empty")
	}

	deadline := time.Now().Add(timeout)
	var lastErr error
	for {
		conn, err := dbus.Connect(address)
		if err == nil {
			_ = conn.Close()
			return nil
		}
		lastErr = err

		if time.Now().After(deadline) {
			return fmt.Errorf("session D-Bus did not become ready: %w", lastErr)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func (c *ComputerUse) startAllProcesses(ctx context.Context) error {
	// Sort processes by priority and start them
	processes := c.getProcessesByPriority()
	started := make([]*Process, 0, len(processes))

	for _, process := range processes {
		if supervisorID, ok := process.beginSupervising(); ok {
			go c.superviseProcess(process, supervisorID)
			started = append(started, process)
		}
		if !process.Required {
			go c.logOptionalReadiness(ctx, process)
			continue
		}
		if err := c.waitProcessReady(ctx, process); err != nil {
			for i := len(started) - 1; i >= 0; i-- {
				c.stopProcess(started[i])
			}
			return err
		}
	}

	return nil
}

func (c *ComputerUse) getProcessesByPriority() []*Process {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var processes []*Process
	for _, p := range c.processes {
		processes = append(processes, p)
	}

	// Sort by priority (lower number = higher priority)
	for i := 0; i < len(processes)-1; i++ {
		for j := i + 1; j < len(processes); j++ {
			if processes[i].Priority > processes[j].Priority {
				processes[i], processes[j] = processes[j], processes[i]
			}
		}
	}

	return processes
}

func (c *ComputerUse) startProcess(process *Process) {
	supervisorID, ok := process.beginSupervising()
	if !ok {
		return
	}
	c.superviseProcess(process, supervisorID)
}

func (process *Process) beginSupervising() (uint64, bool) {
	process.mu.Lock()
	defer process.mu.Unlock()

	if process.supervising {
		return 0, false
	}

	process.supervisorID++
	process.supervising = true
	process.stopping = false
	process.childPID = 0
	process.cancel = nil
	return process.supervisorID, true
}

func (c *ComputerUse) superviseProcess(process *Process, supervisorID uint64) {
	defer func() {
		process.mu.Lock()
		if process.supervisorID == supervisorID {
			process.supervising = false
			process.stopping = false
			process.childPID = 0
			process.cancel = nil
		}
		process.mu.Unlock()
	}()

	for {
		log.Infof("Starting process: %s", process.Name)
		err := c.runProcessOnce(process, supervisorID)
		if err != nil {
			log.Errorf("Process %s exited with error: %v", process.Name, err)
		} else {
			log.Infof("Process %s exited normally", process.Name)
		}

		// Check if we should restart
		if !process.AutoRestart || !process.isSupervising(supervisorID) {
			break
		}

		delay := c.processRestartDelay()
		log.Infof("Restarting process %s in %s...", process.Name, delay)
		time.Sleep(delay)
		if !process.isSupervising(supervisorID) {
			break
		}
	}
}

func (c *ComputerUse) runProcessOnce(process *Process, supervisorID uint64) error {
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, process.Command, process.Args...)
	if len(process.Env) > 0 {
		cmd.Env = os.Environ()
		for key, value := range process.Env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
		}
	}

	logFile, errFile := c.openProcessLogs(process, cmd)
	defer closeProcessFile(process.Name, "log", logFile)
	defer closeProcessFile(process.Name, "error", errFile)

	process.mu.Lock()
	if process.stopping || !process.supervising || process.supervisorID != supervisorID {
		process.mu.Unlock()
		cancel()
		return nil
	}
	process.cancel = cancel
	process.mu.Unlock()

	if err := cmd.Start(); err != nil {
		c.clearChildProcess(process, supervisorID, 0)
		cancel()
		return fmt.Errorf("failed to start process: %w", err)
	}

	process.mu.Lock()
	if process.stopping || !process.supervising || process.supervisorID != supervisorID {
		process.mu.Unlock()
		cancel()
		killProcess(cmd.Process.Pid)
		err := cmd.Wait()
		c.clearChildProcess(process, supervisorID, cmd.Process.Pid)
		return err
	}
	process.childPID = cmd.Process.Pid
	process.mu.Unlock()

	log.Infof("Process %s started with PID: %d", process.Name, cmd.Process.Pid)
	err := cmd.Wait()
	cancel()
	c.clearChildProcess(process, supervisorID, cmd.Process.Pid)
	return err
}

func (c *ComputerUse) clearChildProcess(process *Process, supervisorID uint64, pid int) {
	process.mu.Lock()
	defer process.mu.Unlock()

	if process.supervisorID != supervisorID {
		return
	}
	if pid != 0 && process.childPID == pid {
		process.childPID = 0
	}
	process.cancel = nil
}

func (c *ComputerUse) openProcessLogs(process *Process, cmd *exec.Cmd) (processFile, processFile) {
	var logFile, errFile processFile
	if process.LogFile != "" {
		file, err := openProcessFile(process.LogFile)
		if err != nil {
			log.Errorf("Failed to open log file for %s: %v", process.Name, err)
		} else {
			cmd.Stdout = file
			logFile = file
		}
	}

	if process.ErrFile != "" {
		file, err := openProcessFile(process.ErrFile)
		if err != nil {
			log.Errorf("Failed to open error file for %s: %v", process.Name, err)
		} else {
			cmd.Stderr = file
			errFile = file
		}
	}

	return logFile, errFile
}

func closeProcessFile(processName, fileType string, file processFile) {
	if file == nil {
		return
	}
	if err := file.Close(); err != nil {
		log.Errorf("Failed to close %s file for %s: %v", fileType, processName, err)
	}
}

func (c *ComputerUse) processRestartDelay() time.Duration {
	if c.restartDelay > 0 {
		return c.restartDelay
	}
	return 5 * time.Second
}

func (process *Process) isSupervising(supervisorID uint64) bool {
	process.mu.Lock()
	defer process.mu.Unlock()
	return process.supervising && process.supervisorID == supervisorID
}

func (c *ComputerUse) logOptionalReadiness(ctx context.Context, process *Process) {
	if err := c.waitProcessReady(ctx, process); err != nil {
		log.Warnf("Optional process %s is not ready: %v", process.Name, err)
	}
}

func (c *ComputerUse) waitProcessReady(ctx context.Context, process *Process) error {
	if process.Ready == nil {
		return nil
	}

	timeout := process.ReadyTimeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	probe := process.ReadyName
	if probe == "" {
		probe = "ready"
	}

	readyCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var lastErr error
	for {
		if err := process.Ready(readyCtx, process); err == nil {
			return nil
		} else {
			lastErr = err
		}

		select {
		case <-readyCtx.Done():
			return fmt.Errorf("process %s readiness probe %s timed out after %s: %v", process.Name, probe, timeout, lastErr)
		case <-time.After(100 * time.Millisecond):
		}
	}
}

func readyXvfb(display string) func(context.Context, *Process) error {
	return func(ctx context.Context, _ *Process) error {
		if path, err := exec.LookPath("xdpyinfo"); err == nil {
			return runDisplayCommand(ctx, display, path, "-display", display)
		}

		socket, err := displaySocket(display)
		if err != nil {
			return err
		}
		if _, err := os.Stat(socket); err != nil {
			return fmt.Errorf("X socket %s is not ready: %w", socket, err)
		}
		return nil
	}
}

func readyXfce(display, user string) func(context.Context, *Process) error {
	return func(ctx context.Context, _ *Process) error {
		if path, err := exec.LookPath("xprop"); err == nil {
			return readyWindowManager(ctx, display, path)
		}

		args := []string{"-f", "xfce4-session|xfdesktop|xfwm4"}
		if user != "" {
			args = []string{"-u", user, "-f", "xfce4-session|xfdesktop|xfwm4"}
		}
		if err := exec.CommandContext(ctx, "pgrep", args...).Run(); err != nil {
			return fmt.Errorf("xfce desktop process is not ready: %w", err)
		}
		return nil
	}
}

func readyTCP(host, port string) func(context.Context, *Process) error {
	return func(ctx context.Context, _ *Process) error {
		dialer := net.Dialer{}
		conn, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(host, port))
		if err != nil {
			return err
		}
		return conn.Close()
	}
}

func readyHTTP(host, port string) func(context.Context, *Process) error {
	return func(ctx context.Context, _ *Process) error {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+net.JoinHostPort(host, port)+"/", nil)
		if err != nil {
			return err
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode >= http.StatusInternalServerError {
			return fmt.Errorf("HTTP status %s", resp.Status)
		}
		return nil
	}
}

func runDisplayCommand(ctx context.Context, display, command string, args ...string) error {
	_, err := runDisplayCommandOutput(ctx, display, command, args...)
	return err
}

func runDisplayCommandOutput(ctx context.Context, display, command string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Env = append(os.Environ(), "DISPLAY="+display)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s failed: %w: %s", filepath.Base(command), err, strings.TrimSpace(string(output)))
	}
	return string(output), nil
}

func readyWindowManager(ctx context.Context, display, xprop string) error {
	rootID, err := xpropWindowID(ctx, display, xprop, "-root", "_NET_SUPPORTING_WM_CHECK")
	if err != nil {
		return err
	}

	childID, err := xpropWindowID(ctx, display, xprop, "-id", rootID, "_NET_SUPPORTING_WM_CHECK")
	if err != nil {
		return err
	}
	if childID != rootID {
		return fmt.Errorf("window manager check window mismatch: root points to %s, child points to %s", rootID, childID)
	}
	return nil
}

func xpropWindowID(ctx context.Context, display, xprop string, args ...string) (string, error) {
	output, err := runDisplayCommandOutput(ctx, display, xprop, args...)
	if err != nil {
		return "", err
	}
	return parseXpropWindowID(output)
}

func parseXpropWindowID(output string) (string, error) {
	if strings.Contains(output, "not found") {
		return "", fmt.Errorf("window manager check property is not ready")
	}

	i := strings.Index(output, "#")
	if i < 0 {
		return "", fmt.Errorf("window manager check property has no window id: %s", strings.TrimSpace(output))
	}
	fields := strings.Fields(output[i+1:])
	if len(fields) == 0 {
		return "", fmt.Errorf("window manager check property has an empty window id")
	}

	id := strings.Trim(fields[0], ",")
	n, err := strconv.ParseUint(id, 0, 64)
	if err != nil || n == 0 {
		return "", fmt.Errorf("window manager check property has invalid window id %q", id)
	}
	return fmt.Sprintf("0x%x", n), nil
}

func displaySocket(display string) (string, error) {
	display = strings.TrimPrefix(display, ":")
	display = strings.Split(display, ".")[0]
	display = strings.Split(display, " ")[0]
	n, err := strconv.Atoi(display)
	if err != nil {
		return "", fmt.Errorf("cannot derive X socket from DISPLAY=%q: %w", display, err)
	}
	return fmt.Sprintf("/tmp/.X11-unix/X%d", n), nil
}

func (c *ComputerUse) Stop() (*computeruse.Empty, error) {
	log.Info("Stopping all computer use processes...")

	c.mu.RLock()
	processes := make([]*Process, 0, len(c.processes))
	for _, p := range c.processes {
		processes = append(processes, p)
	}
	c.mu.RUnlock()

	// Stop processes in reverse priority order
	for i := len(processes) - 1; i >= 0; i-- {
		process := processes[i]
		c.stopProcess(process)
	}

	// Release the cached AT-SPI bus connection so a later Start() dials a
	// fresh one — the bus address can change across launcher restarts.
	c.atspiMu.Lock()
	if c.atspiConn != nil {
		_ = c.atspiConn.Close()
		c.atspiConn = nil
	}
	c.atspiMu.Unlock()
	c.setA11yStatus(false)

	return new(computeruse.Empty), nil
}

func (c *ComputerUse) stopProcess(process *Process) {
	process.mu.Lock()
	if !process.supervising && process.childPID == 0 {
		process.mu.Unlock()
		return
	}

	log.Infof("Stopping process: %s", process.Name)
	process.stopping = true
	process.supervising = false
	cancel := process.cancel
	childPID := process.childPID
	process.mu.Unlock()

	// Cancel the context to stop the process
	if cancel != nil {
		cancel()
	}

	// Kill the process if it's still running
	if childPID != 0 {
		killProcess(childPID)
	}
}

func killProcess(pid int) {
	process, err := os.FindProcess(pid)
	if err != nil {
		log.Errorf("Failed to find process %d: %v", pid, err)
		return
	}
	if err := process.Kill(); err != nil && !errors.Is(err, os.ErrProcessDone) {
		log.Errorf("Failed to kill process %d: %v", pid, err)
	}
}

func (c *ComputerUse) GetProcessStatus() (map[string]computeruse.ProcessStatus, error) {
	c.mu.RLock()
	processes := make(map[string]*Process, len(c.processes))
	for name, process := range c.processes {
		processes[name] = process
	}
	c.mu.RUnlock()

	status := make(map[string]computeruse.ProcessStatus, len(processes))
	for name, process := range processes {
		processStatus := getProcessStatus(process)
		if name == "atspi" {
			processStatus.Running = c.cachedA11yAvailable()
			processStatus.Pid = nil
		}
		status[name] = processStatus
	}

	return status, nil
}

func getProcessStatus(process *Process) computeruse.ProcessStatus {
	process.mu.Lock()
	defer process.mu.Unlock()

	status := computeruse.ProcessStatus{
		Running:     false,
		Priority:    process.Priority,
		AutoRestart: process.AutoRestart,
	}

	childPID := process.childPID
	if childPID == 0 {
		return status
	}

	child, err := os.FindProcess(childPID)
	if err == nil && child.Signal(syscall.Signal(0)) == nil {
		status.Running = true
		status.Pid = &childPID
	}

	return status
}

// IsProcessRunning checks if a specific process is running
func (c *ComputerUse) IsProcessRunning(req *computeruse.ProcessRequest) (bool, error) {
	c.mu.RLock()
	process, exists := c.processes[req.ProcessName]
	c.mu.RUnlock()
	if !exists {
		return false, fmt.Errorf("process %s not found", req.ProcessName)
	}
	if req.ProcessName == "atspi" {
		return c.cachedA11yAvailable(), nil
	}

	return getProcessStatus(process).Running, nil
}

// RestartProcess restarts a specific process
func (c *ComputerUse) RestartProcess(req *computeruse.ProcessRequest) (*computeruse.Empty, error) {
	c.mu.RLock()
	process, exists := c.processes[req.ProcessName]
	c.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("process %s not found", req.ProcessName)
	}

	// Stop the process first
	c.stopProcess(process)

	// Wait a moment for cleanup
	time.Sleep(1 * time.Second)

	// Start the process again
	go c.startProcess(process)

	return new(computeruse.Empty), nil
}

// GetProcessLogs returns the logs for a specific process
func (c *ComputerUse) GetProcessLogs(req *computeruse.ProcessRequest) (string, error) {
	c.mu.RLock()
	process, exists := c.processes[req.ProcessName]
	c.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("process %s not found", req.ProcessName)
	}

	if process.LogFile == "" {
		return "", fmt.Errorf("no log file configured for process %s", req.ProcessName)
	}

	content, err := os.ReadFile(process.LogFile)
	if err != nil {
		return "", fmt.Errorf("failed to read log file for %s: %v", req.ProcessName, err)
	}

	return string(content), nil
}

// GetProcessErrors returns the error logs for a specific process
func (c *ComputerUse) GetProcessErrors(req *computeruse.ProcessRequest) (string, error) {
	c.mu.RLock()
	process, exists := c.processes[req.ProcessName]
	c.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("process %s not found", req.ProcessName)
	}

	if process.ErrFile == "" {
		return "", fmt.Errorf("no error file configured for process %s", req.ProcessName)
	}

	content, err := os.ReadFile(process.ErrFile)
	if err != nil {
		return "", fmt.Errorf("failed to read error file for %s: %v", req.ProcessName, err)
	}

	return string(content), nil
}

func (c *ComputerUse) GetStatus() (*computeruse.ComputerUseStatusResponse, error) {
	// Get the current process status
	processStatus, err := c.GetProcessStatus()
	if err != nil {
		return &computeruse.ComputerUseStatusResponse{
			Status: "error",
		}, err
	}

	// Check if all required processes are running
	requiredProcesses := []string{"xvfb", "xfce4", "x11vnc", "novnc"}
	allRunning := true

	for _, processName := range requiredProcesses {
		if status, exists := processStatus[processName]; !exists || !status.Running {
			allRunning = false
			break
		}
	}

	if allRunning {
		return &computeruse.ComputerUseStatusResponse{
			Status: "active",
		}, nil
	}

	// Check if any processes are running
	anyRunning := false
	for _, status := range processStatus {
		if status.Running {
			anyRunning = true
			break
		}
	}

	if anyRunning {
		return &computeruse.ComputerUseStatusResponse{
			Status: "partial",
		}, nil
	}

	return &computeruse.ComputerUseStatusResponse{
		Status: "inactive",
	}, nil
}
