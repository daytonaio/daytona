package ide

import (
	"os"
	"os/exec"

	"github.com/daytonaio/daytona/cmd/daytona/config"
)

func OpenTerminalSsh(activeProfile config.Profile, workspaceId string, projectName string) error {
	err := config.EnsureSshConfigEntryAdded(activeProfile.Id, workspaceId, projectName)
	if err != nil {
		return err
	}

	projectHostname := config.GetProjectHostname(activeProfile.Id, workspaceId, projectName)

	sshCommand := exec.Command("ssh", projectHostname)
	sshCommand.Stdin = os.Stdin
	sshCommand.Stdout = os.Stdout
	sshCommand.Stderr = os.Stderr

	return sshCommand.Run()
}
