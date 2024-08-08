// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"

	"github.com/daytonaio/daytona/pkg/logs"
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

func ReadJSONLog(ctx context.Context, logReader io.Reader, follow bool, retry bool, c chan interface{}, errChan chan error) {
	reader := bufio.NewReader(logReader)
	var buffer bytes.Buffer
	delimiter := []byte(logs.LogDelimiter)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			byteChunk := make([]byte, 1024)

			if retry {
				for {
					n, err := reader.Read(byteChunk)
					if err == nil {
						buffer.Write(byteChunk[:n])
						break
					} else if !follow && err == io.EOF {
						errChan <- io.EOF
						return
					}
				}
			} else {
				n, err := reader.Read(byteChunk)
				if err != nil {
					if err != io.EOF {
						errChan <- err
					} else if !follow && err == io.EOF {
						errChan <- io.EOF
						return
					}
					continue
				}
				buffer.Write(byteChunk[:n])
			}

			data := buffer.Bytes()

			for {
				index := bytes.Index(data, delimiter)
				if index == -1 {
					break
				}

				jsonData := data[:index]
				var logEntry logs.LogEntry
				err := json.Unmarshal(jsonData, &logEntry)
				if err != nil {
					if err != io.EOF {
						errChan <- err
					} else if !follow {
						errChan <- io.EOF
						return
					}
				}
				c <- logEntry

				data = data[index+len(delimiter):]
			}

			buffer.Reset()
			buffer.Write(data)
		}
	}
}
