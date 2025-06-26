// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"net/rpc"
)

type ComputerUseRPCClient struct {
	client *rpc.Client
}

// Type check
var _ IComputerUse = &ComputerUseRPCClient{}

// Process management methods
func (m *ComputerUseRPCClient) Initialize() (*Empty, error) {
	err := m.client.Call("Plugin.Initialize", new(any), new(Empty))
	return new(Empty), err
}

func (m *ComputerUseRPCClient) Start() (*Empty, error) {
	err := m.client.Call("Plugin.Start", new(any), new(Empty))
	return new(Empty), err
}

func (m *ComputerUseRPCClient) Stop() (*Empty, error) {
	err := m.client.Call("Plugin.Stop", new(any), new(Empty))
	return new(Empty), err
}

func (m *ComputerUseRPCClient) GetProcessStatus() (map[string]ProcessStatus, error) {
	resp := map[string]ProcessStatus{}
	err := m.client.Call("Plugin.GetProcessStatus", new(any), &resp)
	return resp, err
}

func (m *ComputerUseRPCClient) IsProcessRunning(req *ProcessRequest) (bool, error) {
	var resp bool
	err := m.client.Call("Plugin.IsProcessRunning", req, &resp)
	return resp, err
}

func (m *ComputerUseRPCClient) RestartProcess(req *ProcessRequest) (*Empty, error) {
	err := m.client.Call("Plugin.RestartProcess", req, new(Empty))
	return new(Empty), err
}

func (m *ComputerUseRPCClient) GetProcessLogs(req *ProcessRequest) (string, error) {
	var resp string
	err := m.client.Call("Plugin.GetProcessLogs", req, &resp)
	return resp, err
}

func (m *ComputerUseRPCClient) GetProcessErrors(req *ProcessRequest) (string, error) {
	var resp string
	err := m.client.Call("Plugin.GetProcessErrors", req, &resp)
	return resp, err
}

// Screenshot methods
func (m *ComputerUseRPCClient) TakeScreenshot(request *ScreenshotRequest) (*ScreenshotResponse, error) {
	var resp ScreenshotResponse
	err := m.client.Call("Plugin.TakeScreenshot", request, &resp)
	return &resp, err
}

func (m *ComputerUseRPCClient) TakeRegionScreenshot(request *RegionScreenshotRequest) (*ScreenshotResponse, error) {
	var resp ScreenshotResponse
	err := m.client.Call("Plugin.TakeRegionScreenshot", request, &resp)
	return &resp, err
}

func (m *ComputerUseRPCClient) TakeCompressedScreenshot(request *CompressedScreenshotRequest) (*ScreenshotResponse, error) {
	var resp ScreenshotResponse
	err := m.client.Call("Plugin.TakeCompressedScreenshot", request, &resp)
	return &resp, err
}

func (m *ComputerUseRPCClient) TakeCompressedRegionScreenshot(request *CompressedRegionScreenshotRequest) (*ScreenshotResponse, error) {
	var resp ScreenshotResponse
	err := m.client.Call("Plugin.TakeCompressedRegionScreenshot", request, &resp)
	return &resp, err
}

// Mouse control methods
func (m *ComputerUseRPCClient) GetMousePosition() (*MousePositionResponse, error) {
	var resp MousePositionResponse
	err := m.client.Call("Plugin.GetMousePosition", new(any), &resp)
	return &resp, err
}

func (m *ComputerUseRPCClient) MoveMouse(request *MoveMouseRequest) (*MousePositionResponse, error) {
	var resp MousePositionResponse
	err := m.client.Call("Plugin.MoveMouse", request, &resp)
	return &resp, err
}

func (m *ComputerUseRPCClient) Click(request *ClickRequest) (*MouseClickResponse, error) {
	var resp MouseClickResponse
	err := m.client.Call("Plugin.Click", request, &resp)
	return &resp, err
}

func (m *ComputerUseRPCClient) Drag(request *DragRequest) (*MouseDragResponse, error) {
	var resp MouseDragResponse
	err := m.client.Call("Plugin.Drag", request, &resp)
	return &resp, err
}

func (m *ComputerUseRPCClient) Scroll(request *ScrollRequest) (*ScrollResponse, error) {
	var resp ScrollResponse
	err := m.client.Call("Plugin.Scroll", request, &resp)
	return &resp, err
}

// Keyboard control methods
func (m *ComputerUseRPCClient) TypeText(request *TypeTextRequest) (*Empty, error) {
	err := m.client.Call("Plugin.TypeText", request, new(Empty))
	return new(Empty), err
}

func (m *ComputerUseRPCClient) PressKey(request *PressKeyRequest) (*Empty, error) {
	err := m.client.Call("Plugin.PressKey", request, new(Empty))
	return new(Empty), err
}

func (m *ComputerUseRPCClient) PressHotkey(request *PressHotkeyRequest) (*Empty, error) {
	err := m.client.Call("Plugin.PressHotkey", request, new(Empty))
	return new(Empty), err
}

// Display info methods
func (m *ComputerUseRPCClient) GetDisplayInfo() (*DisplayInfoResponse, error) {
	var resp DisplayInfoResponse
	err := m.client.Call("Plugin.GetDisplayInfo", new(any), &resp)
	return &resp, err
}

func (m *ComputerUseRPCClient) GetWindows() (*WindowsResponse, error) {
	var resp WindowsResponse
	err := m.client.Call("Plugin.GetWindows", new(any), &resp)
	return &resp, err
}

// Status method
func (m *ComputerUseRPCClient) GetStatus() (*StatusResponse, error) {
	var resp StatusResponse
	err := m.client.Call("Plugin.GetStatus", new(any), &resp)
	return &resp, err
}
