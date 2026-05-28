// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package recordingdashboard

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
	"github.com/daytonaio/daemon/pkg/recording"
	recordingcontroller "github.com/daytonaio/daemon/pkg/toolbox/computeruse/recording"
	"github.com/daytonaio/daemon/pkg/toolbox/config"
	"github.com/gin-gonic/gin"
)

// DashboardServer serves the recording dashboard
type DashboardServer struct {
	logger           *slog.Logger
	recordingService *recording.RecordingService
}

// NewDashboardServer creates a new dashboard server
func NewDashboardServer(logger *slog.Logger, recordingService *recording.RecordingService) *DashboardServer {
	return &DashboardServer{
		logger:           logger.With(slog.String("component", "recordings_dashboard")),
		recordingService: recordingService,
	}
}

// Start starts the dashboard server on the configured port
func (s *DashboardServer) Start() error {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(common_errors.NewErrorMiddleware("DAYTONA_DAEMON", nil))

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
	s.logger.Info("Starting recording dashboard", "port", config.RECORDING_DASHBOARD_PORT)

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
		ctx.Error(common_errors.NewForbiddenError(errors.New("access denied")))
		return
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		ctx.Error(common_errors.NewNotFoundError(errors.New("file not found")))
		return
	}

	ctx.File(filePath)
}

func (s *DashboardServer) listRecordings(ctx *gin.Context) {
	recordings, err := s.recordingService.ListRecordings()
	if err != nil {
		ctx.Error(common_errors.NewInternalServerError(err))
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
		ctx.Error(common_errors.NewInvalidBodyRequestError(err))
		return
	}

	deleted := []string{}
	failed := []string{}

	// Direct calls to data provider
	for _, id := range req.IDs {
		if err := s.recordingService.DeleteRecording(id); err != nil {
			failed = append(failed, id)
			s.logger.Warn("Failed to delete recording", "id", id, "error", err)
		} else {
			deleted = append(deleted, id)
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"deleted": deleted,
		"failed":  failed,
	})
}
