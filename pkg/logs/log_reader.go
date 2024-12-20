// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package logs

import (
	"archive/zip"
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	log "github.com/sirupsen/logrus"
)

func ReadLog(ctx context.Context, logReader io.Reader, follow bool, c chan []byte, errChan chan error) {
	reader := bufio.NewReader(logReader)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			bytes := make([]byte, 1024)
			_, err := reader.Read(bytes)
			if err != nil {
				if err != io.EOF {
					errChan <- err
				} else if !follow {
					errChan <- io.EOF
					return
				}
				continue
			}
			c <- bytes
		}
	}
}

func ReadJSONLog(ctx context.Context, logReader io.Reader, follow bool, c chan interface{}, errChan chan error) {
	reader := bufio.NewReader(logReader)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			line, readErr := reader.ReadString('\n')
			if line != "" {
				stripped := strings.TrimSuffix(line, LogDelimiter)
				var logEntry LogEntry

				err := json.Unmarshal([]byte(stripped), &logEntry)
				if err != nil {
					log.Trace("Failed to parse log entry: ", err, string(stripped))
				}

				c <- logEntry
			}
			if readErr != nil {
				if readErr != io.EOF {
					c <- LogEntry{}
					errChan <- readErr
				} else if !follow {
					c <- LogEntry{}
					errChan <- io.EOF
					return
				}
			}
		}
	}
}

func ReadCompressedFile(filePath string) (io.Reader, error) {
	zipFile, err := zip.OpenReader(filePath)
	if err != nil {
		return nil, err
	}

	if len(zipFile.File) == 0 {
		return nil, fmt.Errorf("empty zip file")
	}

	return zipFile.File[0].Open()
}
