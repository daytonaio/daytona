// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

type ComputerUseRPCServer struct {
	Impl IComputerUse
}

// Process management methods
func (m *ComputerUseRPCServer) Initialize(arg any, resp *Empty) error {
	_, err := m.Impl.Initialize()
	return err
}

func (m *ComputerUseRPCServer) Start(arg any, resp *Empty) error {
	_, err := m.Impl.Start()
	return err
}

func (m *ComputerUseRPCServer) Stop(arg any, resp *Empty) error {
	_, err := m.Impl.Stop()
	return err
}

func (m *ComputerUseRPCServer) GetProcessStatus(arg any, resp *map[string]ProcessStatus) error {
	status, err := m.Impl.GetProcessStatus()
	if err != nil {
		return err
	}
	*resp = status
	return nil
}

func (m *ComputerUseRPCServer) IsProcessRunning(arg *ProcessRequest, resp *bool) error {
	isRunning, err := m.Impl.IsProcessRunning(arg)
	if err != nil {
		return err
	}
	*resp = isRunning
	return nil
}

func (m *ComputerUseRPCServer) RestartProcess(arg *ProcessRequest, resp *Empty) error {
	_, err := m.Impl.RestartProcess(arg)
	return err
}

func (m *ComputerUseRPCServer) GetProcessLogs(arg *ProcessRequest, resp *string) error {
	logs, err := m.Impl.GetProcessLogs(arg)
	if err != nil {
		return err
	}
	*resp = logs
	return nil
}

func (m *ComputerUseRPCServer) GetProcessErrors(arg *ProcessRequest, resp *string) error {
	errors, err := m.Impl.GetProcessErrors(arg)
	if err != nil {
		return err
	}
	*resp = errors
	return nil
}

// Screenshot methods
func (m *ComputerUseRPCServer) TakeScreenshot(arg *ComputerUseRequest, resp *Empty) error {
	_, err := m.Impl.TakeScreenshot(arg)
	return err
}

func (m *ComputerUseRPCServer) TakeRegionScreenshot(arg *ComputerUseRequest, resp *Empty) error {
	_, err := m.Impl.TakeRegionScreenshot(arg)
	return err
}

func (m *ComputerUseRPCServer) TakeCompressedScreenshot(arg *ComputerUseRequest, resp *Empty) error {
	_, err := m.Impl.TakeCompressedScreenshot(arg)
	return err
}

func (m *ComputerUseRPCServer) TakeCompressedRegionScreenshot(arg *ComputerUseRequest, resp *Empty) error {
	_, err := m.Impl.TakeCompressedRegionScreenshot(arg)
	return err
}

// Mouse control methods
func (m *ComputerUseRPCServer) GetMousePosition(arg *ComputerUseRequest, resp *Empty) error {
	_, err := m.Impl.GetMousePosition(arg)
	return err
}

func (m *ComputerUseRPCServer) MoveMouse(arg *ComputerUseRequest, resp *Empty) error {
	_, err := m.Impl.MoveMouse(arg)
	return err
}

func (m *ComputerUseRPCServer) Click(arg *ComputerUseRequest, resp *Empty) error {
	_, err := m.Impl.Click(arg)
	return err
}

func (m *ComputerUseRPCServer) Drag(arg *ComputerUseRequest, resp *Empty) error {
	_, err := m.Impl.Drag(arg)
	return err
}

func (m *ComputerUseRPCServer) Scroll(arg *ComputerUseRequest, resp *Empty) error {
	_, err := m.Impl.Scroll(arg)
	return err
}

// Keyboard control methods
func (m *ComputerUseRPCServer) TypeText(arg *ComputerUseRequest, resp *Empty) error {
	_, err := m.Impl.TypeText(arg)
	return err
}

func (m *ComputerUseRPCServer) PressKey(arg *ComputerUseRequest, resp *Empty) error {
	_, err := m.Impl.PressKey(arg)
	return err
}

func (m *ComputerUseRPCServer) PressHotkey(arg *ComputerUseRequest, resp *Empty) error {
	_, err := m.Impl.PressHotkey(arg)
	return err
}

// Display info methods
func (m *ComputerUseRPCServer) GetDisplayInfo(arg *ComputerUseRequest, resp *Empty) error {
	_, err := m.Impl.GetDisplayInfo(arg)
	return err
}

func (m *ComputerUseRPCServer) GetWindows(arg *ComputerUseRequest, resp *Empty) error {
	_, err := m.Impl.GetWindows(arg)
	return err
}

// Status method
func (m *ComputerUseRPCServer) GetStatus(arg *ComputerUseRequest, resp *Empty) error {
	_, err := m.Impl.GetStatus(arg)
	return err
}
