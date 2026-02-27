// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package controllers

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"context"

	"github.com/daytonaio/runner/cmd/runner/config"
	"github.com/daytonaio/runner/pkg/api/dto"
	"github.com/daytonaio/runner/pkg/runner"
	"github.com/gin-gonic/gin"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
)

// TagImage godoc
//
//	@Tags			snapshots
//	@Summary		Tag an image
//	@Deprecated		New snapshot tags are sent in PullSnapshot
//	@Description	Tag an existing local image with a new target reference
//	@Param			request	body		dto.TagImageRequestDTO	true	"Tag image request"
//	@Success		200		{string}	string					"Image successfully tagged"
//	@Failure		400		{object}	common_errors.ErrorResponse
//	@Failure		401		{object}	common_errors.ErrorResponse
//	@Failure		404		{object}	common_errors.ErrorResponse
//	@Failure		409		{object}	common_errors.ErrorResponse
//	@Failure		500		{object}	common_errors.ErrorResponse
//
//	@Router			/snapshots/tag [post]
//
//	@id				TagImage
func TagImage(ctx *gin.Context) {
	var request dto.TagImageRequestDTO
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.Error(common_errors.NewInvalidBodyRequestError(err))
		return
	}

	runner, err := runner.GetInstance(nil)
	if err != nil {
		ctx.Error(err)
		return
	}

	exists, err := runner.Docker.ImageExists(ctx.Request.Context(), request.SourceImage, false)
	if err != nil {
		ctx.Error(err)
		return
	}

	if !exists {
		ctx.Error(common_errors.NewNotFoundError(fmt.Errorf("source image not found: %s", request.SourceImage)))
		return
	}

	if !strings.Contains(request.TargetImage, ":") || strings.HasSuffix(request.TargetImage, ":") {
		ctx.Error(common_errors.NewBadRequestError(errors.New("targetImage must include a valid tag")))
		return
	}

	if err := runner.Docker.TagImage(ctx.Request.Context(), request.SourceImage, request.TargetImage); err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, "Image tagged successfully")
}

// PullSnapshot godoc
//
//	@Tags			snapshots
//	@Summary		Pull a snapshot
//	@Description	Pull a snapshot from a registry and optionally push to another registry. The operation runs asynchronously and returns 202 immediately.
//	@Param			request	body		dto.PullSnapshotRequestDTO	true	"Pull snapshot"
//	@Success		202		{string}	string						"Snapshot pull started"
//	@Failure		400		{object}	common_errors.ErrorResponse
//	@Failure		401		{object}	common_errors.ErrorResponse
//	@Failure		404		{object}	common_errors.ErrorResponse
//	@Failure		409		{object}	common_errors.ErrorResponse
//	@Failure		500		{object}	common_errors.ErrorResponse
//
//	@Router			/snapshots/pull [post]
//
//	@id				PullSnapshot
func PullSnapshot(generalCtx context.Context, logger *slog.Logger) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		var request dto.PullSnapshotRequestDTO
		err := ctx.ShouldBindJSON(&request)
		if err != nil {
			ctx.Error(common_errors.NewInvalidBodyRequestError(err))
			return
		}

		runner, err := runner.GetInstance(nil)
		if err != nil {
			ctx.Error(err)
			return
		}

		cacheKey := request.Snapshot
		if request.DestinationRef != nil {
			cacheKey = *request.DestinationRef
		}

		err = runner.SnapshotErrorCache.RemoveError(generalCtx, cacheKey)
		if err != nil {
			logger.ErrorContext(generalCtx, "Failed to remove snapshot error cache entry", "cacheKey", cacheKey, "error", err)
		}

		go func() {
			err := runner.Docker.PullSnapshot(generalCtx, request)
			if err != nil {
				logger.DebugContext(generalCtx, "Pull snapshot failed", "cacheKey", cacheKey, "error", err)
				err = runner.SnapshotErrorCache.SetError(generalCtx, cacheKey, err.Error())
				if err != nil {
					logger.ErrorContext(generalCtx, "Failed to set snapshot error cache entry", "cacheKey", cacheKey, "error", err)
				}
			} else {
				err = runner.SnapshotErrorCache.RemoveError(generalCtx, cacheKey)
				if err != nil {
					logger.ErrorContext(generalCtx, "Failed to remove snapshot error cache entry", "cacheKey", cacheKey, "error", err)
				}
			}
		}()

		ctx.JSON(http.StatusAccepted, "Snapshot pull started")
	}
}

// BuildSnapshot godoc
//
//	@Tags			snapshots
//	@Summary		Build a snapshot
//	@Description	Build a snapshot from a Dockerfile and context hashes. The operation runs asynchronously and returns 202 immediately.
//	@Param			request	body		dto.BuildSnapshotRequestDTO	true	"Build snapshot request"
//	@Success		202		{string}	string						"Snapshot build started"
//	@Failure		400		{object}	common_errors.ErrorResponse
//	@Failure		401		{object}	common_errors.ErrorResponse
//	@Failure		404		{object}	common_errors.ErrorResponse
//	@Failure		409		{object}	common_errors.ErrorResponse
//	@Failure		500		{object}	common_errors.ErrorResponse
//
//	@Router			/snapshots/build [post]
//
//	@id				BuildSnapshot
func BuildSnapshot(generalCtx context.Context, logger *slog.Logger) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		var request dto.BuildSnapshotRequestDTO
		err := ctx.ShouldBindJSON(&request)
		if err != nil {
			ctx.Error(common_errors.NewInvalidBodyRequestError(err))
			return
		}

		if !strings.Contains(request.Snapshot, ":") || strings.HasSuffix(request.Snapshot, ":") {
			ctx.Error(common_errors.NewBadRequestError(errors.New("snapshot name must include a valid tag")))
			return
		}

		runner, err := runner.GetInstance(nil)
		if err != nil {
			ctx.Error(err)
			return
		}

		err = runner.SnapshotErrorCache.RemoveError(generalCtx, request.Snapshot)
		if err != nil {
			logger.ErrorContext(generalCtx, "Failed to remove snapshot error cache entry", "cacheKey", request.Snapshot, "error", err)
		}

		go func() {
			err := runner.Docker.BuildSnapshot(generalCtx, request)
			if err != nil {
				logger.DebugContext(generalCtx, "Build snapshot failed", "cacheKey", request.Snapshot, "error", err)
				err = runner.SnapshotErrorCache.SetError(generalCtx, request.Snapshot, err.Error())
				if err != nil {
					logger.ErrorContext(generalCtx, "Failed to set snapshot error cache entry", "cacheKey", request.Snapshot, "error", err)
				}
			} else {
				err = runner.SnapshotErrorCache.RemoveError(generalCtx, request.Snapshot)
				if err != nil {
					logger.ErrorContext(generalCtx, "Failed to remove snapshot error cache entry", "cacheKey", request.Snapshot, "error", err)
				}
			}
		}()

		ctx.JSON(http.StatusAccepted, "Snapshot build started")
	}
}

// SnapshotExists godoc
//
//	@Tags			snapshots
//	@Summary		Check if a snapshot exists
//	@Description	Check if a specified snapshot exists locally
//	@Produce		json
//	@Param			snapshot	query		string	true	"Snapshot name and tag"	example:"nginx:latest"
//	@Success		200			{object}	SnapshotExistsResponse
//	@Failure		400			{object}	common_errors.ErrorResponse
//	@Failure		401			{object}	common_errors.ErrorResponse
//	@Failure		404			{object}	common_errors.ErrorResponse
//	@Failure		409			{object}	common_errors.ErrorResponse
//	@Failure		500			{object}	common_errors.ErrorResponse
//	@Router			/snapshots/exists [get]
//
//	@id				SnapshotExists
func SnapshotExists(ctx *gin.Context) {
	snapshot := ctx.Query("snapshot")
	if snapshot == "" {
		ctx.Error(common_errors.NewBadRequestError(errors.New("snapshot parameter is required")))
		return
	}

	runner, err := runner.GetInstance(nil)
	if err != nil {
		ctx.Error(err)
		return
	}

	exists, err := runner.Docker.ImageExists(ctx.Request.Context(), snapshot, false)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, SnapshotExistsResponse{
		Exists: exists,
	})
}

// RemoveSnapshot godoc
//
//	@Tags			snapshots
//	@Summary		Remove a snapshot
//	@Description	Remove a specified snapshot from the local system
//	@Produce		json
//	@Param			snapshot	query		string	true	"Snapshot name and tag"	example:"nginx:latest"
//	@Success		200			{string}	string	"Snapshot successfully removed"
//	@Failure		400			{object}	common_errors.ErrorResponse
//	@Failure		401			{object}	common_errors.ErrorResponse
//	@Failure		404			{object}	common_errors.ErrorResponse
//	@Failure		409			{object}	common_errors.ErrorResponse
//	@Failure		500			{object}	common_errors.ErrorResponse
//	@Router			/snapshots/remove [post]
//
//	@id				RemoveSnapshot
func RemoveSnapshot(logger *slog.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		snapshot := ctx.Query("snapshot")
		if snapshot == "" {
			ctx.Error(common_errors.NewBadRequestError(errors.New("snapshot parameter is required")))
			return
		}

		runner, err := runner.GetInstance(nil)
		if err != nil {
			ctx.Error(err)
			return
		}

		err = runner.Docker.RemoveImage(ctx.Request.Context(), snapshot, true)
		if err != nil {
			ctx.Error(err)
			return
		}

		err = runner.SnapshotErrorCache.RemoveError(ctx.Request.Context(), snapshot)
		if err != nil {
			logger.ErrorContext(ctx.Request.Context(), "Failed to remove snapshot error cache entry", "cacheKey", snapshot, "error", err)
		}

		ctx.JSON(http.StatusOK, "Snapshot removed successfully")
	}
}

type SnapshotExistsResponse struct {
	Exists bool `json:"exists" example:"true"`
} //	@name	SnapshotExistsResponse

// GetBuildLogs godoc
//
//	@Tags			snapshots
//	@Summary		Get build logs
//	@Description	Stream build logs
//	@Param			snapshotRef	query		string	true	"Snapshot ref"
//	@Param			follow		query		boolean	false	"Whether to follow the log output"
//	@Success		200			{string}	string	"Build logs stream"
//	@Failure		400			{object}	common_errors.ErrorResponse
//	@Failure		401			{object}	common_errors.ErrorResponse
//	@Failure		404			{object}	common_errors.ErrorResponse
//	@Failure		500			{object}	common_errors.ErrorResponse
//
//	@Router			/snapshots/logs [get]
//
//	@id				GetBuildLogs
func GetBuildLogs(logger *slog.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		reqCtx := ctx.Request.Context()
		snapshotRef := ctx.Query("snapshotRef")
		if snapshotRef == "" {
			ctx.Error(common_errors.NewBadRequestError(errors.New("snapshotRef parameter is required")))
			return
		}

		follow := ctx.Query("follow") == "true"

		logFilePath, err := config.GetBuildLogFilePath(snapshotRef)
		if err != nil {
			ctx.Error(common_errors.NewCustomError(http.StatusInternalServerError, err.Error(), "INTERNAL_SERVER_ERROR"))
			return
		}

		if _, err := os.Stat(logFilePath); os.IsNotExist(err) {
			ctx.Error(common_errors.NewNotFoundError(fmt.Errorf("build logs not found for ref: %s", snapshotRef)))
			return
		}

		ctx.Header("Content-Type", "application/octet-stream")

		file, err := os.Open(logFilePath)
		if err != nil {
			ctx.Error(common_errors.NewCustomError(http.StatusInternalServerError, err.Error(), "INTERNAL_SERVER_ERROR"))
			return
		}
		defer file.Close()

		// If not following, just return the entire file content
		if !follow {
			_, err = io.Copy(ctx.Writer, file)
			if err != nil {
				ctx.Error(common_errors.NewCustomError(http.StatusInternalServerError, err.Error(), "INTERNAL_SERVER_ERROR"))
			}
			return
		}

		reader := bufio.NewReader(file)
		runner, err := runner.GetInstance(nil)
		if err != nil {
			ctx.Error(err)
			return
		}

		checkSnapshotRef := snapshotRef

		// Fixed tag for instances where we are not looking for an entry with snapshot ID
		if strings.HasPrefix(snapshotRef, "daytona") {
			checkSnapshotRef = snapshotRef + ":daytona"
		}

		flusher, ok := ctx.Writer.(http.Flusher)
		if !ok {
			ctx.Error(common_errors.NewCustomError(http.StatusInternalServerError, "Streaming not supported", "STREAMING_NOT_SUPPORTED"))
			return
		}

		go func() {
			for {
				line, err := reader.ReadBytes('\n')
				if err != nil && err != io.EOF {
					logger.ErrorContext(reqCtx, "Error reading log file", "error", err)
					break
				}

				if len(line) > 0 {
					_, writeErr := ctx.Writer.Write(line)
					if writeErr != nil {
						logger.ErrorContext(reqCtx, "Error writing to response", "error", writeErr)
						break
					}
					flusher.Flush()
				}
			}
		}()

		for {
			exists, err := runner.Docker.ImageExists(ctx.Request.Context(), checkSnapshotRef, false)
			if err != nil {
				logger.ErrorContext(reqCtx, "Error checking build status", "error", err)
				break
			}

			if exists {
				// If snapshot exists, build is complete, allow time for the last logs to be written and break the loop
				time.Sleep(1 * time.Second)
				break
			}

			time.Sleep(250 * time.Millisecond)
		}
	}
}

// GetSnapshotInfo godoc
//
//	@Tags			snapshots
//	@Summary		Get snapshot information
//	@Description	Get information about a specified snapshot including size and entrypoint. Returns 422 if the last pull/build operation failed, with the error reason in the message.
//	@Produce		json
//	@Param			snapshot	query		string	true	"Snapshot name and tag"	example:"nginx:latest"
//	@Success		200			{object}	dto.SnapshotInfoResponse
//	@Failure		400			{object}	common_errors.ErrorResponse
//	@Failure		401			{object}	common_errors.ErrorResponse
//	@Failure		404			{object}	common_errors.ErrorResponse
//	@Failure		422			{object}	common_errors.ErrorResponse
//	@Failure		500			{object}	common_errors.ErrorResponse
//	@Router			/snapshots/info [get]
//
//	@id				GetSnapshotInfo
func GetSnapshotInfo(ctx *gin.Context) {
	snapshot := ctx.Query("snapshot")
	if snapshot == "" {
		ctx.Error(common_errors.NewBadRequestError(errors.New("snapshot parameter is required")))
		return
	}

	runner, err := runner.GetInstance(nil)
	if err != nil {
		ctx.Error(err)
		return
	}

	exists, err := runner.Docker.ImageExists(ctx.Request.Context(), snapshot, false)
	if err != nil {
		ctx.Error(err)
		return
	}

	if !exists {
		errReason, err := runner.SnapshotErrorCache.GetError(ctx.Request.Context(), snapshot)
		if err == nil && errReason != nil {
			ctx.Error(common_errors.NewUnprocessableEntityError(errors.New(*errReason)))
			return
		}
		ctx.Error(common_errors.NewNotFoundError(fmt.Errorf("snapshot not found: %s", snapshot)))
		return
	}

	info, err := runner.Docker.GetImageInfo(ctx.Request.Context(), snapshot)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, dto.SnapshotInfoResponse{
		Name:       snapshot,
		SizeGB:     float64(info.Size) / (1024 * 1024 * 1024), // Convert bytes to GB
		Entrypoint: info.Entrypoint,
		Cmd:        info.Cmd,
		Hash:       dto.HashWithoutPrefix(info.Hash),
	})
}

// InspectSnapshotInRegistry godoc
//
//	@Tags			snapshots
//	@Summary		Inspect a snapshot in a registry
//	@Description	Inspect a specified snapshot in a registry
//	@Produce		json
//	@Param			request	body		dto.InspectSnapshotInRegistryRequestDTO	true	"Inspect snapshot in registry request"
//	@Success		200		{object}	dto.SnapshotDigestResponse
//	@Failure		400		{object}	common_errors.ErrorResponse
//	@Failure		401		{object}	common_errors.ErrorResponse
//	@Failure		404		{object}	common_errors.ErrorResponse
//	@Failure		500		{object}	common_errors.ErrorResponse
//
//	@Router			/snapshots/inspect [post]
//	@id				InspectSnapshotInRegistry
func InspectSnapshotInRegistry(ctx *gin.Context) {
	var request dto.InspectSnapshotInRegistryRequestDTO
	err := ctx.ShouldBindJSON(&request)
	if err != nil {
		ctx.Error(common_errors.NewInvalidBodyRequestError(err))
		return
	}

	runner, err := runner.GetInstance(nil)
	if err != nil {
		ctx.Error(err)
		return
	}

	digest, err := runner.Docker.InspectImageInRegistry(ctx.Request.Context(), request.Snapshot, request.Registry)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, dto.SnapshotDigestResponse{
		Hash:   dto.HashWithoutPrefix(digest.Digest),
		SizeGB: float64(digest.Size) / (1024 * 1024 * 1024),
	})
}
