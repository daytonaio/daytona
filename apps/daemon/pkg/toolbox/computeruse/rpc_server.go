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
func (m *ComputerUseRPCServer) TakeScreenshot(arg *ScreenshotRequest, resp *ScreenshotResponse) error {
	response, err := m.Impl.TakeScreenshot(arg)
	if err != nil {
		return err
	}
	*resp = *response
	return nil
}

func (m *ComputerUseRPCServer) TakeRegionScreenshot(arg *RegionScreenshotRequest, resp *ScreenshotResponse) error {
	response, err := m.Impl.TakeRegionScreenshot(arg)
	if err != nil {
		return err
	}
	*resp = *response
	return nil
}

func (m *ComputerUseRPCServer) TakeCompressedScreenshot(arg *CompressedScreenshotRequest, resp *ScreenshotResponse) error {
	response, err := m.Impl.TakeCompressedScreenshot(arg)
	if err != nil {
		return err
	}
	*resp = *response
	return nil
}

func (m *ComputerUseRPCServer) TakeCompressedRegionScreenshot(arg *CompressedRegionScreenshotRequest, resp *ScreenshotResponse) error {
	response, err := m.Impl.TakeCompressedRegionScreenshot(arg)
	if err != nil {
		return err
	}
	*resp = *response
	return nil
}

// Mouse control methods
func (m *ComputerUseRPCServer) GetMousePosition(arg any, resp *MousePositionResponse) error {
	response, err := m.Impl.GetMousePosition()
	if err != nil {
		return err
	}
	*resp = *response
	return nil
}

func (m *ComputerUseRPCServer) MoveMouse(arg *MoveMouseRequest, resp *MousePositionResponse) error {
	response, err := m.Impl.MoveMouse(arg)
	if err != nil {
		return err
	}
	*resp = *response
	return nil
}

func (m *ComputerUseRPCServer) Click(arg *ClickRequest, resp *MouseClickResponse) error {
	response, err := m.Impl.Click(arg)
	if err != nil {
		return err
	}
	*resp = *response
	return nil
}

func (m *ComputerUseRPCServer) Drag(arg *DragRequest, resp *MouseDragResponse) error {
	response, err := m.Impl.Drag(arg)
	if err != nil {
		return err
	}
	*resp = *response
	return nil
}

func (m *ComputerUseRPCServer) Scroll(arg *ScrollRequest, resp *ScrollResponse) error {
	response, err := m.Impl.Scroll(arg)
	if err != nil {
		return err
	}
	*resp = *response
	return nil
}

// Keyboard control methods
func (m *ComputerUseRPCServer) TypeText(arg *TypeTextRequest, resp *Empty) error {
	_, err := m.Impl.TypeText(arg)
	return err
}

func (m *ComputerUseRPCServer) PressKey(arg *PressKeyRequest, resp *Empty) error {
	_, err := m.Impl.PressKey(arg)
	return err
}

func (m *ComputerUseRPCServer) PressHotkey(arg *PressHotkeyRequest, resp *Empty) error {
	_, err := m.Impl.PressHotkey(arg)
	return err
}

// Display info methods
func (m *ComputerUseRPCServer) GetDisplayInfo(arg any, resp *DisplayInfoResponse) error {
	response, err := m.Impl.GetDisplayInfo()
	if err != nil {
		return err
	}
	*resp = *response
	return nil
}

func (m *ComputerUseRPCServer) GetWindows(arg any, resp *WindowsResponse) error {
	response, err := m.Impl.GetWindows()
	if err != nil {
		return err
	}
	*resp = *response
	return nil
}

// Status method
func (m *ComputerUseRPCServer) GetStatus(arg any, resp *StatusResponse) error {
	response, err := m.Impl.GetStatus()
	if err != nil {
		return err
	}
	*resp = *response
	return nil
}
