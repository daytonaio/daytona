// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/daytonaio/daemon/pkg/toolbox/computeruse"
	log "github.com/sirupsen/logrus"
)

type Process struct {
	Name        string
	Command     string
	Args        []string
	User        string
	Priority    int
	Env         map[string]string
	LogFile     string
	ErrFile     string
	AutoRestart bool
	cmd         *exec.Cmd
	ctx         context.Context
	cancel      context.CancelFunc
	mu          sync.Mutex
	running     bool
}

type ComputerUse struct {
	processes map[string]*Process
	mu        sync.RWMutex
	configDir string
}

var _ computeruse.IComputerUse = &ComputerUse{}

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

	c.initializeProcesses(homeDir)

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
	c.startAllProcesses()

	// Check process status after starting
	status, err := c.GetProcessStatus()
	if err != nil {
		return nil, fmt.Errorf("failed to get process status after start: %v", err)
	}

	// Check if all required processes are running
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

func (c *ComputerUse) initializeProcesses(homeDir string) {
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

	// Process 1: Xvfb (X Virtual Framebuffer)
	c.processes["xvfb"] = &Process{
		Name:        "xvfb",
		Command:     "/usr/bin/Xvfb",
		Args:        []string{display, "-screen", "0", vncResolution + "x24"},
		User:        user,
		Priority:    100,
		AutoRestart: true,
		Env: map[string]string{
			"DISPLAY": display,
		},
		LogFile: filepath.Join(c.configDir, "xvfb.log"),
		ErrFile: filepath.Join(c.configDir, "xvfb.err"),
	}

	// Process 2: xfce4 (Desktop Environment)
	c.processes["xfce4"] = &Process{
		Name:        "xfce4",
		Command:     "/usr/bin/startxfce4",
		Args:        []string{},
		User:        user,
		Priority:    200,
		AutoRestart: true,
		Env: map[string]string{
			"DISPLAY":                  display,
			"HOME":                     homeDir,
			"USER":                     user,
			"DBUS_SESSION_BUS_ADDRESS": dbusAddress,
		},
		LogFile: filepath.Join(c.configDir, "xfce4.log"),
		ErrFile: filepath.Join(c.configDir, "xfce4.err"),
	}

	// Process 3: x11vnc (VNC Server)
	c.processes["x11vnc"] = &Process{
		Name:        "x11vnc",
		Command:     "/usr/bin/x11vnc",
		Args:        []string{"-display", display, "-forever", "-shared", "-rfbport", vncPort},
		User:        user,
		Priority:    300,
		AutoRestart: true,
		Env: map[string]string{
			"DISPLAY": display,
		},
		LogFile: filepath.Join(c.configDir, "x11vnc.log"),
		ErrFile: filepath.Join(c.configDir, "x11vnc.err"),
	}

	// Process 4: novnc (Web-based VNC client)
	// Determine the best available NoVNC command with fallback options
	var novncCommand string
	var novncArgs []string

	// Priority 1: Try launch.sh (modern NoVNC with enhanced features)
	if _, err := os.Stat("/usr/share/novnc/utils/launch.sh"); err == nil {
		novncCommand = "/usr/share/novnc/utils/launch.sh"
		novncArgs = []string{"--vnc", "localhost:" + vncPort, "--listen", noVncPort}
		log.Infof("Using NoVNC launch.sh (recommended)")
	} else if _, err := os.Stat("/usr/share/novnc/utils/novnc_proxy"); err == nil {
		// Priority 2: Try novnc_proxy (legacy NoVNC script)
		novncCommand = "/usr/share/novnc/utils/novnc_proxy"
		novncArgs = []string{"--vnc", "localhost:" + vncPort, "--listen", noVncPort}
		log.Infof("Using NoVNC novnc_proxy (legacy)")
	} else {
		// Priority 3: Fallback to direct websockify (always available)
		novncCommand = "websockify"
		novncArgs = []string{"--web=/usr/share/novnc/", noVncPort, "localhost:" + vncPort}
		log.Infof("Using direct websockify (fallback)")
	}

	c.processes["novnc"] = &Process{
		Name:        "novnc",
		Command:     novncCommand,
		Args:        novncArgs,
		User:        user,
		Priority:    400,
		AutoRestart: true,
		Env: map[string]string{
			"DISPLAY": display,
		},
		LogFile: filepath.Join(c.configDir, "novnc.log"),
		ErrFile: filepath.Join(c.configDir, "novnc.err"),
	}
}

func (c *ComputerUse) startAllProcesses() {
	// Sort processes by priority and start them
	processes := c.getProcessesByPriority()

	for _, process := range processes {
		go c.startProcess(process)
		// Wait a bit between starting processes to ensure proper initialization
		time.Sleep(2 * time.Second)
	}
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
	process.mu.Lock()
	if process.running {
		process.mu.Unlock()
		return
	}
	process.running = true
	process.mu.Unlock()

	for {
		log.Infof("Starting process: %s", process.Name)

		// Create context for the process
		process.ctx, process.cancel = context.WithCancel(context.Background())

		// Create command
		process.cmd = exec.CommandContext(process.ctx, process.Command, process.Args...)

		// Set environment variables
		if len(process.Env) > 0 {
			process.cmd.Env = os.Environ()
			for key, value := range process.Env {
				process.cmd.Env = append(process.cmd.Env, fmt.Sprintf("%s=%s", key, value))
			}
		}

		// Set up logging
		if process.LogFile != "" {
			logFile, err := os.OpenFile(process.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				log.Errorf("Failed to open log file for %s: %v", process.Name, err)
			} else {
				process.cmd.Stdout = logFile
				defer logFile.Close()
			}
		}

		if process.ErrFile != "" {
			errFile, err := os.OpenFile(process.ErrFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				log.Errorf("Failed to open error file for %s: %v", process.Name, err)
			} else {
				process.cmd.Stderr = errFile
				defer errFile.Close()
			}
		}

		// Start the process
		err := process.cmd.Start()
		if err != nil {
			log.Errorf("Failed to start process %s: %v", process.Name, err)
			if !process.AutoRestart {
				break
			}
			time.Sleep(5 * time.Second)
			continue
		}

		log.Infof("Process %s started with PID: %d", process.Name, process.cmd.Process.Pid)

		// Wait for the process to finish
		err = process.cmd.Wait()
		if err != nil {
			log.Errorf("Process %s exited with error: %v", process.Name, err)
		} else {
			log.Infof("Process %s exited normally", process.Name)
		}

		// Check if we should restart
		if !process.AutoRestart {
			break
		}

		log.Infof("Restarting process %s in 5 seconds...", process.Name)
		time.Sleep(5 * time.Second)
	}

	process.mu.Lock()
	process.running = false
	process.mu.Unlock()
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

	return new(computeruse.Empty), nil
}

func (c *ComputerUse) stopProcess(process *Process) {
	process.mu.Lock()
	defer process.mu.Unlock()

	if !process.running {
		return
	}

	log.Infof("Stopping process: %s", process.Name)

	// Cancel the context to stop the process
	if process.cancel != nil {
		process.cancel()
	}

	// Kill the process if it's still running
	if process.cmd != nil && process.cmd.Process != nil {
		err := process.cmd.Process.Kill()
		if err != nil {
			log.Errorf("Failed to kill process %s: %v", process.Name, err)
		}
	}

	process.running = false
}

func (c *ComputerUse) GetProcessStatus() (map[string]computeruse.ProcessStatus, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	status := make(map[string]computeruse.ProcessStatus)
	for name, process := range c.processes {
		process.mu.Lock()
		processStatus := computeruse.ProcessStatus{
			Running:     false,
			Priority:    process.Priority,
			AutoRestart: process.AutoRestart,
		}

		if process.cmd != nil && process.cmd.Process != nil {
			// Check if the process is still alive
			if err := process.cmd.Process.Signal(syscall.Signal(0)); err == nil {
				processStatus.Running = true
				processStatus.Pid = &process.cmd.Process.Pid
			}
		}

		process.mu.Unlock()
		status[name] = processStatus
	}

	return status, nil
}

// IsProcessRunning checks if a specific process is running
func (c *ComputerUse) IsProcessRunning(req *computeruse.ProcessRequest) (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	process, exists := c.processes[req.ProcessName]
	if !exists {
		return false, fmt.Errorf("process %s not found", req.ProcessName)
	}

	process.mu.Lock()
	defer process.mu.Unlock()
	return process.running, nil
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

func (c *ComputerUse) GetStatus() (*computeruse.StatusResponse, error) {
	// Get the current process status
	processStatus, err := c.GetProcessStatus()
	if err != nil {
		return &computeruse.StatusResponse{
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
		return &computeruse.StatusResponse{
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
		return &computeruse.StatusResponse{
			Status: "partial",
		}, nil
	}

	return &computeruse.StatusResponse{
		Status: "inactive",
	}, nil
}
