package plugin_loader

import (
	"github.com/gin-gonic/gin"
)

// PluginInterface defines the interface that the computeruse plugin must implement
type PluginInterface interface {
	// Process management
	Start() error
	Stop()
	GetProcessStatus() map[string]interface{}
	IsProcessRunning(processName string) bool
	RestartProcess(processName string) error
	GetProcessLogs(processName string) (string, error)
	GetProcessErrors(processName string) (string, error)

	// Screenshot methods
	TakeScreenshot(c *gin.Context)
	TakeRegionScreenshot(c *gin.Context)
	TakeCompressedScreenshot(c *gin.Context)
	TakeCompressedRegionScreenshot(c *gin.Context)

	// Mouse control methods
	GetMousePosition(c *gin.Context)
	MoveMouse(c *gin.Context)
	Click(c *gin.Context)
	Drag(c *gin.Context)
	Scroll(c *gin.Context)

	// Keyboard control methods
	TypeText(c *gin.Context)
	PressKey(c *gin.Context)
	PressHotkey(c *gin.Context)

	// Display info methods
	GetDisplayInfo(c *gin.Context)
	GetWindows(c *gin.Context)

	// Status method
	GetStatus(c *gin.Context)
}

// PluginSymbolName is the name of the exported symbol in the plugin
const PluginSymbolName = "ComputerUsePlugin"
