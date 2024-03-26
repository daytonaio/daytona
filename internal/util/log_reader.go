package util

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"os"
	"time"
)

func ReadLog(ctx context.Context, filePath *string, follow bool, c chan []byte, errChan chan error) {
	if filePath == nil {
		errChan <- os.ErrInvalid
		return
	}

	file, err := os.Open(*filePath)
	if err != nil {
		errChan <- err
		return
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			line, err := reader.ReadBytes('\n')
			if err != nil {
				if err != io.EOF {
					errChan <- err
				} else if !follow {
					errChan <- io.EOF
					return
				}
				time.Sleep(500 * time.Millisecond) // Sleep to avoid busy loop
				continue
			}
			// Trim the newline character
			line = bytes.TrimRight(line, "\n")
			c <- line
		}
	}
}
