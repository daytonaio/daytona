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

func ReadJSONLog(ctx context.Context, logReader io.Reader, follow bool, c chan interface{}, errChan chan error) {
	var buffer bytes.Buffer
	reader := bufio.NewReader(logReader)
	delimiter := []byte(logs.LogDelimiter)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			byteChunk := make([]byte, 1024)

			n, err := reader.Read(byteChunk)
			if err != nil {
				if err != io.EOF {
					errChan <- err
				} else if !follow {
					errChan <- io.EOF
					return
				}
			}
			buffer.Write(byteChunk[:n])

			data := buffer.Bytes()

			index := bytes.Index(data, delimiter)

			if index != -1 { // if the delimiter is found, process the log entry

				var logEntry logs.LogEntry

				err = json.Unmarshal(data[:index], &logEntry)
				if err != nil {
					return
				}

				c <- logEntry
				buffer.Reset()
				buffer.Write(data[index+len(delimiter):]) // write remaining data to buffer
			}
		}
	}
}
