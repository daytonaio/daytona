package util

import (
	"os/exec"

	"github.com/daytonaio/daytona/common/os"
)

func GetOperatingSystem() (*os.OperatingSystem, error) {
	cmd := exec.Command("uname", "-a")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return os.OSFromUnameA(string(output))
}
