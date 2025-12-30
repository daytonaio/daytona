// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	// noVNC websockify port (serves web-based VNC client)
	NoVNCPort = 6080
	// TightVNC RFB port (internal, used by websockify)
	VNCPort = 5900
)

// ComputerUseStatusResponse represents the status of computer use functionality
type ComputerUseStatusResponse struct {
	Status string `json:"status"`
} // @name ComputerUseStatusResponse

// checkNoVNCRunning checks if noVNC (websockify) is listening on the configured port
func checkNoVNCRunning() bool {
	address := fmt.Sprintf("127.0.0.1:%d", NoVNCPort)
	conn, err := net.DialTimeout("tcp", address, 2*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// ComputerUseStartResponse represents the response from starting computer use
type ComputerUseStartResponse struct {
	Message string `json:"message"`
} // @name ComputerUseStartResponse

// ComputerUseStopResponse represents the response from stopping computer use
type ComputerUseStopResponse struct {
	Message string `json:"message"`
} // @name ComputerUseStopResponse

// GetStatus godoc
//
//	@Summary		Get computer use status
//	@Description	Get the current status of the computer use system (VNC on Windows)
//	@Tags			computer-use
//	@Produce		json
//	@Success		200	{object}	ComputerUseStatusResponse
//	@Router			/computeruse/status [get]
//
//	@id				GetComputerUseSystemStatus
func GetStatus(c *gin.Context) {
	if checkNoVNCRunning() {
		c.JSON(http.StatusOK, ComputerUseStatusResponse{
			Status: "active",
		})
		return
	}

	c.JSON(http.StatusOK, ComputerUseStatusResponse{
		Status: "inactive",
	})
}

// Start godoc
//
//	@Summary		Start computer use
//	@Description	Start the computer use system (VNC on Windows)
//	@Tags			computer-use
//	@Produce		json
//	@Success		200	{object}	ComputerUseStartResponse
//	@Router			/computeruse/start [post]
//
//	@id				StartComputerUse
func Start(c *gin.Context) {
	c.JSON(http.StatusOK, ComputerUseStartResponse{
		Message: "Computer use started",
	})
}

// Stop godoc
//
//	@Summary		Stop computer use
//	@Description	Stop the computer use system (VNC on Windows)
//	@Tags			computer-use
//	@Produce		json
//	@Success		200	{object}	ComputerUseStopResponse
//	@Router			/computeruse/stop [post]
//
//	@id				StopComputerUse
func Stop(c *gin.Context) {
	c.JSON(http.StatusOK, ComputerUseStopResponse{
		Message: "Computer use stopped",
	})
}

// ComputerUseDisabledMiddleware returns a middleware that handles requests when computer-use is disabled
func ComputerUseDisabledMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"message":  "Computer-use functionality is not available",
			"details":  "This computer-use operation is not yet implemented for Windows sandboxes. VNC is available on port 6080.",
			"solution": "Use a VNC client to connect to port 6080 for remote desktop access.",
		})
		c.Abort()
	}
}
