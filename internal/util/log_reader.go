package util

import (
	"bufio"
	"context"
	"os/exec"
)

func ReadLog(ctx context.Context, filePath *string, follow bool, c chan []byte, errChan chan error) {
	if filePath == nil {
		return
	}

	ctxCancel, cancel := context.WithCancel(ctx)
	defer cancel()

	tailCmd := exec.CommandContext(ctxCancel, "tail", "-n", "+1")
	if follow {
		tailCmd.Args = append(tailCmd.Args, "-f")
	}
	tailCmd.Args = append(tailCmd.Args, *filePath)

	reader, err := tailCmd.StdoutPipe()
	if err != nil {
		errChan <- err
		return
	}
	scanner := bufio.NewScanner(reader)
	go func() {
		for scanner.Scan() {
			c <- scanner.Bytes()
		}
	}()

	err = tailCmd.Start()
	if err != nil {
		errChan <- err
		return
	}

	errChan <- tailCmd.Wait()
}
