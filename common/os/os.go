package os

import (
	"fmt"
	"strings"
)

type OperatingSystem string

const (
	Linux_64_86  OperatingSystem = "linux-amd64"
	Linux_arm64  OperatingSystem = "linux-arm64"
	Darwin_64_86 OperatingSystem = "darwin-amd64"
	Darwin_arm64 OperatingSystem = "darwin-arm64"
)

func OSFromUnameA(unameA string) (*OperatingSystem, error) {
	fields := strings.Fields(unameA)
	if len(fields) < 13 {
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
