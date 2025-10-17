// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package controllers

import (
	"net/http"
	"time"

	"github.com/daytonaio/runner/pkg/api/dto"
	"github.com/daytonaio/runner/pkg/runner"
	"github.com/gin-gonic/gin"
)

// PushDisk 			godoc
//
//	@Summary		Push disk
//	@Description	Push disk to object storage
//	@Produce		json
//	@Tags			disk
//	@Param			diskId	path		string	true	"Disk ID"
//	@Success		200		{object}	map[string]interface{}
//	@Router			/disk/{diskId}/push [post]
//
//	@id				PushDisk
func PushDisk(ctx *gin.Context) {
	diskId := ctx.Param("diskId")
	if diskId == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "diskId is required"})
		return
	}

	runner := runner.GetInstance(nil)

	disk, err := (*runner.SDisk).Open(ctx.Request.Context(), diskId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if disk.IsMounted() {
		err = disk.Unmount(ctx.Request.Context())
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	err = disk.Push(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Disk pushed",
		"diskId":  diskId,
	})
}

// PullDisk 			godoc
//
//	@Summary		Pull disk
//	@Description	Pull disk from object storage
//	@Produce		json
//	@Tags			disk
//	@Param			diskId	path		string	true	"Disk ID"
//	@Success		200		{object}	map[string]interface{}
//	@Router			/disk/{diskId}/pull [post]
//
//	@id				PullDisk
func PullDisk(ctx *gin.Context) {
	diskId := ctx.Param("diskId")
	if diskId == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "diskId is required"})
		return
	}

	runner := runner.GetInstance(nil)

	disks, err := (*runner.SDisk).List(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	for _, disk := range disks {
		if disk.Name == diskId {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Disk already exists on runner"})
			return
		}
	}

	_, err = (*runner.SDisk).Pull(ctx.Request.Context(), diskId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Disk pulled",
		"diskId":  diskId,
	})
}

// DiskInfo 			godoc
//
//	@Summary		Get disk info
//	@Description	Get detailed information about a specific disk
//	@Produce		json
//	@Tags			disk
//	@Param			diskId	path		string	true	"Disk ID"
//	@Success		200		{object}	dto.DiskInfoDTO
//	@Failure		400		{object}	map[string]interface{}
//	@Failure		404		{object}	map[string]interface{}
//	@Failure		500		{object}	map[string]interface{}
//	@Router			/disk/{diskId}/info [get]
//
//	@id				DiskInfo
func DiskInfo(ctx *gin.Context) {
	diskId := ctx.Param("diskId")
	if diskId == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "diskId is required"})
		return
	}

	runner := runner.GetInstance(nil)

	disk, err := (*runner.SDisk).Open(ctx.Request.Context(), diskId)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Disk not found"})
		return
	}
	defer disk.Close()

	info := disk.Info()

	response := dto.DiskInfoDTO{
		Name:         info.Name,
		SizeGB:       info.SizeGB,
		ActualSizeGB: info.ActualSizeGB,
		Created:      info.Created.Format(time.RFC3339),
		Modified:     info.Modified.Format(time.RFC3339),
		IsMounted:    info.IsMounted,
		MountPath:    disk.MountPath(),
		InS3:         info.InS3,
		Checksum:     info.Checksum,
	}

	ctx.JSON(http.StatusOK, response)
}
