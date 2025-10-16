// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package controllers

import (
	"net/http"

	"github.com/daytonaio/runner/pkg/api/dto"
	"github.com/daytonaio/runner/pkg/runner"
	"github.com/gin-gonic/gin"
)

// ArchiveDisk 			godoc
//
//	@Summary		Archive disk
//	@Description	Archive disk to object storage
//	@Produce		json
//	@Tags			disk
//	@Param			diskId	path		string				true	"Disk ID"
//	@Param			request	body		dto.ArchiveDiskDTO	true	"Archive disk request"
//	@Success		200		{object}	map[string]interface{}
//	@Router			/disk/{diskId}/archive [post]
//
//	@id				ArchiveDisk
func ArchiveDisk(ctx *gin.Context) {
	diskId := ctx.Param("diskId")
	if diskId == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "diskId is required"})
		return
	}

	var request dto.ArchiveDiskDTO
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		"message": "Disk archived",
		"diskId":  diskId,
	})
}

// RestoreDisk 			godoc
//
//	@Summary		Restore disk
//	@Description	Restore disk from object storage
//	@Produce		json
//	@Tags			disk
//	@Param			diskId	path		string				true	"Disk ID"
//	@Param			request	body		dto.RestoreDiskDTO	true	"Restore disk request"
//	@Success		200		{object}	map[string]interface{}
//	@Router			/disk/{diskId}/restore [post]
//
//	@id				RestoreDisk
func RestoreDisk(ctx *gin.Context) {
	diskId := ctx.Param("diskId")
	if diskId == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "diskId is required"})
		return
	}

	var request dto.RestoreDiskDTO
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		"message": "Disk restored",
		"diskId":  diskId,
	})
}
