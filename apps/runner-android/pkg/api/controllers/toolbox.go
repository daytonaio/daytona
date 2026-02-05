// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package controllers

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/daytonaio/runner-android/pkg/cuttlefish"
	"github.com/daytonaio/runner-android/pkg/runner"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// ExecuteRequest represents a command execution request
type ExecuteRequest struct {
	Command string `json:"command" binding:"required"`
	Cwd     string `json:"cwd,omitempty"`
	Timeout int    `json:"timeout,omitempty"` // timeout in seconds
}

// ExecuteResponse represents a command execution response
type ExecuteResponse struct {
	Result   string `json:"result"`
	ExitCode int    `json:"exitCode"`
}

// HandleToolboxRequest routes toolbox requests to the appropriate handler
// This is called from ProxyRequest when a toolbox endpoint is requested
func HandleToolboxRequest(ctx *gin.Context) {
	path := ctx.Param("path")
	method := ctx.Request.Method

	// Remove leading slash if present
	path = strings.TrimPrefix(path, "/")

	log.Debugf("Toolbox request: %s %s", method, path)

	// Route to appropriate handler
	switch {
	// Process execution
	case method == "POST" && (path == "process/execute" || path == "process"):
		handleProcessExecute(ctx)
	case method == "GET" && strings.HasPrefix(path, "process/commands/"):
		handleProcessCommandLogs(ctx, strings.TrimPrefix(path, "process/commands/"))

	// File operations
	case method == "GET" && (path == "files" || strings.HasPrefix(path, "files/")):
		if ctx.Query("download") == "true" || strings.HasPrefix(path, "files/download") {
			handleFilesDownload(ctx)
		} else if strings.HasPrefix(path, "files/info") {
			handleFilesInfo(ctx)
		} else {
			handleFilesList(ctx)
		}
	case method == "POST" && strings.HasPrefix(path, "files/upload"):
		handleFilesUpload(ctx)
	case method == "POST" && strings.HasPrefix(path, "files/folder"):
		handleFilesCreateFolder(ctx)
	case method == "POST" && strings.HasPrefix(path, "files/move"):
		handleFilesMove(ctx)
	case method == "DELETE" && strings.HasPrefix(path, "files"):
		handleFilesDelete(ctx)

	// Git operations (not supported for Android)
	case strings.HasPrefix(path, "git/"):
		ctx.JSON(http.StatusNotImplemented, gin.H{
			"error": "Git operations not supported on Android devices",
		})

	// Workspace operations
	case strings.HasPrefix(path, "workspace"):
		handleWorkspace(ctx)

	// Computer use (screenshots, input)
	case strings.HasPrefix(path, "computeruse/"):
		HandleComputerUse(ctx, path)

	default:
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("Unknown toolbox endpoint: %s %s", method, path),
		})
	}
}

// handleProcessExecute handles POST /process/execute
func handleProcessExecute(ctx *gin.Context) {
	r := runner.GetInstance()
	if r == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Runner not initialized"})
		return
	}

	sandboxId := ctx.Param("sandboxId")

	var req ExecuteRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	serial, err := r.CVDClient.GetADBSerial(sandboxId)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Sandbox not found: %v", err)})
		return
	}

	// Build command with optional cwd
	command := req.Command
	if req.Cwd != "" {
		command = fmt.Sprintf("cd %s && %s", req.Cwd, req.Command)
	}

	log.Infof("Executing command on sandbox %s: %s", sandboxId, command)

	// Execute via ADB
	output, exitCode, err := r.CVDClient.ADB().ShellWithExitCode(ctx.Request.Context(), serial, command)
	if err != nil {
		log.Warnf("Command execution failed: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":    fmt.Sprintf("Command execution failed: %v", err),
			"result":   output,
			"exitCode": exitCode,
		})
		return
	}

	ctx.JSON(http.StatusOK, ExecuteResponse{
		Result:   output,
		ExitCode: exitCode,
	})
}

// handleProcessCommandLogs handles GET /process/commands/{commandId} for logs
func handleProcessCommandLogs(ctx *gin.Context, commandId string) {
	// Android doesn't have persistent command sessions like Linux
	ctx.JSON(http.StatusNotImplemented, gin.H{
		"error":     "Command logs not available for Android. Commands execute synchronously.",
		"commandId": commandId,
	})
}

// handleFilesList handles GET /files to list files in a directory
func handleFilesList(ctx *gin.Context) {
	r := runner.GetInstance()
	if r == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Runner not initialized"})
		return
	}

	sandboxId := ctx.Param("sandboxId")
	path := ctx.Query("path")
	if path == "" {
		path = "/sdcard"
	}

	serial, err := r.CVDClient.GetADBSerial(sandboxId)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Sandbox not found: %v", err)})
		return
	}

	files, err := r.CVDClient.ADB().ListFiles(ctx.Request.Context(), serial, path)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to list files: %v", err)})
		return
	}

	ctx.JSON(http.StatusOK, files)
}

// handleFilesInfo handles GET /files/info to get file metadata
func handleFilesInfo(ctx *gin.Context) {
	r := runner.GetInstance()
	if r == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Runner not initialized"})
		return
	}

	sandboxId := ctx.Param("sandboxId")
	path := ctx.Query("path")
	if path == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "path query parameter is required"})
		return
	}

	serial, err := r.CVDClient.GetADBSerial(sandboxId)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Sandbox not found: %v", err)})
		return
	}

	info, err := r.CVDClient.ADB().Stat(ctx.Request.Context(), serial, path)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("File not found: %v", err)})
		return
	}

	ctx.JSON(http.StatusOK, info)
}

// handleFilesDownload handles GET /files/download to download a file
func handleFilesDownload(ctx *gin.Context) {
	r := runner.GetInstance()
	if r == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Runner not initialized"})
		return
	}

	sandboxId := ctx.Param("sandboxId")
	path := ctx.Query("path")
	if path == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "path query parameter is required"})
		return
	}

	serial, err := r.CVDClient.GetADBSerial(sandboxId)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Sandbox not found: %v", err)})
		return
	}

	// Get file info for content-length and filename
	info, err := r.CVDClient.ADB().Stat(ctx.Request.Context(), serial, path)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("File not found: %v", err)})
		return
	}

	// Set headers for download
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", info.Name))
	ctx.Header("Content-Type", "application/octet-stream")
	if info.Size > 0 {
		ctx.Header("Content-Length", strconv.FormatInt(info.Size, 10))
	}

	// Stream file content via ADB pull
	err = r.CVDClient.ADB().PullToWriter(ctx.Request.Context(), serial, path, ctx.Writer)
	if err != nil {
		log.Errorf("Failed to download file: %v", err)
		// Error already partially written, can't send JSON
		return
	}
}

// UploadRequest represents a file upload request
type UploadRequest struct {
	Path      string `json:"path" binding:"required"`
	Content   string `json:"content"`        // base64 encoded content
	Mode      string `json:"mode,omitempty"` // file mode (not supported on Android)
	Overwrite bool   `json:"overwrite,omitempty"`
}

// handleFilesUpload handles POST /files/upload to upload a file
func handleFilesUpload(ctx *gin.Context) {
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

	// Check content type - support both JSON and multipart
	contentType := ctx.GetHeader("Content-Type")

	if strings.HasPrefix(contentType, "multipart/form-data") {
		// Handle multipart upload
		file, header, err := ctx.Request.FormFile("file")
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Failed to get file: %v", err)})
			return
		}
		defer file.Close()

		remotePath := ctx.PostForm("path")
		if remotePath == "" {
			remotePath = "/sdcard/" + header.Filename
		}

		err = r.CVDClient.ADB().PushFromReader(ctx.Request.Context(), serial, file, remotePath)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to upload file: %v", err)})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"path": remotePath, "size": header.Size})
	} else {
		// Handle JSON upload
		var req UploadRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
			return
		}

		// Decode base64 content
		content, err := base64.StdEncoding.DecodeString(req.Content)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid base64 content: %v", err)})
			return
		}

		reader := strings.NewReader(string(content))
		err = r.CVDClient.ADB().PushFromReader(ctx.Request.Context(), serial, reader, req.Path)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to upload file: %v", err)})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"path": req.Path, "size": len(content)})
	}
}

// FolderRequest represents a folder creation request
type FolderRequest struct {
	Path string `json:"path" binding:"required"`
	Mode string `json:"mode,omitempty"`
}

// handleFilesCreateFolder handles POST /files/folder to create a directory
func handleFilesCreateFolder(ctx *gin.Context) {
	r := runner.GetInstance()
	if r == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Runner not initialized"})
		return
	}

	sandboxId := ctx.Param("sandboxId")

	var req FolderRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	serial, err := r.CVDClient.GetADBSerial(sandboxId)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Sandbox not found: %v", err)})
		return
	}

	err = r.CVDClient.ADB().Mkdir(ctx.Request.Context(), serial, req.Path, true)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create folder: %v", err)})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"path": req.Path})
}

// MoveRequest represents a file move/rename request
type MoveRequest struct {
	Source      string `json:"source" binding:"required"`
	Destination string `json:"destination" binding:"required"`
}

// handleFilesMove handles POST /files/move to move/rename files
func handleFilesMove(ctx *gin.Context) {
	r := runner.GetInstance()
	if r == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Runner not initialized"})
		return
	}

	sandboxId := ctx.Param("sandboxId")

	var req MoveRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	serial, err := r.CVDClient.GetADBSerial(sandboxId)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Sandbox not found: %v", err)})
		return
	}

	err = r.CVDClient.ADB().Move(ctx.Request.Context(), serial, req.Source, req.Destination)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to move file: %v", err)})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"source":      req.Source,
		"destination": req.Destination,
	})
}

// handleFilesDelete handles DELETE /files to delete files/directories
func handleFilesDelete(ctx *gin.Context) {
	r := runner.GetInstance()
	if r == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Runner not initialized"})
		return
	}

	sandboxId := ctx.Param("sandboxId")
	path := ctx.Query("path")
	if path == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "path query parameter is required"})
		return
	}

	serial, err := r.CVDClient.GetADBSerial(sandboxId)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Sandbox not found: %v", err)})
		return
	}

	// Use recursive delete
	err = r.CVDClient.ADB().Remove(ctx.Request.Context(), serial, path, true)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to delete: %v", err)})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"deleted": path})
}

// handleWorkspace handles workspace-related requests
func handleWorkspace(ctx *gin.Context) {
	// For Android, return a default workspace info
	ctx.JSON(http.StatusOK, gin.H{
		"name":     "android",
		"path":     "/sdcard",
		"projects": []string{},
	})
}

// HandleComputerUse handles computer use (screenshot/input) requests
func HandleComputerUse(ctx *gin.Context, path string) {
	// Remove leading slash and 'computeruse/' prefix
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimPrefix(path, "computeruse/")
	method := ctx.Request.Method

	switch {
	case path == "status":
		getComputerUseStatus(ctx)
	case method == "GET" && path == "screenshot":
		handleScreenshot(ctx)
	case method == "POST" && path == "screenshot":
		handleScreenshot(ctx)
	case method == "POST" && strings.HasPrefix(path, "keyboard/type"):
		handleKeyboardType(ctx)
	case method == "POST" && strings.HasPrefix(path, "keyboard/key"):
		handleKeyboardKey(ctx)
	case method == "POST" && strings.HasPrefix(path, "mouse/click"):
		handleMouseClick(ctx)
	case method == "POST" && strings.HasPrefix(path, "mouse/move"):
		handleMouseMove(ctx)
	case method == "POST" && strings.HasPrefix(path, "mouse/drag"):
		handleMouseDrag(ctx)
	case method == "POST" && strings.HasPrefix(path, "mouse/scroll"):
		handleMouseScroll(ctx)
	default:
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("Unknown computeruse endpoint: %s %s", method, path),
		})
	}
}

// getComputerUseStatus handles computer use status requests from toolbox path
func getComputerUseStatus(ctx *gin.Context) {
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

	status := "inactive"
	if instance.State == cuttlefish.InstanceStateRunning {
		status = "active"
	} else if instance.State == cuttlefish.InstanceStateStarting {
		status = "partial"
	}

	ctx.JSON(http.StatusOK, gin.H{"status": status})
}

// handleScreenshot handles screenshot requests
func handleScreenshot(ctx *gin.Context) {
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

	screenshot, err := r.CVDClient.ADB().Screencap(ctx.Request.Context(), serial)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to capture screenshot: %v", err)})
		return
	}

	// Check Accept header for response format
	accept := ctx.GetHeader("Accept")
	if strings.Contains(accept, "image/png") || ctx.Query("format") == "png" {
		ctx.Data(http.StatusOK, "image/png", screenshot)
	} else {
		// Return base64 encoded by default
		encoded := base64.StdEncoding.EncodeToString(screenshot)
		ctx.JSON(http.StatusOK, gin.H{
			"image":    encoded,
			"format":   "png",
			"encoding": "base64",
		})
	}
}

// KeyboardTypeRequest represents a keyboard type request
type KeyboardTypeRequest struct {
	Text string `json:"text" binding:"required"`
}

// handleKeyboardType handles keyboard text input
func handleKeyboardType(ctx *gin.Context) {
	r := runner.GetInstance()
	if r == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Runner not initialized"})
		return
	}

	sandboxId := ctx.Param("sandboxId")

	var req KeyboardTypeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	serial, err := r.CVDClient.GetADBSerial(sandboxId)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Sandbox not found: %v", err)})
		return
	}

	err = r.CVDClient.ADB().TypeText(ctx.Request.Context(), serial, req.Text)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to type text: %v", err)})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// KeyboardKeyRequest represents a keyboard key event request
type KeyboardKeyRequest struct {
	Key string `json:"key" binding:"required"` // Android keycode (e.g., "KEYCODE_HOME", "KEYCODE_BACK", "66" for ENTER)
}

// handleKeyboardKey handles keyboard key events
func handleKeyboardKey(ctx *gin.Context) {
	r := runner.GetInstance()
	if r == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Runner not initialized"})
		return
	}

	sandboxId := ctx.Param("sandboxId")

	var req KeyboardKeyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	serial, err := r.CVDClient.GetADBSerial(sandboxId)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Sandbox not found: %v", err)})
		return
	}

	err = r.CVDClient.ADB().KeyEvent(ctx.Request.Context(), serial, req.Key)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to send key event: %v", err)})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// MouseClickRequest represents a mouse click request
type MouseClickRequest struct {
	X      int    `json:"x" binding:"required"`
	Y      int    `json:"y" binding:"required"`
	Button string `json:"button,omitempty"` // "left" (default), "right", "middle" - Android only supports tap
}

// handleMouseClick handles mouse/tap click events
func handleMouseClick(ctx *gin.Context) {
	r := runner.GetInstance()
	if r == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Runner not initialized"})
		return
	}

	sandboxId := ctx.Param("sandboxId")

	var req MouseClickRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	serial, err := r.CVDClient.GetADBSerial(sandboxId)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Sandbox not found: %v", err)})
		return
	}

	err = r.CVDClient.ADB().Tap(ctx.Request.Context(), serial, req.X, req.Y)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to tap: %v", err)})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// MouseMoveRequest represents a mouse move request
type MouseMoveRequest struct {
	X int `json:"x" binding:"required"`
	Y int `json:"y" binding:"required"`
}

// handleMouseMove handles mouse move events (Android doesn't have cursor, returns OK)
func handleMouseMove(ctx *gin.Context) {
	// Android doesn't have a mouse cursor concept
	ctx.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"note":   "Android does not have a mouse cursor. Use tap/swipe instead.",
	})
}

// MouseDragRequest represents a mouse drag/swipe request
type MouseDragRequest struct {
	StartX   int `json:"startX" binding:"required"`
	StartY   int `json:"startY" binding:"required"`
	EndX     int `json:"endX" binding:"required"`
	EndY     int `json:"endY" binding:"required"`
	Duration int `json:"duration,omitempty"` // milliseconds
}

// handleMouseDrag handles mouse drag/swipe events
func handleMouseDrag(ctx *gin.Context) {
	r := runner.GetInstance()
	if r == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Runner not initialized"})
		return
	}

	sandboxId := ctx.Param("sandboxId")

	var req MouseDragRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	serial, err := r.CVDClient.GetADBSerial(sandboxId)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Sandbox not found: %v", err)})
		return
	}

	duration := req.Duration
	if duration <= 0 {
		duration = 300 // default 300ms
	}

	err = r.CVDClient.ADB().Swipe(ctx.Request.Context(), serial, req.StartX, req.StartY, req.EndX, req.EndY, duration)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to swipe: %v", err)})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// MouseScrollRequest represents a scroll request
type MouseScrollRequest struct {
	X         int `json:"x" binding:"required"`
	Y         int `json:"y" binding:"required"`
	Direction int `json:"direction"` // positive = down, negative = up
	Amount    int `json:"amount,omitempty"`
}

// handleMouseScroll handles scroll events (implemented as swipe on Android)
func handleMouseScroll(ctx *gin.Context) {
	r := runner.GetInstance()
	if r == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Runner not initialized"})
		return
	}

	sandboxId := ctx.Param("sandboxId")

	var req MouseScrollRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	serial, err := r.CVDClient.GetADBSerial(sandboxId)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Sandbox not found: %v", err)})
		return
	}

	// Convert scroll to swipe
	amount := req.Amount
	if amount <= 0 {
		amount = 200 // default scroll amount in pixels
	}

	startY := req.Y
	endY := req.Y
	if req.Direction > 0 {
		// Scroll down = swipe up
		endY = startY - amount
	} else {
		// Scroll up = swipe down
		endY = startY + amount
	}

	err = r.CVDClient.ADB().Swipe(ctx.Request.Context(), serial, req.X, startY, req.X, endY, 200)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to scroll: %v", err)})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// StreamLogcat handles logcat streaming (SSE)
func StreamLogcat(ctx *gin.Context) {
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

	// Set SSE headers
	ctx.Header("Content-Type", "text/event-stream")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")

	// Get logcat stream
	logcatReader, err := r.CVDClient.ADB().Logcat(ctx.Request.Context(), serial, "-v", "time")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to start logcat: %v", err)})
		return
	}
	defer logcatReader.Close()

	// Stream logcat output as SSE events
	buf := make([]byte, 4096)
	for {
		select {
		case <-ctx.Request.Context().Done():
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
