// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package remotelogs

import (
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/gorilla/websocket"
)

type RemoteLogger struct {
	localLogger logs.Logger
	conn        *websocket.Conn
	baseUrl     string
}

func (r *RemoteLogger) Write(p []byte) (n int, err error) {
	if r.conn != nil {
		b, err := r.localLogger.ConstructJsonLogEntry(p)
		if err != nil {
			return len(p), err
		}

		err = r.conn.WriteMessage(websocket.TextMessage, b)
		if err != nil {
			return len(p), err
		}
	}

	return r.localLogger.Write(p)
}

func (r *RemoteLogger) Cleanup() error {
	if r.conn != nil {
		err := r.conn.Close()
		r.conn = nil
		if err != nil {
			return err
		}
	}

	return r.localLogger.Cleanup()
}

func (r *RemoteLogger) Close() error {
	if r.conn != nil {
		err := r.conn.Close()
		r.conn = nil
		if err != nil {
			return err
		}
	}

	return r.localLogger.Close()
}

func (r *RemoteLogger) ConstructJsonLogEntry(p []byte) ([]byte, error) {
	return r.localLogger.ConstructJsonLogEntry(p)
}
