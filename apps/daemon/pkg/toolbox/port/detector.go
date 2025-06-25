// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package port

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/cakturk/go-netstat/netstat"
	"github.com/gin-gonic/gin"
	cmap "github.com/orcaman/concurrent-map/v2"
)

type portsDetector struct {
	portMap cmap.ConcurrentMap[string, bool]
}

func NewPortsDetector() *portsDetector {
	return &portsDetector{
		portMap: cmap.New[bool](),
	}
}

func (d *portsDetector) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			time.Sleep(1 * time.Second)
			// get only listening TCP sockets
			tabs, err := netstat.TCPSocks(func(s *netstat.SockTabEntry) bool {
				return s.State == netstat.Listen
			})
			if err != nil {
				continue
			}

			freshMap := map[string]bool{}
			for _, e := range tabs {
				s := strconv.Itoa(int(e.LocalAddr.Port))
				freshMap[s] = true
				d.portMap.Set(s, true)
			}

			for _, port := range d.portMap.Keys() {
				if !freshMap[port] {
					d.portMap.Remove(port)
				}
			}
		}
	}
}

func (d *portsDetector) GetPorts(c *gin.Context) {
	ports := PortList{
		Ports: []uint{},
	}

	for _, port := range d.portMap.Keys() {
		portInt, err := strconv.Atoi(port)
		if err != nil {
			continue
		}
		ports.Ports = append(ports.Ports, uint(portInt))
	}

	c.JSON(http.StatusOK, ports)
}

func (d *portsDetector) IsPortInUse(c *gin.Context) {
	portParam := c.Param("port")

	port, err := strconv.Atoi(portParam)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, errors.New("invalid port: must be a number between 3000 and 9999"))
		return
	}

	if port < 3000 || port > 9999 {
		c.AbortWithError(http.StatusBadRequest, errors.New("port out of range: must be between 3000 and 9999"))
		return
	}

	portStr := strconv.Itoa(port)

	if d.portMap.Has(portStr) {
		c.JSON(http.StatusOK, IsPortInUseResponse{
			IsInUse: true,
		})
	} else {
		// If the port is not in the map, we check synchronously if it's in use and update the map
		_, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", port), 50*time.Millisecond)
		if err != nil {
			c.JSON(http.StatusOK, IsPortInUseResponse{
				IsInUse: false,
			})
		} else {
			d.portMap.Set(portStr, true)
			c.JSON(http.StatusOK, IsPortInUseResponse{
				IsInUse: true,
			})
		}
	}

}
