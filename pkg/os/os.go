// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package os

import (
	"bytes"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
)

type OperatingSystem string

const (
	Linux_64_86   OperatingSystem = "linux-amd64"
	Linux_arm64   OperatingSystem = "linux-arm64"
	Darwin_64_86  OperatingSystem = "darwin-amd64"
	Darwin_arm64  OperatingSystem = "darwin-arm64"
	Windows_64_86 OperatingSystem = "windows-amd64"
	Windows_arm64 OperatingSystem = "windows-arm64"
)

func OSFromUnameA(unameA string) (*OperatingSystem, error) {
	fields := strings.Fields(unameA)
	if len(fields) < 3 {
		return nil, fmt.Errorf("unexpected output format")
	}

	if strings.Contains(unameA, "Darwin") && strings.Contains(unameA, "arm64") {
		arch := Darwin_arm64
		return &arch, nil
	} else if strings.Contains(unameA, "Darwin") && strings.Contains(unameA, "x86_64") {
		arch := Darwin_64_86
		return &arch, nil
	} else if strings.Contains(unameA, "arm64") || strings.Contains(unameA, "aarch64") {
		arch := Linux_arm64
		return &arch, nil
	} else if strings.Contains(unameA, "x86_64") {
		arch := Linux_64_86
		return &arch, nil
	} else {
		return nil, fmt.Errorf("unsupported architecture in uname -a output")
	}
}

func OSFromEchoProcessor(output string) (*OperatingSystem, error) {
	if strings.Contains(output, "ARM64") {
		arch := Windows_arm64
		return &arch, nil
	} else if strings.Contains(output, "AMD64") || strings.Contains(output, "Intel") {
		arch := Windows_64_86
		return &arch, nil
	} else {
		return nil, fmt.Errorf("unsupported architecture in echo PROCESSOR_ARCHITECTURE output")
	}
}

func OSFromSysCallUtsName(uname *syscall.Utsname) (*OperatingSystem, error) {
	var arch OperatingSystem
	sysname := unameToString(uname.Sysname)
	machine := unameToString(uname.Machine)
	if sysname == "Darwin" && machine == "arm64" {
		arch = Darwin_arm64
		return &arch, nil
	} else if sysname == "Darwin" && machine == "x86_64" {
		arch = Darwin_64_86
		return &arch, nil
	} else if sysname == "arm64" || sysname == "aarch64" {
		arch = Linux_arm64
		return &arch, nil
	} else if machine == "x86_64" {
		arch = Linux_64_86
		return &arch, nil
	} else {
		return nil, fmt.Errorf("unsupported architecture in `syscall.Uname` output")
	}
}

func GetOperatingSystem() (*OperatingSystem, error) {
	if runtime.GOOS == "windows" {
		return GetOperatingSystemWindows()
	}

	var uname *syscall.Utsname
	if err := syscall.Uname(uname); err == nil {
		return OSFromSysCallUtsName(uname)
	}

	cmd := exec.Command("uname", "-a")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return OSFromUnameA(string(output))
}

func GetOperatingSystemWindows() (*OperatingSystem, error) {
	cmd := exec.Command("systeminfo")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return OSFromEchoProcessor(string(output))
}

func unameToString(input [65]int8) string {
	var buffer bytes.Buffer
	for _, value := range input {
		if value == 0 {
			break
		}
		buffer.WriteByte(uint8(value))
	}
	return buffer.String()
}
