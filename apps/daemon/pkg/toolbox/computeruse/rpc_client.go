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
func (m *ComputerUseRPCClient) TakeScreenshot(request *ComputerUseRequest) (*Empty, error) {
	err := m.client.Call("Plugin.TakeScreenshot", request, new(Empty))
	return new(Empty), err
}

func (m *ComputerUseRPCClient) TakeRegionScreenshot(request *ComputerUseRequest) (*Empty, error) {
	err := m.client.Call("Plugin.TakeRegionScreenshot", request, new(Empty))
	return new(Empty), err
}

func (m *ComputerUseRPCClient) TakeCompressedScreenshot(request *ComputerUseRequest) (*Empty, error) {
	err := m.client.Call("Plugin.TakeCompressedScreenshot", request, new(Empty))
	return new(Empty), err
}

func (m *ComputerUseRPCClient) TakeCompressedRegionScreenshot(request *ComputerUseRequest) (*Empty, error) {
	err := m.client.Call("Plugin.TakeCompressedRegionScreenshot", request, new(Empty))
	return new(Empty), err
}

// Mouse control methods
func (m *ComputerUseRPCClient) GetMousePosition(request *ComputerUseRequest) (*Empty, error) {
	err := m.client.Call("Plugin.GetMousePosition", request, new(Empty))
	return new(Empty), err
}

func (m *ComputerUseRPCClient) MoveMouse(request *ComputerUseRequest) (*Empty, error) {
	err := m.client.Call("Plugin.MoveMouse", request, new(Empty))
	return new(Empty), err
}

func (m *ComputerUseRPCClient) Click(request *ComputerUseRequest) (*Empty, error) {
	err := m.client.Call("Plugin.Click", request, new(Empty))
	return new(Empty), err
}

func (m *ComputerUseRPCClient) Drag(request *ComputerUseRequest) (*Empty, error) {
	err := m.client.Call("Plugin.Drag", request, new(Empty))
	return new(Empty), err
}

func (m *ComputerUseRPCClient) Scroll(request *ComputerUseRequest) (*Empty, error) {
	err := m.client.Call("Plugin.Scroll", request, new(Empty))
	return new(Empty), err
}

// Keyboard control methods
func (m *ComputerUseRPCClient) TypeText(request *ComputerUseRequest) (*Empty, error) {
	err := m.client.Call("Plugin.TypeText", request, new(Empty))
	return new(Empty), err
}

func (m *ComputerUseRPCClient) PressKey(request *ComputerUseRequest) (*Empty, error) {
	err := m.client.Call("Plugin.PressKey", request, new(Empty))
	return new(Empty), err
}

func (m *ComputerUseRPCClient) PressHotkey(request *ComputerUseRequest) (*Empty, error) {
	err := m.client.Call("Plugin.PressHotkey", request, new(Empty))
	return new(Empty), err
}

// Display info methods
func (m *ComputerUseRPCClient) GetDisplayInfo(request *ComputerUseRequest) (*Empty, error) {
	err := m.client.Call("Plugin.GetDisplayInfo", request, new(Empty))
	return new(Empty), err
}

func (m *ComputerUseRPCClient) GetWindows(request *ComputerUseRequest) (*Empty, error) {
	err := m.client.Call("Plugin.GetWindows", request, new(Empty))
	return new(Empty), err
}

// Status method
func (m *ComputerUseRPCClient) GetStatus(request *ComputerUseRequest) (*Empty, error) {
	err := m.client.Call("Plugin.GetStatus", request, new(Empty))
	return new(Empty), err
}
