// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	wire "github.com/daytonaio/daemon/pkg/toolbox/computeruse"
	"github.com/gin-gonic/gin"
)

func TestLiveAccessibilityWaitHTTP(t *testing.T) {
	if os.Getenv("DAYTONA_LIVE_A11Y_SMOKE") != "1" {
		t.Skip("set DAYTONA_LIVE_A11Y_SMOKE=1 to run the live AT-SPI smoke")
	}

	startLiveA11yDesktop(t)
	startGTKSmokeApp(t)

	cu := &ComputerUse{}
	button, err := cu.WaitAccessibility(&wire.AccessibilityWaitRequest{
		Condition: "exists",
		Query: &wire.FindAccessibilityNodesRequest{
			Scope:     "all",
			Name:      "Create alert",
			NameMatch: "exact",
		},
		TimeoutMs:      10000,
		PollIntervalMs: 100,
	})
	if err != nil {
		t.Fatalf("wait for button: %v", err)
	}
	if !button.Matched || len(button.Matches) == 0 {
		t.Fatalf("button did not appear: %+v; visible nodes: %s", button, liveNodeSummary(t, cu))
	}
	if _, err := cu.InvokeAccessibilityNode(&wire.AccessibilityInvokeRequest{ID: button.Matches[0].ID}); err != nil {
		t.Fatalf("invoke button: %v", err)
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/computeruse/a11y/wait", wire.WrapWaitAccessibilityHandler(cu.WaitAccessibility))

	reqBody := bytes.NewBufferString(`{
		"condition":"exists",
		"query":{"scope":"all","name":"Saved","nameMatch":"exact"},
		"timeoutMs":10000,
		"pollIntervalMs":100
	}`)
	req := httptest.NewRequest(http.MethodPost, "/computeruse/a11y/wait", reqBody)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body.String())
	}

	var resp wire.AccessibilityWaitResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !resp.Matched || resp.TimedOut || len(resp.Matches) == 0 || resp.Matches[0].Name != "Saved" {
		t.Fatalf("unexpected wait response: %+v; visible nodes: %s", resp, liveNodeSummary(t, cu))
	}
}

func liveNodeSummary(t *testing.T, cu *ComputerUse) string {
	t.Helper()
	nodes, _, err := cu.findAccessibilityNodes(A11yScopeAll, 0, A11yFilter{}, 20)
	if err != nil {
		return "find failed: " + err.Error()
	}
	parts := make([]string, 0, len(nodes))
	for _, node := range nodes {
		parts = append(parts, fmt.Sprintf("%s:%q", node.Role, node.Name))
	}
	return strings.Join(parts, ", ")
}

func startLiveA11yDesktop(t *testing.T) {
	t.Helper()
	requireCommand(t, "Xvfb")
	requireCommand(t, "dbus-launch")

	display := ":99"
	unsetEnv(t, "AT_SPI_BUS_ADDRESS")
	t.Setenv("DISPLAY", display)
	t.Setenv("GDK_BACKEND", "x11")
	t.Setenv("GTK_MODULES", "gail:atk-bridge")
	t.Setenv("QT_ACCESSIBILITY", "1")
	t.Setenv("NO_AT_BRIDGE", "0")

	startProcess(t, exec.Command("Xvfb", display, "-screen", "0", "1280x720x24"))

	out, err := exec.Command("dbus-launch", "--sh-syntax").Output()
	if err != nil {
		t.Fatalf("start dbus-launch: %v", err)
	}
	for _, stmt := range strings.Split(string(out), ";") {
		parts := strings.SplitN(strings.TrimSpace(stmt), "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.Trim(strings.TrimSpace(parts[1]), "'")
		if strings.HasPrefix(key, "DBUS_SESSION_BUS_") {
			t.Setenv(key, value)
		}
	}
	if pid := os.Getenv("DBUS_SESSION_BUS_PID"); pid != "" {
		t.Cleanup(func() { _ = exec.Command("kill", pid).Run() })
	}

	atspi := atspiLauncherPath(t)
	startProcess(t, exec.Command(atspi, "--launch-immediately"))
	time.Sleep(500 * time.Millisecond)
}

func startGTKSmokeApp(t *testing.T) {
	t.Helper()
	python := gtkPythonPath(t)

	app := filepath.Join(t.TempDir(), "a11y_smoke.py")
	err := os.WriteFile(app, []byte(`
import gi
gi.require_version("Gtk", "3.0")
from gi.repository import Gtk

win = Gtk.Window(title="Daytona A11y Smoke")
win.connect("destroy", Gtk.main_quit)
box = Gtk.Box(orientation=Gtk.Orientation.VERTICAL, spacing=8)
button = Gtk.Button(label="Create alert")

def clicked(_button):
    label = Gtk.Label(label="Saved")
    box.pack_start(label, True, True, 0)
    label.show()

button.connect("clicked", clicked)
box.pack_start(button, True, True, 0)
win.add(box)
win.show_all()
Gtk.main()
`), 0600)
	if err != nil {
		t.Fatalf("write GTK app: %v", err)
	}

	cmd := exec.Command(python, app)
	cmd.Env = os.Environ()
	startProcess(t, cmd)
	time.Sleep(500 * time.Millisecond)
}

func startProcess(t *testing.T, cmd *exec.Cmd) {
	t.Helper()
	cmd.Env = os.Environ()
	if err := cmd.Start(); err != nil {
		t.Fatalf("start %s: %v", cmd.Path, err)
	}
	t.Cleanup(func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	})
}

func requireCommand(t *testing.T, name string) {
	t.Helper()
	if _, err := exec.LookPath(name); err != nil {
		t.Fatalf("%s is required for live a11y smoke: %v", name, err)
	}
}

func gtkPythonPath(t *testing.T) string {
	t.Helper()
	for _, path := range []string{"/usr/bin/python3", "python3"} {
		if _, err := exec.LookPath(path); err != nil {
			continue
		}
		if err := exec.Command(path, "-c", `import gi; gi.require_version("Gtk", "3.0")`).Run(); err == nil {
			return path
		}
	}
	t.Fatal("python3 with GTK gi bindings is required for live a11y smoke")
	return ""
}

func unsetEnv(t *testing.T, key string) {
	t.Helper()
	value, ok := os.LookupEnv(key)
	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("unset %s: %v", key, err)
	}
	t.Cleanup(func() {
		if ok {
			_ = os.Setenv(key, value)
			return
		}
		_ = os.Unsetenv(key)
	})
}

func atspiLauncherPath(t *testing.T) string {
	t.Helper()
	for _, path := range []string{
		"/usr/libexec/at-spi-bus-launcher",
		"/usr/lib/at-spi2-core/at-spi-bus-launcher",
	} {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	if path, err := exec.LookPath("at-spi-bus-launcher"); err == nil {
		return path
	}
	t.Fatal(fmt.Errorf("at-spi-bus-launcher is required for live a11y smoke"))
	return ""
}
