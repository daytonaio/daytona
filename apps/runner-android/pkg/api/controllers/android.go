// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package controllers

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/daytonaio/runner-android/pkg/runner"
	"github.com/daytonaio/runner-android/pkg/sshgateway"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// ADBInfoResponse contains ADB connection information
type ADBInfoResponse struct {
	SandboxId        string `json:"sandboxId"`
	ADBPort          int    `json:"adbPort"`
	ADBSerial        string `json:"adbSerial"`
	InstanceNum      int    `json:"instanceNum"`
	SSHGatewayPort   int    `json:"sshGatewayPort"`
	SSHTunnelCommand string `json:"sshTunnelCommand"`
}

// GetADBInfo returns ADB connection information for a sandbox
//
//	@Tags			android
//	@Summary		Get ADB connection info
//	@Description	Returns ADB port and connection details for the sandbox
//	@Param			sandboxId	path		string	true	"Sandbox ID"
//	@Success		200			{object}	ADBInfoResponse
//	@Failure		404			{object}	string	"Sandbox not found"
//	@Failure		500			{object}	string	"Internal server error"
//	@Router			/sandboxes/{sandboxId}/adb/info [get]
func GetADBInfo(ctx *gin.Context) {
	r := runner.GetInstance()
	if r == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Runner not initialized"})
		return
	}

	sandboxId := ctx.Param("sandboxId")

	instance, exists := r.CVDClient.GetInstance(sandboxId)
	if !exists {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Sandbox not found"})
		return
	}

	gatewayPort := sshgateway.GetSSHGatewayPort()
	tunnelCmd := fmt.Sprintf("ssh -L 5555:localhost:%d -p %d %s@<gateway-host>",
		instance.ADBPort, gatewayPort, sandboxId)

	ctx.JSON(http.StatusOK, ADBInfoResponse{
		SandboxId:        sandboxId,
		ADBPort:          instance.ADBPort,
		ADBSerial:        instance.ADBSerial,
		InstanceNum:      instance.InstanceNum,
		SSHGatewayPort:   gatewayPort,
		SSHTunnelCommand: tunnelCmd,
	})
}

// InstallAPKRequest represents an APK installation request
type InstallAPKRequest struct {
	APKContent string   `json:"apkContent,omitempty"` // base64 encoded APK
	APKPath    string   `json:"apkPath,omitempty"`    // path to APK on device
	Flags      []string `json:"flags,omitempty"`      // install flags like -r, -t
}

// InstallAPK installs an APK on the Android device
//
//	@Tags			android
//	@Summary		Install APK
//	@Description	Installs an APK on the Android device
//	@Param			sandboxId	path		string				true	"Sandbox ID"
//	@Param			request		body		InstallAPKRequest	true	"Install request"
//	@Success		200			{object}	map[string]string	"Installation result"
//	@Failure		400			{object}	string				"Bad request"
//	@Failure		404			{object}	string				"Sandbox not found"
//	@Failure		500			{object}	string				"Internal server error"
//	@Router			/sandboxes/{sandboxId}/android/install [post]
func InstallAPK(ctx *gin.Context) {
	r := runner.GetInstance()
	if r == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Runner not initialized"})
		return
	}

	sandboxId := ctx.Param("sandboxId")

	serial, err := r.CVDClient.GetADBSerial(sandboxId)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Sandbox not found: %v", err)})
		return
	}

	// Check content type
	contentType := ctx.GetHeader("Content-Type")

	if strings.HasPrefix(contentType, "multipart/form-data") {
		// Handle multipart upload
		file, _, err := ctx.Request.FormFile("apk")
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Failed to get APK file: %v", err)})
			return
		}
		defer file.Close()

		flags := ctx.PostFormArray("flags")
		err = r.CVDClient.ADB().InstallFromReader(ctx.Request.Context(), serial, file, flags...)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to install APK: %v", err)})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"status": "installed"})
	} else {
		// Handle JSON request
		var req InstallAPKRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
			return
		}

		if req.APKContent != "" {
			// Install from base64 content
			content, err := base64.StdEncoding.DecodeString(req.APKContent)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid base64 APK content: %v", err)})
				return
			}

			reader := strings.NewReader(string(content))
			err = r.CVDClient.ADB().InstallFromReader(ctx.Request.Context(), serial, reader, req.Flags...)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to install APK: %v", err)})
				return
			}
		} else if req.APKPath != "" {
			// Install from device path
			err := r.CVDClient.ADB().Install(ctx.Request.Context(), serial, req.APKPath, req.Flags...)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to install APK: %v", err)})
				return
			}
		} else {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Either apkContent or apkPath is required"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"status": "installed"})
	}
}

// UninstallRequest represents an app uninstallation request
type UninstallRequest struct {
	PackageName string `json:"packageName" binding:"required"`
}

// UninstallApp uninstalls an app from the Android device
//
//	@Tags			android
//	@Summary		Uninstall app
//	@Description	Uninstalls an app from the Android device
//	@Param			sandboxId	path		string				true	"Sandbox ID"
//	@Param			request		body		UninstallRequest	true	"Uninstall request"
//	@Success		200			{object}	map[string]string	"Uninstallation result"
//	@Failure		400			{object}	string				"Bad request"
//	@Failure		404			{object}	string				"Sandbox not found"
//	@Failure		500			{object}	string				"Internal server error"
//	@Router			/sandboxes/{sandboxId}/android/uninstall [post]
func UninstallApp(ctx *gin.Context) {
	r := runner.GetInstance()
	if r == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Runner not initialized"})
		return
	}

	sandboxId := ctx.Param("sandboxId")

	var req UninstallRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	serial, err := r.CVDClient.GetADBSerial(sandboxId)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Sandbox not found: %v", err)})
		return
	}

	err = r.CVDClient.ADB().Uninstall(ctx.Request.Context(), serial, req.PackageName)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to uninstall: %v", err)})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "uninstalled", "package": req.PackageName})
}

// ListPackages lists installed packages on the Android device
//
//	@Tags			android
//	@Summary		List packages
//	@Description	Lists installed packages on the Android device
//	@Param			sandboxId	path		string		true	"Sandbox ID"
//	@Param			filter		query		string		false	"Filter: -3 (third-party), -s (system), -d (disabled), -e (enabled)"
//	@Success		200			{object}	[]string	"List of package names"
//	@Failure		404			{object}	string		"Sandbox not found"
//	@Failure		500			{object}	string		"Internal server error"
//	@Router			/sandboxes/{sandboxId}/android/packages [get]
func ListPackages(ctx *gin.Context) {
	r := runner.GetInstance()
	if r == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Runner not initialized"})
		return
	}

	sandboxId := ctx.Param("sandboxId")

	serial, err := r.CVDClient.GetADBSerial(sandboxId)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Sandbox not found: %v", err)})
		return
	}

	var flags []string
	if filter := ctx.Query("filter"); filter != "" {
		flags = strings.Split(filter, ",")
	}

	packages, err := r.CVDClient.ADB().ListPackages(ctx.Request.Context(), serial, flags...)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to list packages: %v", err)})
		return
	}

	ctx.JSON(http.StatusOK, packages)
}

// LaunchRequest represents an app launch request
type LaunchRequest struct {
	PackageName string   `json:"packageName" binding:"required"`
	Activity    string   `json:"activity,omitempty"` // optional, will use main activity if not specified
	Extras      []string `json:"extras,omitempty"`   // intent extras
}

// LaunchApp launches an app on the Android device
//
//	@Tags			android
//	@Summary		Launch app
//	@Description	Launches an app on the Android device
//	@Param			sandboxId	path		string			true	"Sandbox ID"
//	@Param			request		body		LaunchRequest	true	"Launch request"
//	@Success		200			{object}	map[string]string
//	@Failure		400			{object}	string	"Bad request"
//	@Failure		404			{object}	string	"Sandbox not found"
//	@Failure		500			{object}	string	"Internal server error"
//	@Router			/sandboxes/{sandboxId}/android/launch [post]
func LaunchApp(ctx *gin.Context) {
	r := runner.GetInstance()
	if r == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Runner not initialized"})
		return
	}

	sandboxId := ctx.Param("sandboxId")

	var req LaunchRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	serial, err := r.CVDClient.GetADBSerial(sandboxId)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Sandbox not found: %v", err)})
		return
	}

	// Build the component string
	component := req.PackageName
	if req.Activity != "" {
		component = fmt.Sprintf("%s/%s", req.PackageName, req.Activity)
	} else {
		// Use monkey to launch the main activity if none specified
		cmd := fmt.Sprintf("monkey -p %s -c android.intent.category.LAUNCHER 1", req.PackageName)
		output, exitCode, err := r.CVDClient.ADB().ShellWithExitCode(ctx.Request.Context(), serial, cmd)
		if err != nil || exitCode != 0 {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":  fmt.Sprintf("Failed to launch app: %v", err),
				"output": output,
			})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"status": "launched", "package": req.PackageName})
		return
	}

	err = r.CVDClient.ADB().StartActivity(ctx.Request.Context(), serial, component, req.Extras...)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to launch activity: %v", err)})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "launched", "component": component})
}

// ForceStopRequest represents a force stop request
type ForceStopRequest struct {
	PackageName string `json:"packageName" binding:"required"`
}

// ForceStopApp force stops an app on the Android device
//
//	@Tags			android
//	@Summary		Force stop app
//	@Description	Force stops an app on the Android device
//	@Param			sandboxId	path		string				true	"Sandbox ID"
//	@Param			request		body		ForceStopRequest	true	"Force stop request"
//	@Success		200			{object}	map[string]string
//	@Failure		400			{object}	string	"Bad request"
//	@Failure		404			{object}	string	"Sandbox not found"
//	@Failure		500			{object}	string	"Internal server error"
//	@Router			/sandboxes/{sandboxId}/android/stop [post]
func ForceStopApp(ctx *gin.Context) {
	r := runner.GetInstance()
	if r == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Runner not initialized"})
		return
	}

	sandboxId := ctx.Param("sandboxId")

	var req ForceStopRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	serial, err := r.CVDClient.GetADBSerial(sandboxId)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Sandbox not found: %v", err)})
		return
	}

	err = r.CVDClient.ADB().ForceStop(ctx.Request.Context(), serial, req.PackageName)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to force stop: %v", err)})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "stopped", "package": req.PackageName})
}

// GetSystemProps gets system properties from the Android device
//
//	@Tags			android
//	@Summary		Get system properties
//	@Description	Gets system properties from the Android device
//	@Param			sandboxId	path		string				true	"Sandbox ID"
//	@Param			prop		query		string				false	"Specific property to get (returns all if not specified)"
//	@Success		200			{object}	map[string]string	"System properties"
//	@Failure		404			{object}	string				"Sandbox not found"
//	@Failure		500			{object}	string				"Internal server error"
//	@Router			/sandboxes/{sandboxId}/android/props [get]
func GetSystemProps(ctx *gin.Context) {
	r := runner.GetInstance()
	if r == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Runner not initialized"})
		return
	}

	sandboxId := ctx.Param("sandboxId")

	serial, err := r.CVDClient.GetADBSerial(sandboxId)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Sandbox not found: %v", err)})
		return
	}

	prop := ctx.Query("prop")
	if prop != "" {
		// Get specific property
		value, err := r.CVDClient.ADB().GetProp(ctx.Request.Context(), serial, prop)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get property: %v", err)})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{prop: value})
	} else {
		// Get all properties
		output, _, err := r.CVDClient.ADB().Shell(ctx.Request.Context(), serial, "getprop")
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get properties: %v", err)})
			return
		}

		// Parse properties
		props := make(map[string]string)
		for _, line := range strings.Split(output, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			// Parse [key]: [value] format
			if strings.HasPrefix(line, "[") {
				parts := strings.SplitN(line, "]: [", 2)
				if len(parts) == 2 {
					key := strings.TrimPrefix(parts[0], "[")
					value := strings.TrimSuffix(parts[1], "]")
					props[key] = value
				}
			}
		}

		ctx.JSON(http.StatusOK, props)
	}
}

// StreamLogcat streams logcat output as Server-Sent Events
//
//	@Tags			android
//	@Summary		Stream logcat
//	@Description	Streams logcat output as Server-Sent Events
//	@Param			sandboxId	path	string	true	"Sandbox ID"
//	@Param			tag			query	string	false	"Filter by tag"
//	@Param			level		query	string	false	"Minimum log level (V, D, I, W, E, F)"
//	@Success		200			{string}	string	"SSE stream of logcat output"
//	@Failure		404			{object}	string	"Sandbox not found"
//	@Failure		500			{object}	string	"Internal server error"
//	@Router			/sandboxes/{sandboxId}/android/logcat [get]
func StreamLogcatSSE(ctx *gin.Context) {
	r := runner.GetInstance()
	if r == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Runner not initialized"})
		return
	}

	sandboxId := ctx.Param("sandboxId")

	serial, err := r.CVDClient.GetADBSerial(sandboxId)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Sandbox not found: %v", err)})
		return
	}

	// Build logcat arguments
	args := []string{"-v", "time"}
	if tag := ctx.Query("tag"); tag != "" {
		args = append(args, "-s", tag)
	}
	if level := ctx.Query("level"); level != "" {
		args = append(args, fmt.Sprintf("*:%s", strings.ToUpper(level)))
	}

	// Set SSE headers
	ctx.Header("Content-Type", "text/event-stream")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")

	// Create a context that's cancelled when the client disconnects
	streamCtx, cancel := context.WithCancel(ctx.Request.Context())
	defer cancel()

	// Get logcat stream
	logcatReader, err := r.CVDClient.ADB().Logcat(streamCtx, serial, args...)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to start logcat: %v", err)})
		return
	}
	defer logcatReader.Close()

	// Stream logcat output as SSE events
	buf := make([]byte, 4096)
	for {
		select {
		case <-streamCtx.Done():
			return
		default:
			n, err := logcatReader.Read(buf)
			if err != nil {
				if err != io.EOF {
					log.Errorf("Logcat read error: %v", err)
				}
				return
			}
			if n > 0 {
				// Send as SSE data event
				ctx.SSEvent("message", string(buf[:n]))
				ctx.Writer.Flush()
			}
		}
	}
}

// GetDeviceInfo gets information about the Android device
//
//	@Tags			android
//	@Summary		Get device info
//	@Description	Gets information about the Android device (model, version, etc.)
//	@Param			sandboxId	path		string				true	"Sandbox ID"
//	@Success		200			{object}	map[string]string	"Device information"
//	@Failure		404			{object}	string				"Sandbox not found"
//	@Failure		500			{object}	string				"Internal server error"
//	@Router			/sandboxes/{sandboxId}/android/device [get]
func GetDeviceInfo(ctx *gin.Context) {
	r := runner.GetInstance()
	if r == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Runner not initialized"})
		return
	}

	sandboxId := ctx.Param("sandboxId")

	serial, err := r.CVDClient.GetADBSerial(sandboxId)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Sandbox not found: %v", err)})
		return
	}

	adb := r.CVDClient.ADB()
	reqCtx := ctx.Request.Context()

	// Get various device properties
	info := make(map[string]string)

	props := []struct {
		name string
		prop string
	}{
		{"model", "ro.product.model"},
		{"manufacturer", "ro.product.manufacturer"},
		{"brand", "ro.product.brand"},
		{"device", "ro.product.device"},
		{"androidVersion", "ro.build.version.release"},
		{"sdkVersion", "ro.build.version.sdk"},
		{"buildId", "ro.build.id"},
		{"buildType", "ro.build.type"},
		{"serial", "ro.serialno"},
		{"board", "ro.product.board"},
		{"cpu", "ro.product.cpu.abi"},
	}

	for _, p := range props {
		if value, err := adb.GetProp(reqCtx, serial, p.prop); err == nil && value != "" {
			info[p.name] = value
		}
	}

	// Add ADB connection info
	info["adbSerial"] = serial

	ctx.JSON(http.StatusOK, info)
}
