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

// ComputerUseStartResponse represents the response from starting computer use
type ComputerUseStartResponse struct {
	Message string `json:"message"`
} // @name ComputerUseStartResponse

// ComputerUseStopResponse represents the response from stopping computer use
type ComputerUseStopResponse struct {
	Message string `json:"message"`
} // @name ComputerUseStopResponse

// ComputerUse implements IComputerUse for Windows
type ComputerUse struct{}

// Ensure ComputerUse implements IComputerUse
var _ IComputerUse = &ComputerUse{}

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

// checkVNCRunning checks if VNC server is listening on the configured port
func checkVNCRunning() bool {
	address := fmt.Sprintf("127.0.0.1:%d", VNCPort)
	conn, err := net.DialTimeout("tcp", address, 2*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// GetStatus returns the current status of the computer use system
func (c *ComputerUse) GetStatus() (*ComputerUseStatusResponse, error) {
	// Check if either noVNC or VNC is running
	if checkNoVNCRunning() || checkVNCRunning() {
		return &ComputerUseStatusResponse{
			Status: "active",
		}, nil
	}

	return &ComputerUseStatusResponse{
		Status: "inactive",
	}, nil
}

// GetStatusHandler handles HTTP requests for status
//
//	@Summary		Get computer use status
//	@Description	Get the current status of the computer use system (VNC on Windows)
//	@Tags			computer-use
//	@Produce		json
//	@Success		200	{object}	ComputerUseStatusResponse
//	@Router			/computeruse/status [get]
//
//	@id				GetComputerUseSystemStatus
func GetStatusHandler(c *gin.Context) {
	cu := &ComputerUse{}
	status, err := cu.GetStatus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, status)
}

// StartHandler handles HTTP requests to start computer use
//
//	@Summary		Start computer use
//	@Description	Start the computer use system (VNC on Windows - assumes pre-installed)
//	@Tags			computer-use
//	@Produce		json
//	@Success		200	{object}	ComputerUseStartResponse
//	@Router			/computeruse/start [post]
//
//	@id				StartComputerUse
func StartHandler(c *gin.Context) {
	// On Windows, VNC is assumed to be pre-installed and running
	// This endpoint just confirms the status
	c.JSON(http.StatusOK, ComputerUseStartResponse{
		Message: "Computer use is available. VNC server is assumed to be pre-installed.",
	})
}

// StopHandler handles HTTP requests to stop computer use
//
//	@Summary		Stop computer use
//	@Description	Stop the computer use system (VNC on Windows - no-op as VNC is managed externally)
//	@Tags			computer-use
//	@Produce		json
//	@Success		200	{object}	ComputerUseStopResponse
//	@Router			/computeruse/stop [post]
//
//	@id				StopComputerUse
func StopHandler(c *gin.Context) {
	// On Windows, VNC is managed externally, so this is a no-op
	c.JSON(http.StatusOK, ComputerUseStopResponse{
		Message: "Computer use stop request acknowledged. VNC server is managed externally.",
	})
}

// ComputerUseDisabledMiddleware returns a middleware that handles requests when computer-use is disabled
func ComputerUseDisabledMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"message":  "Computer-use functionality is not available",
			"details":  "This computer-use operation is not yet implemented for Windows sandboxes.",
			"solution": "Use a VNC client to connect to port 6080 for remote desktop access.",
		})
		c.Abort()
	}
}
