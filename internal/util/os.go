package util

import (
	"os/exec"
	"runtime"

	"github.com/daytonaio/daytona/common/os"
)

func GetOperatingSystem() (*os.OperatingSystem, error) {
	if runtime.GOOS == "windows" {
		return GetOperatingSystemWindows()
	}

	cmd := exec.Command("uname", "-a")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return os.OSFromUnameA(string(output))
}

func GetOperatingSystemWindows() (*os.OperatingSystem, error) {
	cmd := exec.Command("systeminfo")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return os.OSFromEchoProcessor(string(output))
}
