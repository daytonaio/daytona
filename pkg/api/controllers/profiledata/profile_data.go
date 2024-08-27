// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package profiledata

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/pkg/profiledata"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/gin-gonic/gin"
)

// GetProfileData godoc
//
//	@Tags			profile
//	@Summary		Get profile data
//	@Description	Get profile data
//	@Accept			json
//	@Success		200 {object} profiledata.ProfileData
//	@Router			/profile [get]
//
//	@id				GetProfileData
func GetProfileData(ctx *gin.Context) {
	server := server.GetInstance(nil)
	profileData, err := server.ProfileDataService.Get()
	if err != nil {
		if profiledata.IsProfileDataNotFound(err) {
			ctx.JSON(200, &profiledata.ProfileData{})
			return
		}
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get profile data: %s", err.Error()))
		return
	}

	ctx.JSON(200, profileData)
}

// SetProfileData godoc
//
//	@Tags			profile
//	@Summary		Set profile data
//	@Description	Set profile data
//	@Accept			json
//	@Param			profileData	body	profiledata.ProfileData	true	"Profile data"
//	@Success		201
//	@Router			/profile [put]
//
//	@id				SetProfileData
func SetProfileData(ctx *gin.Context) {
	var req profiledata.ProfileData
	err := ctx.BindJSON(&req)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %s", err.Error()))
		return
	}

	server := server.GetInstance(nil)
	err = server.ProfileDataService.Save(&req)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to save profile data: %s", err.Error()))
		return
	}

	ctx.Status(201)
}

// DeleteProfileData godoc
//
//	@Tags			profile
//	@Summary		Delete profile data
//	@Description	Delete profile data
//	@Success		204
//	@Router			/profile [delete]
//
//	@id				DeleteProfileData
func DeleteProfileData(ctx *gin.Context) {
	server := server.GetInstance(nil)
	err := server.ProfileDataService.Delete()
	if err != nil {
		if profiledata.IsProfileDataNotFound(err) {
			ctx.Status(204)
			return
		}
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get profile data: %s", err.Error()))
		return
	}

	ctx.Status(204)
}
