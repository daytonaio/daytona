// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package controllers

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/daytonaio/runner/cmd/runner/config"
	"github.com/daytonaio/runner/pkg/api/dto"
	"github.com/daytonaio/runner/pkg/common"
	"github.com/daytonaio/runner/pkg/runner"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// PullSnapshot godoc
//
//	@Tags			snapshots
//	@Summary		Pull a snapshot
//	@Description	Pull a snapshot from a registry
//	@Param			request	body		dto.PullSnapshotRequestDTO	true	"Pull snapshot"
//	@Success		200		{string}	string						"Snapshot successfully pulled"
//	@Failure		400		{object}	common.ErrorResponse
//	@Failure		401		{object}	common.ErrorResponse
//	@Failure		404		{object}	common.ErrorResponse
//	@Failure		409		{object}	common.ErrorResponse
//	@Failure		500		{object}	common.ErrorResponse
//
//	@Router			/snapshots/pull [post]
//
//	@id				PullSnapshot
func PullSnapshot(ctx *gin.Context) {
	var request dto.PullSnapshotRequestDTO
	err := ctx.ShouldBindJSON(&request)
	if err != nil {
		ctx.Error(common.NewInvalidBodyRequestError(err))
		return
	}

	runner := runner.GetInstance(nil)

	err = runner.Docker.PullImage(ctx.Request.Context(), request.Snapshot, request.Registry)
	if err != nil {
		ctx.Error(err)
		return
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
//	@Failure		400		{object}	common.ErrorResponse
//	@Failure		401		{object}	common.ErrorResponse
//	@Failure		404		{object}	common.ErrorResponse
//	@Failure		409		{object}	common.ErrorResponse
//	@Failure		500		{object}	common.ErrorResponse
//
//	@Router			/snapshots/build [post]
//
//	@id				BuildSnapshot
func BuildSnapshot(ctx *gin.Context) {
	var request dto.BuildSnapshotRequestDTO
	err := ctx.ShouldBindJSON(&request)
	if err != nil {
		ctx.Error(common.NewInvalidBodyRequestError(err))
		return
	}

	if !strings.Contains(request.Snapshot, ":") || strings.HasSuffix(request.Snapshot, ":") {
		ctx.Error(common.NewBadRequestError(errors.New("snapshot name must include a valid tag")))
		return
	}

	runner := runner.GetInstance(nil)

	err = runner.Docker.BuildImage(ctx.Request.Context(), request)
	if err != nil {
		ctx.Error(err)
		return
	}

	tag := request.Snapshot

	if request.PushToInternalRegistry {
		if request.Registry.Project == nil {
			ctx.Error(common.NewBadRequestError(errors.New("project is required when pushing to internal registry")))
			return
		}
		tag = fmt.Sprintf("%s/%s/%s", request.Registry.Url, *request.Registry.Project, request.Snapshot)
	}

	err = runner.Docker.TagImage(ctx.Request.Context(), request.Snapshot, tag)
	if err != nil {
		ctx.Error(err)
		return
	}

	if request.PushToInternalRegistry {
		err = runner.Docker.PushImage(ctx.Request.Context(), tag, request.Registry)
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
//	@Failure		400			{object}	common.ErrorResponse
//	@Failure		401			{object}	common.ErrorResponse
//	@Failure		404			{object}	common.ErrorResponse
//	@Failure		409			{object}	common.ErrorResponse
//	@Failure		500			{object}	common.ErrorResponse
//	@Router			/snapshots/exists [get]
//
//	@id				SnapshotExists
func SnapshotExists(ctx *gin.Context) {
	snapshot := ctx.Query("snapshot")
	if snapshot == "" {
		ctx.Error(common.NewBadRequestError(errors.New("snapshot parameter is required")))
		return
	}

	runner := runner.GetInstance(nil)

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
//	@Failure		400			{object}	common.ErrorResponse
//	@Failure		401			{object}	common.ErrorResponse
//	@Failure		404			{object}	common.ErrorResponse
//	@Failure		409			{object}	common.ErrorResponse
//	@Failure		500			{object}	common.ErrorResponse
//	@Router			/snapshots/remove [post]
//
//	@id				RemoveSnapshot
func RemoveSnapshot(ctx *gin.Context) {
	snapshot := ctx.Query("snapshot")
	if snapshot == "" {
		ctx.Error(common.NewBadRequestError(errors.New("snapshot parameter is required")))
		return
	}

	runner := runner.GetInstance(nil)

	err := runner.Docker.RemoveImage(ctx.Request.Context(), snapshot, true)
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
//	@Param			snapshotRef	query		string	true	"Snapshot ID or snapshot ref without the tag"
//	@Param			follow		query		boolean	false	"Whether to follow the log output"
//	@Success		200			{string}	string	"Build logs stream"
//	@Failure		400			{object}	common.ErrorResponse
//	@Failure		401			{object}	common.ErrorResponse
//	@Failure		404			{object}	common.ErrorResponse
//	@Failure		500			{object}	common.ErrorResponse
//
//	@Router			/snapshots/logs [get]
//
//	@id				GetBuildLogs
func GetBuildLogs(ctx *gin.Context) {
	snapshotRef := ctx.Query("snapshotRef")
	if snapshotRef == "" {
		ctx.Error(common.NewBadRequestError(errors.New("snapshotRef parameter is required")))
		return
	}

	follow := ctx.Query("follow") == "true"

	logFilePath, err := config.GetBuildLogFilePath(snapshotRef)
	if err != nil {
		ctx.Error(common.NewCustomError(http.StatusInternalServerError, err.Error(), "INTERNAL_SERVER_ERROR"))
		return
	}

	if _, err := os.Stat(logFilePath); os.IsNotExist(err) {
		ctx.Error(common.NewNotFoundError(fmt.Errorf("build logs not found for ref: %s", snapshotRef)))
		return
	}

	ctx.Header("Content-Type", "application/octet-stream")

	file, err := os.Open(logFilePath)
	if err != nil {
		ctx.Error(common.NewCustomError(http.StatusInternalServerError, err.Error(), "INTERNAL_SERVER_ERROR"))
		return
	}
	defer file.Close()

	// If not following, just return the entire file content
	if !follow {
		_, err = io.Copy(ctx.Writer, file)
		if err != nil {
			ctx.Error(common.NewCustomError(http.StatusInternalServerError, err.Error(), "INTERNAL_SERVER_ERROR"))
		}
		return
	}

	reader := bufio.NewReader(file)
	runner := runner.GetInstance(nil)

	checkSnapshotRef := snapshotRef

	// Fixed tag for instances where we are not looking for an entry with snapshot ID
	if strings.HasPrefix(snapshotRef, "daytona") {
		checkSnapshotRef = snapshotRef + ":daytona"
	}

	flusher, ok := ctx.Writer.(http.Flusher)
	if !ok {
		ctx.Error(common.NewCustomError(http.StatusInternalServerError, "Streaming not supported", "STREAMING_NOT_SUPPORTED"))
		return
	}

	go func() {
		for {
			line, err := reader.ReadBytes('\n')
			if err != nil && err != io.EOF {
				log.Errorf("Error reading log file: %v", err)
				break
			}

			if len(line) > 0 {
				_, writeErr := ctx.Writer.Write(line)
				if writeErr != nil {
					log.Errorf("Error writing to response: %v", writeErr)
					break
				}
				flusher.Flush()
			}
		}
	}()

	for {
		exists, err := runner.Docker.ImageExists(ctx.Request.Context(), checkSnapshotRef, false)
		if err != nil {
			log.Errorf("Error checking build status: %v", err)
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
