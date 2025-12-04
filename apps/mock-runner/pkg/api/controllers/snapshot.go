// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/daytonaio/mock-runner/pkg/api/dto"
	"github.com/daytonaio/mock-runner/pkg/runner"
	"github.com/gin-gonic/gin"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
)

// TagImage godoc
//
//	@Tags			snapshots
//	@Summary		Tag an image
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

	runner := runner.GetInstance(nil)

	exists, err := runner.Mock.ImageExists(ctx.Request.Context(), request.SourceImage, false)
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

	if err := runner.Mock.TagImage(ctx.Request.Context(), request.SourceImage, request.TargetImage); err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, "Image tagged successfully")
}

// PullSnapshot godoc
//
//	@Tags			snapshots
//	@Summary		Pull a snapshot
//	@Description	Pull a snapshot from a registry and optionally push to another registry
//	@Param			request	body		dto.PullSnapshotRequestDTO	true	"Pull snapshot"
//	@Success		200		{string}	string						"Snapshot successfully pulled"
//	@Failure		400		{object}	common_errors.ErrorResponse
//	@Failure		401		{object}	common_errors.ErrorResponse
//	@Failure		404		{object}	common_errors.ErrorResponse
//	@Failure		409		{object}	common_errors.ErrorResponse
//	@Failure		500		{object}	common_errors.ErrorResponse
//
//	@Router			/snapshots/pull [post]
//
//	@id				PullSnapshot
func PullSnapshot(ctx *gin.Context) {
	var request dto.PullSnapshotRequestDTO
	err := ctx.ShouldBindJSON(&request)
	if err != nil {
		ctx.Error(common_errors.NewInvalidBodyRequestError(err))
		return
	}

	runner := runner.GetInstance(nil)

	// Pull the image
	err = runner.Mock.PullImage(ctx.Request.Context(), request.Snapshot, request.Registry)
	if err != nil {
		ctx.Error(err)
		return
	}

	if request.DestinationRegistry != nil {
		if request.DestinationRegistry.Project == nil {
			ctx.Error(common_errors.NewBadRequestError(errors.New("project is required when pushing to registry")))
			return
		}

		var targetRef string

		if request.DestinationRef != nil {
			targetRef = *request.DestinationRef
		} else {
			imageInfo, err := runner.Mock.GetImageInfo(ctx.Request.Context(), request.Snapshot)
			if err != nil {
				ctx.Error(err)
				return
			}

			ref := "daytona-" + getHashWithoutPrefix(imageInfo.Hash) + ":daytona"
			targetRef = fmt.Sprintf("%s/%s/%s", request.DestinationRegistry.Url, *request.DestinationRegistry.Project, ref)
		}

		err = runner.Mock.TagImage(ctx.Request.Context(), request.Snapshot, targetRef)
		if err != nil {
			ctx.Error(err)
			return
		}

		err = runner.Mock.PushImage(ctx.Request.Context(), targetRef, request.DestinationRegistry)
		if err != nil {
			ctx.Error(err)
			return
		}
	}

	ctx.JSON(http.StatusOK, "Snapshot pulled successfully")
}

// BuildSnapshot godoc
//
//	@Tags			snapshots
//	@Summary		Build a snapshot
//	@Description	Build a snapshot from a Dockerfile and context hashes
//	@Param			request	body		dto.BuildSnapshotRequestDTO	true	"Build snapshot request"
//	@Success		200		{string}	string						"Snapshot successfully built"
//	@Failure		400		{object}	common_errors.ErrorResponse
//	@Failure		401		{object}	common_errors.ErrorResponse
//	@Failure		404		{object}	common_errors.ErrorResponse
//	@Failure		409		{object}	common_errors.ErrorResponse
//	@Failure		500		{object}	common_errors.ErrorResponse
//
//	@Router			/snapshots/build [post]
//
//	@id				BuildSnapshot
func BuildSnapshot(ctx *gin.Context) {
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

	runner := runner.GetInstance(nil)

	err = runner.Mock.BuildImage(ctx.Request.Context(), request)
	if err != nil {
		ctx.Error(err)
		return
	}

	tag := request.Snapshot

	if request.PushToInternalRegistry {
		if request.Registry.Project == nil {
			ctx.Error(common_errors.NewBadRequestError(errors.New("project is required when pushing to internal registry")))
			return
		}
		tag = fmt.Sprintf("%s/%s/%s", request.Registry.Url, *request.Registry.Project, request.Snapshot)
	}

	err = runner.Mock.TagImage(ctx.Request.Context(), request.Snapshot, tag)
	if err != nil {
		ctx.Error(err)
		return
	}

	if request.PushToInternalRegistry {
		err = runner.Mock.PushImage(ctx.Request.Context(), tag, request.Registry)
		if err != nil {
			ctx.Error(err)
			return
		}
	}

	ctx.JSON(http.StatusOK, "Snapshot built successfully")
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

	runner := runner.GetInstance(nil)

	exists, err := runner.Mock.ImageExists(ctx.Request.Context(), snapshot, false)
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
func RemoveSnapshot(ctx *gin.Context) {
	snapshot := ctx.Query("snapshot")
	if snapshot == "" {
		ctx.Error(common_errors.NewBadRequestError(errors.New("snapshot parameter is required")))
		return
	}

	runner := runner.GetInstance(nil)

	err := runner.Mock.RemoveImage(ctx.Request.Context(), snapshot, true)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, "Snapshot removed successfully")
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
func GetBuildLogs(ctx *gin.Context) {
	snapshotRef := ctx.Query("snapshotRef")
	if snapshotRef == "" {
		ctx.Error(common_errors.NewBadRequestError(errors.New("snapshotRef parameter is required")))
		return
	}

	// For mock, return mock build logs
	buildId := snapshotRef
	if colonIndex := strings.Index(snapshotRef, ":"); colonIndex != -1 {
		buildId = snapshotRef[:colonIndex]
	}

	logDir := filepath.Join(os.TempDir(), "mock-runner", "builds")
	logPath := filepath.Join(logDir, buildId)

	ctx.Header("Content-Type", "application/octet-stream")

	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		// Return mock log if file doesn't exist
		mockLog := fmt.Sprintf("Mock build log for %s\nBuild completed successfully.\n", snapshotRef)
		ctx.Writer.Write([]byte(mockLog))
		return
	}

	content, err := os.ReadFile(logPath)
	if err != nil {
		ctx.Error(common_errors.NewCustomError(http.StatusInternalServerError, err.Error(), "INTERNAL_SERVER_ERROR"))
		return
	}

	ctx.Writer.Write(content)
}

// GetSnapshotInfo godoc
//
//	@Tags			snapshots
//	@Summary		Get snapshot information
//	@Description	Get information about a specified snapshot including size and entrypoint
//	@Produce		json
//	@Param			snapshot	query		string	true	"Snapshot name and tag"	example:"nginx:latest"
//	@Success		200			{object}	SnapshotInfoResponse
//	@Failure		400			{object}	common_errors.ErrorResponse
//	@Failure		401			{object}	common_errors.ErrorResponse
//	@Failure		404			{object}	common_errors.ErrorResponse
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

	runner := runner.GetInstance(nil)

	exists, err := runner.Mock.ImageExists(ctx.Request.Context(), snapshot, false)
	if err != nil {
		ctx.Error(err)
		return
	}

	if !exists {
		ctx.Error(common_errors.NewNotFoundError(fmt.Errorf("snapshot not found: %s", snapshot)))
		return
	}

	info, err := runner.Mock.GetImageInfo(ctx.Request.Context(), snapshot)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, SnapshotInfoResponse{
		Name:       snapshot,
		SizeGB:     float64(info.Size) / (1024 * 1024 * 1024),
		Entrypoint: info.Entrypoint,
		Cmd:        info.Cmd,
		Hash:       getHashWithoutPrefix(info.Hash),
	})
}

type SnapshotInfoResponse struct {
	Name       string   `json:"name" example:"nginx:latest"`
	SizeGB     float64  `json:"sizeGB" example:"0.13"`
	Entrypoint []string `json:"entrypoint,omitempty" example:"[\"nginx\",\"-g\",\"daemon off;\"]"`
	Cmd        []string `json:"cmd,omitempty" example:"[\"nginx\",\"-g\",\"daemon off;\"]"`
	Hash       string   `json:"hash,omitempty" example:"a7be6198544f09a75b26e6376459b47c5b9972e7351d440e092c4faa9ea064ff"`
} //	@name	SnapshotInfoResponse

func getHashWithoutPrefix(hash string) string {
	return strings.TrimPrefix(hash, "sha256:")
}



