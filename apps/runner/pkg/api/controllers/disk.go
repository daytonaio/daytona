// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package controllers

import (
	"fmt"
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

// DeleteDisk 			godoc
//
//	@Summary		Delete disk
//	@Description	Delete a disk from the runner
//	@Produce		json
//	@Tags			disk
//	@Param			diskId	path		string	true	"Disk ID"
//	@Param			force	query		bool	false	"Force delete mounted disk (unmounts first)"
//	@Success		200		{object}	map[string]interface{}
//	@Failure		400		{object}	map[string]interface{}
//	@Failure		404		{object}	map[string]interface{}
//	@Failure		409		{object}	map[string]interface{}
//	@Failure		500		{object}	map[string]interface{}
//	@Router			/disk/{diskId}/delete [delete]
//
//	@id				DeleteDisk
func DeleteDisk(ctx *gin.Context) {
	diskId := ctx.Param("diskId")
	if diskId == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "diskId is required"})
		return
	}

	// Check for force parameter
	force := ctx.Query("force") == "true"

	runner := runner.GetInstance(nil)

	// Check if disk exists
	disk, err := (*runner.SDisk).Open(ctx.Request.Context(), diskId)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Disk not found"})
		return
	}
	defer disk.Close()

	// Check if disk is mounted
	if disk.IsMounted() {
		if !force {
			ctx.JSON(http.StatusConflict, gin.H{"error": "Cannot delete mounted disk. Please unmount first or use ?force=true"})
			return
		}

		// Force unmount the disk
		if err := disk.Unmount(ctx.Request.Context()); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to unmount disk: %v", err)})
			return
		}
	}

	// Delete the disk
	err = (*runner.SDisk).Delete(ctx.Request.Context(), diskId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Disk deleted",
		"diskId":  diskId,
	})
}

// ForkDisk 			godoc
//
//	@Summary		Fork disk
//	@Description	Create a new disk that shares all existing layers of the source disk. Both disks will have independent write layers.
//	@Produce		json
//	@Tags			disk
//	@Param			diskId	path		string			true	"Source Disk ID"
//	@Param			request	body		dto.ForkDiskDTO	true	"Fork disk request"
//	@Success		200		{object}	map[string]interface{}
//	@Failure		400		{object}	map[string]interface{}
//	@Failure		404		{object}	map[string]interface{}
//	@Failure		409		{object}	map[string]interface{}
//	@Failure		500		{object}	map[string]interface{}
//	@Router			/disk/fork/{diskId} [post]
//
//	@id				ForkDisk
func ForkDisk(ctx *gin.Context) {
	sourceDiskId := ctx.Param("diskId")
	if sourceDiskId == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "diskId is required"})
		return
	}

	var forkDiskDto dto.ForkDiskDTO
	if err := ctx.ShouldBindJSON(&forkDiskDto); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	newDiskId := forkDiskDto.NewDiskId
	if newDiskId == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "newDiskId is required"})
		return
	}

	runner := runner.GetInstance(nil)

	// Validate source disk exists and is not mounted
	sourceDisk, err := (*runner.SDisk).Open(ctx.Request.Context(), sourceDiskId)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Source disk not found"})
		return
	}
	defer sourceDisk.Close()

	if sourceDisk.IsMounted() {
		ctx.JSON(http.StatusConflict, gin.H{"error": "Cannot fork mounted disk. Please unmount first"})
		return
	}

	// Check if new disk already exists
	disks, err := (*runner.SDisk).List(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	for _, disk := range disks {
		if disk.Name == newDiskId {
			ctx.JSON(http.StatusConflict, gin.H{"error": "New disk already exists"})
			return
		}
	}

	// Fork the disk
	newDisk, err := (*runner.SDisk).Fork(ctx.Request.Context(), sourceDiskId, newDiskId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer newDisk.Close()

	ctx.JSON(http.StatusOK, gin.H{
		"message":      "Disk forked successfully",
		"sourceDiskId": sourceDiskId,
		"newDiskId":    newDiskId,
	})
}
