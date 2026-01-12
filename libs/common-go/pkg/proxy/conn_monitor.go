// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package proxy

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

type ConnectionMonitor struct {
	gin.ResponseWriter
	OnConnClosed func()
}

func (cm *ConnectionMonitor) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := cm.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("ResponseWriter doesn't support Hijacker interface")
	}

	conn, rw, err := hijacker.Hijack()
	if err != nil {
		return conn, rw, err
	}

	// Wrap the connection to detect when it's closed
	wrappedConn := &monitoredConn{
		Conn:         conn,
		onConnClosed: cm.OnConnClosed,
	}

	return wrappedConn, rw, nil
}

type monitoredConn struct {
	net.Conn
	onConnClosed func()
	closeOnce    sync.Once
}

func (mc *monitoredConn) Close() error {
	mc.closeOnce.Do(func() {
		if mc.onConnClosed != nil {
			mc.onConnClosed()
		}
	})
	return mc.Conn.Close()
}

func (mc *monitoredConn) Read(b []byte) (n int, err error) {
	n, err = mc.Conn.Read(b)
	if err != nil {
		mc.closeOnce.Do(func() {
			if mc.onConnClosed != nil {
				mc.onConnClosed()
			}
		})
	}
	return n, err
}

func (mc *monitoredConn) Write(b []byte) (n int, err error) {
	n, err = mc.Conn.Write(b)
	if err != nil {
		mc.closeOnce.Do(func() {
			if mc.onConnClosed != nil {
				mc.onConnClosed()
			}
		})
	}
	return n, err
}
