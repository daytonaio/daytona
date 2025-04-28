// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/daytonaio/runner/pkg/api/dto"
	"github.com/daytonaio/runner/pkg/common"
	"github.com/daytonaio/runner/pkg/runner"
	"github.com/gin-gonic/gin"
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
		// TODO: parameterize project ID
		tag = fmt.Sprintf("%s/daytona/%s", request.Registry.Url, request.Image)
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
