// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/daytonaio/daemon/pkg/toolbox/computeruse"
)

const (
	defaultBrowserCDPPort = 9222
	defaultBrowserTimeout = 15 * time.Second
	defaultBrowserPoll    = 200 * time.Millisecond
)

type browserManager struct {
	mu           sync.Mutex
	cmd          *exec.Cmd
	done         chan error
	port         int
	profileDir   string
	binary       string
	localCDPURL  string
	httpClient   *http.Client
	readyTimeout time.Duration
	pollInterval time.Duration
	env          map[string]string
}

type cdpVersion struct {
	WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
}

func (c *ComputerUse) ensureBrowserManager() (*browserManager, error) {
	c.browserMu.Lock()
	defer c.browserMu.Unlock()

	if c.browser != nil {
		return c.browser, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}
	configDir := c.configDir
	if configDir == "" {
		configDir = filepath.Join(homeDir, ".daytona", "computeruse")
	}

	binary, err := locateChromium()
	if err != nil {
		return nil, err
	}

	c.browser = &browserManager{
		port:         browserPort(),
		profileDir:   filepath.Join(configDir, "browser-profile"),
		binary:       binary,
		httpClient:   &http.Client{Timeout: 2 * time.Second},
		readyTimeout: defaultBrowserTimeout,
		pollInterval: defaultBrowserPoll,
		env:          desktopEnv(homeDir),
	}
	return c.browser, nil
}

func locateChromium() (string, error) {
	for _, envKey := range []string{"DAYTONA_BROWSER_PATH", "CHROME_PATH"} {
		if path := os.Getenv(envKey); path != "" {
			if _, err := os.Stat(path); err == nil {
				return path, nil
			}
		}
	}

	for _, name := range []string{"chromium", "chromium-browser", "google-chrome", "google-chrome-stable", "chrome"} {
		if path, err := exec.LookPath(name); err == nil {
			return path, nil
		}
	}

	for _, path := range []string{
		"/usr/bin/chromium",
		"/usr/bin/chromium-browser",
		"/usr/bin/google-chrome",
		"/usr/bin/google-chrome-stable",
	} {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("chromium binary not found")
}

func browserPort() int {
	if raw := os.Getenv("DAYTONA_BROWSER_CDP_PORT"); raw != "" {
		if port, err := strconv.Atoi(raw); err == nil && port > 0 {
			return port
		}
	}
	return defaultBrowserCDPPort
}

func desktopEnv(homeDir string) map[string]string {
	display := os.Getenv("DISPLAY")
	if display == "" {
		display = ":0"
	}

	user := os.Getenv("VNC_USER")
	if user == "" {
		user = os.Getenv("DAYTONA_SANDBOX_USER")
		if user == "" {
			user = os.Getenv("USER")
		}
	}

	return map[string]string{
		"DISPLAY":                  display,
		"HOME":                     homeDir,
		"USER":                     user,
		"DBUS_SESSION_BUS_ADDRESS": os.Getenv("DBUS_SESSION_BUS_ADDRESS"),
		"DBUS_SESSION_BUS_PID":     os.Getenv("DBUS_SESSION_BUS_PID"),
	}
}

func (b *browserManager) getCDP(externalBaseURL string) (*computeruse.BrowserCDPResponse, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.isRunningLocked() && b.localCDPURL != "" {
		return b.response(externalBaseURL), nil
	}
	if err := b.startLocked(); err != nil {
		return nil, err
	}
	return b.response(externalBaseURL), nil
}

func (b *browserManager) status(externalBaseURL string) *computeruse.BrowserStatusResponse {
	b.mu.Lock()
	defer b.mu.Unlock()

	running := b.isRunningLocked()
	status := "stopped"
	var pid *int
	if running {
		status = "running"
		p := b.cmd.Process.Pid
		pid = &p
	}

	response := &computeruse.BrowserStatusResponse{
		Status:                    status,
		Running:                   running,
		Pid:                       pid,
		Port:                      b.port,
		LocalWebSocketDebuggerURL: b.localCDPURL,
		ProxyPath:                 computeruse.BrowserProxyPath(b.localCDPURL),
	}
	response.WebSocketDebuggerURL = computeruse.RewriteBrowserCDPURL(b.localCDPURL, externalBaseURL)
	return response
}

func (b *browserManager) stop() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.stopLocked()
}

func (b *browserManager) startLocked() error {
	b.stopLocked()

	if err := os.MkdirAll(b.profileDir, 0755); err != nil {
		return fmt.Errorf("failed to create browser profile directory: %w", err)
	}

	cmd := exec.Command(b.binary, b.args()...)
	cmd.Env = os.Environ()
	for key, value := range b.env {
		if value != "" {
			cmd.Env = append(cmd.Env, key+"="+value)
		}
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start browser: %w", err)
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	b.cmd = cmd
	b.done = done

	localURL, err := b.waitForReady(done)
	if err != nil {
		b.stopLocked()
		return err
	}
	b.localCDPURL = localURL
	return nil
}

func (b *browserManager) args() []string {
	return []string{
		"--remote-debugging-address=127.0.0.1",
		"--remote-debugging-port=" + strconv.Itoa(b.port),
		"--user-data-dir=" + b.profileDir,
		"--no-first-run",
		"--no-default-browser-check",
		"--disable-session-crashed-bubble",
		"--disable-infobars",
		"--disable-background-networking",
		"--password-store=basic",
		"--use-mock-keychain",
		"--no-sandbox",
		"about:blank",
	}
}

func (b *browserManager) waitForReady(done <-chan error) (string, error) {
	deadline := time.Now().Add(b.readyTimeout)
	var lastErr error

	for {
		select {
		case err := <-done:
			if err != nil {
				return "", fmt.Errorf("browser exited before CDP became ready: %w", err)
			}
			return "", fmt.Errorf("browser exited before CDP became ready")
		default:
		}

		localURL, err := b.fetchCDPURL()
		if err == nil {
			return localURL, nil
		}
		lastErr = err

		if time.Now().After(deadline) {
			return "", fmt.Errorf("browser CDP did not become ready: %w", lastErr)
		}
		time.Sleep(b.pollInterval)
	}
}

func (b *browserManager) fetchCDPURL() (string, error) {
	resp, err := b.httpClient.Get("http://127.0.0.1:" + strconv.Itoa(b.port) + "/json/version")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected CDP status: %d", resp.StatusCode)
	}

	var version cdpVersion
	if err := json.NewDecoder(resp.Body).Decode(&version); err != nil {
		return "", err
	}
	if version.WebSocketDebuggerURL == "" {
		return "", fmt.Errorf("CDP version response did not include webSocketDebuggerUrl")
	}
	return version.WebSocketDebuggerURL, nil
}

func (b *browserManager) response(externalBaseURL string) *computeruse.BrowserCDPResponse {
	return &computeruse.BrowserCDPResponse{
		WebSocketDebuggerURL:      computeruse.RewriteBrowserCDPURL(b.localCDPURL, externalBaseURL),
		LocalWebSocketDebuggerURL: b.localCDPURL,
		ProxyPath:                 computeruse.BrowserProxyPath(b.localCDPURL),
		Port:                      b.port,
	}
}

func (b *browserManager) isRunningLocked() bool {
	return b.cmd != nil && b.cmd.Process != nil && b.cmd.Process.Signal(syscall.Signal(0)) == nil
}

func (b *browserManager) stopLocked() {
	if b.cmd != nil && b.cmd.Process != nil {
		_ = b.cmd.Process.Kill()
	}
	if b.done != nil {
		select {
		case <-b.done:
		case <-time.After(2 * time.Second):
		}
	}
	b.cmd = nil
	b.done = nil
	b.localCDPURL = ""
}

func (c *ComputerUse) GetBrowserCDP(req *computeruse.BrowserCDPRequest) (*computeruse.BrowserCDPResponse, error) {
	manager, err := c.ensureBrowserManager()
	if err != nil {
		return nil, err
	}
	externalBaseURL := ""
	if req != nil {
		externalBaseURL = req.ExternalBaseURL
	}
	return manager.getCDP(externalBaseURL)
}

func (c *ComputerUse) GetBrowserStatus() (*computeruse.BrowserStatusResponse, error) {
	manager, err := c.ensureBrowserManager()
	if err != nil {
		return nil, err
	}
	return manager.status(""), nil
}

func (c *ComputerUse) StopBrowser() (*computeruse.Empty, error) {
	c.browserMu.Lock()
	manager := c.browser
	c.browserMu.Unlock()

	if manager != nil {
		manager.stop()
	}
	return new(computeruse.Empty), nil
}
