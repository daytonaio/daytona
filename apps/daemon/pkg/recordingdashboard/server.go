// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package recordingdashboard

import (
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/daytonaio/daemon/pkg/recording"
	recordingcontroller "github.com/daytonaio/daemon/pkg/toolbox/computeruse/recording"
	"github.com/daytonaio/daemon/pkg/toolbox/config"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// DashboardServer serves the recording dashboard
type DashboardServer struct {
	recordingService *recording.RecordingService
}

// NewDashboardServer creates a new dashboard server
func NewDashboardServer(recordingService *recording.RecordingService) *DashboardServer {
	return &DashboardServer{
		recordingService: recordingService,
	}
}

// Start starts the dashboard server on the configured port
func (s *DashboardServer) Start() error {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	// Prepare the embedded frontend files
	// Serve the files from the embedded filesystem
	staticFS, err := fs.Sub(static, "static")
	if err != nil {
		return fmt.Errorf("failed to create sub filesystem: %w", err)
	}

	// Serve dashboard HTML from embedded files
	r.GET("/", gin.WrapH(http.FileServer(http.FS(staticFS))))

	// Serve video files
	r.GET("/videos/:filename", s.serveVideo)

	// API endpoints
	r.GET("/api/recordings", s.listRecordings)
	r.DELETE("/api/recordings", s.deleteRecordings)

	addr := fmt.Sprintf(":%d", config.RECORDING_DASHBOARD_PORT)
	log.Println("Starting recording dashboard on port", config.RECORDING_DASHBOARD_PORT)

	err = r.Run(addr)
	return err
}

func (s *DashboardServer) serveVideo(ctx *gin.Context) {
	filename := ctx.Param("filename")
	recordingsDir := s.recordingService.GetRecordingsDir()
	filePath := filepath.Join(recordingsDir, filename)

	// Security check - prevent path traversal
	// filepath.Rel returns a path with ".." if target is outside base directory
	rel, err := filepath.Rel(recordingsDir, filePath)
	if err != nil || strings.Contains(rel, "..") {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	ctx.File(filePath)
}

func (s *DashboardServer) listRecordings(ctx *gin.Context) {
	recordings, err := s.recordingService.ListRecordings()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	recordingDTOs := make([]recordingcontroller.RecordingDTO, 0, len(recordings))
	for _, rec := range recordings {
		recordingDTOs = append(recordingDTOs, *recordingcontroller.RecordingToDTO(&rec))
	}

	ctx.JSON(http.StatusOK, gin.H{"recordings": recordingDTOs})
}

type deleteRequest struct {
	IDs []string `json:"ids"`
}

func (s *DashboardServer) deleteRecordings(ctx *gin.Context) {
	var req deleteRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	deleted := []string{}
	failed := []string{}

	// Direct calls to data provider
	for _, id := range req.IDs {
		if err := s.recordingService.DeleteRecording(id); err != nil {
			failed = append(failed, id)
			log.Warnf("Failed to delete recording %s: %v", id, err)
		} else {
			deleted = append(deleted, id)
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"deleted": deleted,
		"failed":  failed,
	})
}
