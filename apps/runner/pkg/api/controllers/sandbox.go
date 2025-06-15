// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package controllers

import (
	"net/http"

	"github.com/daytonaio/runner/pkg/api/dto"
	"github.com/daytonaio/runner/pkg/common"
	"github.com/daytonaio/runner/pkg/models/enums"
	"github.com/daytonaio/runner/pkg/runner"
	"github.com/gin-gonic/gin"
)

// Create 			godoc
//
//	@Tags			sandbox
//	@Summary		Create a sandbox
//	@Description	Create a sandbox
//	@Param			sandbox	body	dto.CreateSandboxDTO	true	"Create sandbox"
//	@Produce		json
//	@Success		201	{string}	containerId
//	@Failure		400	{object}	common.ErrorResponse
//	@Failure		401	{object}	common.ErrorResponse
//	@Failure		404	{object}	common.ErrorResponse
//	@Failure		409	{object}	common.ErrorResponse
//	@Failure		500	{object}	common.ErrorResponse
//	@Router			/sandboxes [post]
//
//	@id				Create
func Create(ctx *gin.Context) {
	var createSandboxDto dto.CreateSandboxDTO
	err := ctx.ShouldBindJSON(&createSandboxDto)
	if err != nil {
		ctx.Error(common.NewInvalidBodyRequestError(err))
		return
	}

	runner := runner.GetInstance(nil)

	containerId, err := runner.Docker.Create(ctx.Request.Context(), createSandboxDto)
	if err != nil {
		runner.Cache.SetSandboxState(ctx, createSandboxDto.Id, enums.SandboxStateError)
		common.ContainerOperationCount.WithLabelValues("create", string(common.PrometheusOperationStatusFailure)).Inc()
		ctx.Error(err)
		return
	}

	common.ContainerOperationCount.WithLabelValues("create", string(common.PrometheusOperationStatusSuccess)).Inc()

	ctx.JSON(http.StatusCreated, containerId)
}

// Destroy 			godoc
//
//	@Tags			sandbox
//	@Summary		Destroy sandbox
//	@Description	Destroy sandbox
//	@Produce		json
//	@Param			sandboxId	path		string	true	"Sandbox ID"
//	@Success		200			{string}	string	"Sandbox destroyed"
//	@Failure		400			{object}	common.ErrorResponse
//	@Failure		401			{object}	common.ErrorResponse
//	@Failure		404			{object}	common.ErrorResponse
//	@Failure		409			{object}	common.ErrorResponse
//	@Failure		500			{object}	common.ErrorResponse
//	@Router			/sandboxes/{sandboxId}/destroy [post]
//
//	@id				Destroy
func Destroy(ctx *gin.Context) {
	sandboxId := ctx.Param("sandboxId")

	runner := runner.GetInstance(nil)

	err := runner.Docker.Destroy(ctx.Request.Context(), sandboxId)
	if err != nil {
		runner.Cache.SetSandboxState(ctx, sandboxId, enums.SandboxStateError)
		common.ContainerOperationCount.WithLabelValues("destroy", string(common.PrometheusOperationStatusFailure)).Inc()
		ctx.Error(err)
		return
	}

	common.ContainerOperationCount.WithLabelValues("destroy", string(common.PrometheusOperationStatusSuccess)).Inc()

	ctx.JSON(http.StatusOK, "Sandbox destroyed")
}

// CreateBackup godoc
//
//	@Tags			sandbox
//	@Summary		Create sandbox backup
//	@Description	Create sandbox backup
//	@Produce		json
//	@Param			sandboxId	path		string				true	"Sandbox ID"
//	@Param			sandbox		body		dto.CreateBackupDTO	true	"Create backup"
//	@Success		201			{string}	string				"Backup created"
//	@Failure		400			{object}	common.ErrorResponse
//	@Failure		401			{object}	common.ErrorResponse
//	@Failure		404			{object}	common.ErrorResponse
//	@Failure		409			{object}	common.ErrorResponse
//	@Failure		500			{object}	common.ErrorResponse
//	@Router			/sandboxes/{sandboxId}/backup [post]
//
//	@id				CreateBackup
func CreateBackup(ctx *gin.Context) {
	sandboxId := ctx.Param("sandboxId")

	var createBackupDTO dto.CreateBackupDTO
	err := ctx.ShouldBindJSON(&createBackupDTO)
	if err != nil {
		ctx.Error(common.NewInvalidBodyRequestError(err))
		return
	}

	runner := runner.GetInstance(nil)

	err = runner.Docker.CreateBackup(ctx.Request.Context(), sandboxId, createBackupDTO)
	if err != nil {
		runner.Cache.SetBackupState(ctx, sandboxId, enums.BackupStateFailed)
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusCreated, "Backup created")
}

// Resize 			godoc
//
//	@Tags			sandbox
//	@Summary		Resize sandbox
//	@Description	Resize sandbox
//	@Produce		json
//	@Param			sandboxId	path		string					true	"Sandbox ID"
//	@Param			sandbox		body		dto.ResizeSandboxDTO	true	"Resize sandbox"
//	@Success		200			{string}	string					"Sandbox resized"
//	@Failure		400			{object}	common.ErrorResponse
//	@Failure		401			{object}	common.ErrorResponse
//	@Failure		404			{object}	common.ErrorResponse
//	@Failure		409			{object}	common.ErrorResponse
//	@Failure		500			{object}	common.ErrorResponse
//	@Router			/sandboxes/{sandboxId}/resize [post]
//
//	@id				Resize
func Resize(ctx *gin.Context) {
	var resizeDto dto.ResizeSandboxDTO
	err := ctx.ShouldBindJSON(&resizeDto)
	if err != nil {
		ctx.Error(common.NewInvalidBodyRequestError(err))
		return
	}

	sandboxId := ctx.Param("sandboxId")

	runner := runner.GetInstance(nil)

	err = runner.Docker.Resize(ctx.Request.Context(), sandboxId, resizeDto)
	if err != nil {
		runner.Cache.SetSandboxState(ctx, sandboxId, enums.SandboxStateError)
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, "Sandbox resized")
}

// Start 			godoc
//
//	@Tags			sandbox
//	@Summary		Start sandbox
//	@Description	Start sandbox
//	@Produce		json
//	@Param			sandboxId	path		string	true	"Sandbox ID"
//	@Success		200			{string}	string	"Sandbox started"
//	@Failure		400			{object}	common.ErrorResponse
//	@Failure		401			{object}	common.ErrorResponse
//	@Failure		404			{object}	common.ErrorResponse
//	@Failure		409			{object}	common.ErrorResponse
//	@Failure		500			{object}	common.ErrorResponse
//	@Router			/sandboxes/{sandboxId}/start [post]
//
//	@id				Start
func Start(ctx *gin.Context) {
	sandboxId := ctx.Param("sandboxId")

	runner := runner.GetInstance(nil)

	err := runner.Docker.Start(ctx.Request.Context(), sandboxId)
	if err != nil {
		runner.Cache.SetSandboxState(ctx, sandboxId, enums.SandboxStateError)
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, "Sandbox started")
}

// Stop 			godoc
//
//	@Tags			sandbox
//	@Summary		Stop sandbox
//	@Description	Stop sandbox
//	@Produce		json
//	@Param			sandboxId	path		string	true	"Sandbox ID"
//	@Success		200			{string}	string	"Sandbox stopped"
//	@Failure		400			{object}	common.ErrorResponse
//	@Failure		401			{object}	common.ErrorResponse
//	@Failure		404			{object}	common.ErrorResponse
//	@Failure		409			{object}	common.ErrorResponse
//	@Failure		500			{object}	common.ErrorResponse
//	@Router			/sandboxes/{sandboxId}/stop [post]
//
//	@id				Stop
func Stop(ctx *gin.Context) {
	sandboxId := ctx.Param("sandboxId")

	runner := runner.GetInstance(nil)

	err := runner.Docker.Stop(ctx.Request.Context(), sandboxId)
	if err != nil {
		runner.Cache.SetSandboxState(ctx, sandboxId, enums.SandboxStateError)
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, "Sandbox stopped")
}

// Info godoc
//
//	@Tags			sandbox
//	@Summary		Get sandbox info
//	@Description	Get sandbox info
//	@Produce		json
//	@Param			sandboxId	path		string				true	"Sandbox ID"
//	@Success		200			{object}	SandboxInfoResponse	"Sandbox info"
//	@Failure		400			{object}	common.ErrorResponse
//	@Failure		401			{object}	common.ErrorResponse
//	@Failure		404			{object}	common.ErrorResponse
//	@Failure		409			{object}	common.ErrorResponse
//	@Failure		500			{object}	common.ErrorResponse
//	@Router			/sandboxes/{sandboxId} [get]
//
//	@id				Info
func Info(ctx *gin.Context) {
	sandboxId := ctx.Param("sandboxId")

	runner := runner.GetInstance(nil)

	info := runner.SandboxService.GetSandboxStatesInfo(ctx.Request.Context(), sandboxId)

	ctx.JSON(http.StatusOK, SandboxInfoResponse{
		State:       info.SandboxState,
		BackupState: info.BackupState,
	})
}

type SandboxInfoResponse struct {
	State       enums.SandboxState `json:"state"`
	BackupState enums.BackupState  `json:"backupState"`
} //	@name	SandboxInfoResponse

// RemoveDestroyed godoc
//
//	@Tags			sandbox
//	@Summary		Remove a destroyed sandbox
//	@Description	Remove a sandbox that has been previously destroyed
//	@Produce		json
//	@Param			sandboxId	path		string	true	"Sandbox ID"
//	@Success		200			{string}	string	"Sandbox removed"
//	@Failure		400			{object}	common.ErrorResponse
//	@Failure		401			{object}	common.ErrorResponse
//	@Failure		404			{object}	common.ErrorResponse
//	@Failure		409			{object}	common.ErrorResponse
//	@Failure		500			{object}	common.ErrorResponse
//	@Router			/sandboxes/{sandboxId} [delete]
//
//	@id				RemoveDestroyed
func RemoveDestroyed(ctx *gin.Context) {
	sandboxId := ctx.Param("sandboxId")

	runner := runner.GetInstance(nil)

	err := runner.SandboxService.RemoveDestroyedSandbox(ctx.Request.Context(), sandboxId)
	if err != nil {
		if !common.IsNotFoundError(err) {
			ctx.Error(err)
			return
		}
	}

	ctx.JSON(http.StatusOK, "Sandbox removed")
}
