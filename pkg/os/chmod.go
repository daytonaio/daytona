package os

import "os/exec"

func ChmodX(filePath string) error {
	err := exec.Command("chmod", "+x", filePath).Run()
	if err != nil {
		return err
	}

	return nil
}
