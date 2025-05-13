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

// PullImage godoc
//
//	@Tags			images
//	@Summary		Pull a Docker image
//	@Description	Pull a Docker image from a registry
//	@Param			request	body		dto.PullImageRequestDTO	true	"Pull image"
//	@Success		200		{string}	string					"Image successfully pulled"
//	@Failure		400		{object}	common.ErrorResponse
//	@Failure		401		{object}	common.ErrorResponse
//	@Failure		404		{object}	common.ErrorResponse
//	@Failure		409		{object}	common.ErrorResponse
//	@Failure		500		{object}	common.ErrorResponse
//
//	@Router			/images/pull [post]
//
//	@id				PullImage
func PullImage(ctx *gin.Context) {
	var request dto.PullImageRequestDTO
	err := ctx.ShouldBindJSON(&request)
	if err != nil {
		ctx.Error(common.NewInvalidBodyRequestError(err))
		return
	}

	runner := runner.GetInstance(nil)

	err = runner.Docker.PullImage(ctx.Request.Context(), request.Image, request.Registry)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, "Image pulled successfully")
}

// BuildImage godoc
//
//	@Tags			images
//	@Summary		Build a Docker image
//	@Description	Build a Docker image from a Dockerfile and context hashes
//	@Param			request	body		dto.BuildImageRequestDTO	true	"Build image request"
//	@Success		200		{string}	string						"Image successfully built"
//	@Failure		400		{object}	common.ErrorResponse
//	@Failure		401		{object}	common.ErrorResponse
//	@Failure		404		{object}	common.ErrorResponse
//	@Failure		409		{object}	common.ErrorResponse
//	@Failure		500		{object}	common.ErrorResponse
//
//	@Router			/images/build [post]
//
//	@id				BuildImage
func BuildImage(ctx *gin.Context) {
	var request dto.BuildImageRequestDTO
	err := ctx.ShouldBindJSON(&request)
	if err != nil {
		ctx.Error(common.NewInvalidBodyRequestError(err))
		return
	}

	if !strings.Contains(request.Image, ":") || strings.HasSuffix(request.Image, ":") {
		ctx.Error(common.NewBadRequestError(errors.New("image name must include a valid tag")))
		return
	}

	runner := runner.GetInstance(nil)

	err = runner.Docker.BuildImage(ctx.Request.Context(), request)
	if err != nil {
		ctx.Error(err)
		return
	}

	tag := request.Image

	if request.PushToInternalRegistry {
		if request.Registry.Project == nil {
			ctx.Error(common.NewBadRequestError(errors.New("project is required when pushing to internal registry")))
			return
		}
		tag = fmt.Sprintf("%s/%s/%s", request.Registry.Url, *request.Registry.Project, request.Image)
	}

	err = runner.Docker.TagImage(ctx.Request.Context(), request.Image, tag)
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

	ctx.JSON(http.StatusOK, "Image built successfully")
}

// ImageExists godoc
//
//	@Tags			images
//	@Summary		Check if a Docker image exists
//	@Description	Check if a specified Docker image exists locally
//	@Produce		json
//	@Param			image	query		string	true	"Image name and tag"	example:"nginx:latest"
//	@Success		200		{object}	ImageExistsResponse
//	@Failure		400		{object}	common.ErrorResponse
//	@Failure		401		{object}	common.ErrorResponse
//	@Failure		404		{object}	common.ErrorResponse
//	@Failure		409		{object}	common.ErrorResponse
//	@Failure		500		{object}	common.ErrorResponse
//	@Router			/images/exists [get]
//
//	@id				ImageExists
func ImageExists(ctx *gin.Context) {
	image := ctx.Query("image")
	if image == "" {
		ctx.Error(common.NewBadRequestError(errors.New("image parameter is required")))
		return
	}

	runner := runner.GetInstance(nil)

	exists, err := runner.Docker.ImageExists(ctx.Request.Context(), image, false)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, ImageExistsResponse{
		Exists: exists,
	})
}

// RemoveImage godoc
//
//	@Tags			images
//	@Summary		Remove a Docker image
//	@Description	Remove a specified Docker image from the local system
//	@Produce		json
//	@Param			image	query		string	true	"Image name and tag"	example:"nginx:latest"
//	@Success		200		{string}	string	"Image successfully removed"
//	@Failure		400		{object}	common.ErrorResponse
//	@Failure		401		{object}	common.ErrorResponse
//	@Failure		404		{object}	common.ErrorResponse
//	@Failure		409		{object}	common.ErrorResponse
//	@Failure		500		{object}	common.ErrorResponse
//	@Router			/images/remove [post]
//
//	@id				RemoveImage
func RemoveImage(ctx *gin.Context) {
	image := ctx.Query("image")
	if image == "" {
		ctx.Error(common.NewBadRequestError(errors.New("image parameter is required")))
		return
	}

	runner := runner.GetInstance(nil)

	err := runner.Docker.RemoveImage(ctx.Request.Context(), image, true)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, "Image removed successfully")
}

type ImageExistsResponse struct {
	Exists bool `json:"exists" example:"true"`
} //	@name	ImageExistsResponse

// GetBuildLogs godoc
//
//	@Tags			images
//	@Summary		Get build logs
//	@Description	Stream build logs
//	@Param			imageRef	query		string	true	"Image ID or image ref without the tag"
//	@Param			follow		query		boolean	false	"Whether to follow the log output"
//	@Success		200			{string}	string	"Build logs stream"
//	@Failure		400			{object}	common.ErrorResponse
//	@Failure		401			{object}	common.ErrorResponse
//	@Failure		404			{object}	common.ErrorResponse
//	@Failure		500			{object}	common.ErrorResponse
//
//	@Router			/images/logs [get]
//
//	@id				GetBuildLogs
func GetBuildLogs(ctx *gin.Context) {
	imageRef := ctx.Query("imageRef")
	if imageRef == "" {
		ctx.Error(common.NewBadRequestError(errors.New("imageRef parameter is required")))
		return
	}

	follow := ctx.Query("follow") == "true"

	logFilePath, err := config.GetBuildLogFilePath(imageRef)
	if err != nil {
		ctx.Error(common.NewCustomError(http.StatusInternalServerError, err.Error(), "INTERNAL_SERVER_ERROR"))
		return
	}

	if _, err := os.Stat(logFilePath); os.IsNotExist(err) {
		ctx.Error(common.NewNotFoundError(fmt.Errorf("build logs not found for ref: %s", imageRef)))
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

	checkImageRef := imageRef

	// Fixed tag for instances where we are not looking for an entry with image ID
	if strings.HasPrefix(imageRef, "daytona") {
		checkImageRef = imageRef + ":daytona"
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
		exists, err := runner.Docker.ImageExists(ctx.Request.Context(), checkImageRef, false)
		if err != nil {
			log.Errorf("Error checking build status: %v", err)
			break
		}

		if exists {
			// If image exists, build is complete, allow time for the last logs to be written and break the loop
			time.Sleep(1 * time.Second)
			break
		}

		time.Sleep(250 * time.Millisecond)
	}
}
