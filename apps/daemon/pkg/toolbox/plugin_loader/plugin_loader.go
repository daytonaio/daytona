package plugin_loader

import (
	"fmt"
	"net/http"
	"plugin"
	"sync"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// PluginLoader handles loading and managing the computeruse plugin
type PluginLoader struct {
	mu         sync.RWMutex
	plugin     *plugin.Plugin
	pluginImpl PluginInterface
	loaded     bool
	loadError  error
	loadOnce   sync.Once
	pluginPath string
}

// NewPluginLoader creates a new plugin loader
func NewPluginLoader(pluginPath string) *PluginLoader {
	return &PluginLoader{
		pluginPath: pluginPath,
	}
}

// ensureLoaded ensures the plugin is loaded
func (p *PluginLoader) ensureLoaded() error {
	p.loadOnce.Do(func() {
		p.mu.Lock()
		defer p.mu.Unlock()

		if p.loaded {
			return
		}

		log.Infof("Attempting to load computer use plugin from: %s", p.pluginPath)

		// Try to load the plugin
		plug, err := plugin.Open(p.pluginPath)
		if err != nil {
			p.loadError = fmt.Errorf("failed to load plugin %s: %w", p.pluginPath, err)
			log.Warnf("Computer use plugin not available: %v", p.loadError)
			return
		}

		log.Infof("Plugin file loaded successfully, looking for symbol: %s", PluginSymbolName)

		// Look up the exported symbol
		sym, err := plug.Lookup(PluginSymbolName)
		if err != nil {
			p.loadError = fmt.Errorf("failed to find symbol %s in plugin: %w", PluginSymbolName, err)
			log.Warnf("Computer use plugin symbol not found: %v", p.loadError)
			return
		}

		log.Infof("Symbol found, type: %T", sym)

		// Type assert to our interface
		impl, ok := sym.(PluginInterface)
		if !ok {
			p.loadError = fmt.Errorf("plugin does not implement PluginInterface, got type: %T", sym)
			log.Warnf("Computer use plugin interface mismatch: %v", p.loadError)
			return
		}

		p.plugin = plug
		p.pluginImpl = impl
		p.loaded = true
		log.Info("Computer use plugin loaded successfully")
	})

	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.loadError
}

// isAvailable checks if the plugin is available
func (p *PluginLoader) isAvailable() bool {
	return p.ensureLoaded() == nil
}

// getPluginImpl returns the plugin implementation if available
func (p *PluginLoader) getPluginImpl() (PluginInterface, error) {
	if err := p.ensureLoaded(); err != nil {
		return nil, err
	}
	return p.pluginImpl, nil
}

// errorResponse returns a standardized error response for unavailable functionality
func (p *PluginLoader) errorResponse(c *gin.Context) {
	errorDetails := "Plugin or X11 dependencies are missing. Please install required libraries or use a system with X11 support."
	solution := "Install X11 libraries: sudo apt-get install libx11-6 libxrandr2 libxext6 libxrender1 libxfixes3"

	// Include specific error details if available
	if p.loadError != nil {
		errorDetails = fmt.Sprintf("Plugin loading failed: %v", p.loadError)
	}

	c.JSON(http.StatusServiceUnavailable, gin.H{
		"error":         "Computer use functionality not available",
		"details":       errorDetails,
		"solution":      solution,
		"plugin_path":   p.pluginPath,
		"plugin_loaded": p.loaded,
	})
}

// Process management methods

func (p *PluginLoader) Start() error {
	impl, err := p.getPluginImpl()
	if err != nil {
		return fmt.Errorf("computer use not available: %w", err)
	}
	return impl.Start()
}

func (p *PluginLoader) Stop() {
	if impl, err := p.getPluginImpl(); err == nil {
		impl.Stop()
	}
}

func (p *PluginLoader) GetProcessStatus() map[string]interface{} {
	if impl, err := p.getPluginImpl(); err == nil {
		return impl.GetProcessStatus()
	}
	return map[string]interface{}{
		"error":   "Computer use not available",
		"details": p.loadError.Error(),
	}
}

func (p *PluginLoader) IsProcessRunning(processName string) bool {
	if impl, err := p.getPluginImpl(); err == nil {
		return impl.IsProcessRunning(processName)
	}
	return false
}

func (p *PluginLoader) RestartProcess(processName string) error {
	impl, err := p.getPluginImpl()
	if err != nil {
		return fmt.Errorf("computer use not available: %w", err)
	}
	return impl.RestartProcess(processName)
}

func (p *PluginLoader) GetProcessLogs(processName string) (string, error) {
	impl, err := p.getPluginImpl()
	if err != nil {
		return "", fmt.Errorf("computer use not available: %w", err)
	}
	return impl.GetProcessLogs(processName)
}

func (p *PluginLoader) GetProcessErrors(processName string) (string, error) {
	impl, err := p.getPluginImpl()
	if err != nil {
		return "", fmt.Errorf("computer use not available: %w", err)
	}
	return impl.GetProcessErrors(processName)
}

// HTTP Handlers

func (p *PluginLoader) TakeScreenshot(c *gin.Context) {
	if impl, err := p.getPluginImpl(); err == nil {
		impl.TakeScreenshot(c)
	} else {
		p.errorResponse(c)
	}
}

func (p *PluginLoader) TakeRegionScreenshot(c *gin.Context) {
	if impl, err := p.getPluginImpl(); err == nil {
		impl.TakeRegionScreenshot(c)
	} else {
		p.errorResponse(c)
	}
}

func (p *PluginLoader) TakeCompressedScreenshot(c *gin.Context) {
	if impl, err := p.getPluginImpl(); err == nil {
		impl.TakeCompressedScreenshot(c)
	} else {
		p.errorResponse(c)
	}
}

func (p *PluginLoader) TakeCompressedRegionScreenshot(c *gin.Context) {
	if impl, err := p.getPluginImpl(); err == nil {
		impl.TakeCompressedRegionScreenshot(c)
	} else {
		p.errorResponse(c)
	}
}

func (p *PluginLoader) GetMousePosition(c *gin.Context) {
	if impl, err := p.getPluginImpl(); err == nil {
		impl.GetMousePosition(c)
	} else {
		p.errorResponse(c)
	}
}

func (p *PluginLoader) MoveMouse(c *gin.Context) {
	if impl, err := p.getPluginImpl(); err == nil {
		impl.MoveMouse(c)
	} else {
		p.errorResponse(c)
	}
}

func (p *PluginLoader) Click(c *gin.Context) {
	if impl, err := p.getPluginImpl(); err == nil {
		impl.Click(c)
	} else {
		p.errorResponse(c)
	}
}

func (p *PluginLoader) Drag(c *gin.Context) {
	if impl, err := p.getPluginImpl(); err == nil {
		impl.Drag(c)
	} else {
		p.errorResponse(c)
	}
}

func (p *PluginLoader) Scroll(c *gin.Context) {
	if impl, err := p.getPluginImpl(); err == nil {
		impl.Scroll(c)
	} else {
		p.errorResponse(c)
	}
}

func (p *PluginLoader) TypeText(c *gin.Context) {
	if impl, err := p.getPluginImpl(); err == nil {
		impl.TypeText(c)
	} else {
		p.errorResponse(c)
	}
}

func (p *PluginLoader) PressKey(c *gin.Context) {
	if impl, err := p.getPluginImpl(); err == nil {
		impl.PressKey(c)
	} else {
		p.errorResponse(c)
	}
}

func (p *PluginLoader) PressHotkey(c *gin.Context) {
	if impl, err := p.getPluginImpl(); err == nil {
		impl.PressHotkey(c)
	} else {
		p.errorResponse(c)
	}
}

func (p *PluginLoader) GetDisplayInfo(c *gin.Context) {
	if impl, err := p.getPluginImpl(); err == nil {
		impl.GetDisplayInfo(c)
	} else {
		p.errorResponse(c)
	}
}

func (p *PluginLoader) GetWindows(c *gin.Context) {
	if impl, err := p.getPluginImpl(); err == nil {
		impl.GetWindows(c)
	} else {
		p.errorResponse(c)
	}
}

// TestPluginFunctionality tests if the plugin can perform basic operations
func (p *PluginLoader) TestPluginFunctionality() map[string]interface{} {
	result := map[string]interface{}{
		"plugin_loaded": p.loaded,
		"plugin_path":   p.pluginPath,
	}

	if p.loadError != nil {
		result["error"] = p.loadError.Error()
		return result
	}

	if !p.loaded {
		result["error"] = "Plugin not loaded"
		return result
	}

	// Test basic functionality
	if impl, err := p.getPluginImpl(); err == nil {
		// Try to get process status as a test
		status := impl.GetProcessStatus()
		result["process_status"] = status
		result["status"] = "Plugin functional"
	} else {
		result["error"] = fmt.Sprintf("Failed to get plugin implementation: %v", err)
	}

	return result
}

// GetStatus provides detailed status information
func (p *PluginLoader) GetStatus(c *gin.Context) {
	if impl, err := p.getPluginImpl(); err == nil {
		impl.GetStatus(c)
	} else {
		// Return detailed status information
		status := p.TestPluginFunctionality()
		c.JSON(http.StatusOK, status)
	}
}
