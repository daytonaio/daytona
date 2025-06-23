// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"net/http"
	"net/rpc"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/go-plugin"
)

// PluginInterface defines the interface that the computeruse plugin must implement
type IComputerUse interface {
	// Process management
	Initialize() (*Empty, error)
	Start() (*Empty, error)
	Stop() (*Empty, error)
	GetProcessStatus() (map[string]ProcessStatus, error)
	IsProcessRunning(req *ProcessRequest) (bool, error)
	RestartProcess(req *ProcessRequest) (*Empty, error)
	GetProcessLogs(req *ProcessRequest) (string, error)
	GetProcessErrors(req *ProcessRequest) (string, error)

	// Screenshot methods
	TakeScreenshot(*ComputerUseRequest) (*Empty, error)
	TakeRegionScreenshot(*ComputerUseRequest) (*Empty, error)
	TakeCompressedScreenshot(*ComputerUseRequest) (*Empty, error)
	TakeCompressedRegionScreenshot(*ComputerUseRequest) (*Empty, error)

	// Mouse control methods
	GetMousePosition(*ComputerUseRequest) (*Empty, error)
	MoveMouse(*ComputerUseRequest) (*Empty, error)
	Click(*ComputerUseRequest) (*Empty, error)
	Drag(*ComputerUseRequest) (*Empty, error)
	Scroll(*ComputerUseRequest) (*Empty, error)

	// Keyboard control methods
	TypeText(*ComputerUseRequest) (*Empty, error)
	PressKey(*ComputerUseRequest) (*Empty, error)
	PressHotkey(*ComputerUseRequest) (*Empty, error)

	// Display info methods
	GetDisplayInfo(*ComputerUseRequest) (*Empty, error)
	GetWindows(*ComputerUseRequest) (*Empty, error)

	// Status method
	GetStatus(*ComputerUseRequest) (*Empty, error)
}

type ComputerUsePlugin struct {
	Impl IComputerUse
}

type ComputerUseRequest struct {
	RequestContext *gin.Context
}

type ProcessStatus struct {
	Running     bool
	Priority    int
	AutoRestart bool
	Pid         *int
}

type ProcessRequest struct {
	ProcessName string
}

type Empty struct{}

func (p *ComputerUsePlugin) Server(*plugin.MuxBroker) (any, error) {
	return &ComputerUseRPCServer{Impl: p.Impl}, nil
}

func (p *ComputerUsePlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (any, error) {
	return &ComputerUseRPCClient{client: c}, nil
}

func WrapRequest(fn func(*ComputerUseRequest) (*Empty, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		request := &ComputerUseRequest{
			RequestContext: c,
		}
		_, err := fn(request)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
	}
}
