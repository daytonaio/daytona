// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package fs

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strconv"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
	"github.com/daytonaio/daemon/pkg/common"
	"github.com/gin-gonic/gin"
)

// SetFilePermissions godoc
//
//	@Summary		Set file permissions
//	@Description	Set file permissions, ownership, and group for a file or directory
//	@Tags			file-system
//	@Param			path	query	string	true	"File or directory path"
//	@Param			owner	query	string	false	"Owner (username or UID)"
//	@Param			group	query	string	false	"Group (group name or GID)"
//	@Param			mode	query	string	false	"File mode in octal format (e.g., 0755)"
//	@Success		200
//	@Failure		400	{object}	common.ErrorResponse
//	@Failure		403	{object}	common.ErrorResponse
//	@Failure		404	{object}	common.ErrorResponse
//	@Router			/files/permissions [post]
//
//	@id				SetFilePermissions
func SetFilePermissions(c *gin.Context) {
	path := c.Query("path")
	ownerParam := c.Query("owner")
	groupParam := c.Query("group")
	mode := c.Query("mode")

	if path == "" {
		c.Error(common_errors.NewBadRequestError(errors.New("path is required")))
		return
	}

	// convert to absolute path and check existence
	absPath, err := filepath.Abs(path)
	if err != nil {
		c.Error(common_errors.NewBadRequestError(errors.New("invalid path")))
		return
	}

	_, err = os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			c.Error(common.NewFileNotFoundError(err.Error()))
			return
		}
		if os.IsPermission(err) {
			c.Error(common.NewFileAccessDeniedError(err.Error()))
			return
		}
		c.Error(common_errors.NewBadRequestError(err))
		return
	}

	// handle mode change
	if mode != "" {
		modeNum, err := strconv.ParseUint(mode, 8, 32)
		if err != nil {
			c.Error(common_errors.NewBadRequestError(errors.New("invalid mode format")))
			return
		}

		if err := os.Chmod(absPath, os.FileMode(modeNum)); err != nil {
			c.Error(classifyChmodChownError(absPath, "failed to change mode", err))
			return
		}
	}

	// handle ownership change
	if ownerParam != "" || groupParam != "" {
		uid := -1
		gid := -1

		// resolve owner
		if ownerParam != "" {
			// first try as numeric UID
			if uidNum, err := strconv.Atoi(ownerParam); err == nil {
				uid = uidNum
			} else {
				// try as username
				if u, err := user.Lookup(ownerParam); err == nil {
					if uid, err = strconv.Atoi(u.Uid); err != nil {
						c.Error(common_errors.NewBadRequestError(errors.New("invalid user ID")))
						return
					}
				} else {
					c.Error(common_errors.NewBadRequestError(errors.New("user not found")))
					return
				}
			}
		}

		// resolve group
		if groupParam != "" {
			// first try as numeric GID
			if gidNum, err := strconv.Atoi(groupParam); err == nil {
				gid = gidNum
			} else {
				// try as group name
				if g, err := user.LookupGroup(groupParam); err == nil {
					if gid, err = strconv.Atoi(g.Gid); err != nil {
						c.Error(common_errors.NewBadRequestError(errors.New("invalid group ID")))
						return
					}
				} else {
					c.Error(common_errors.NewBadRequestError(errors.New("group not found")))
					return
				}
			}
		}

		if err := os.Chown(absPath, uid, gid); err != nil {
			c.Error(classifyChmodChownError(absPath, "failed to change ownership", err))
			return
		}
	}

	c.Status(http.StatusOK)
}

// classifyChmodChownError preserves typed daemon error codes for permission
// metadata mutations: not-found and access-denied conditions surface as
// FILE_NOT_FOUND / FILE_ACCESS_DENIED rather than a generic 400.
func classifyChmodChownError(path, action string, err error) error {
	wrapped := fmt.Errorf("%s for %s: %w", action, path, err)
	if os.IsNotExist(err) {
		return common.NewFileNotFoundError(wrapped.Error())
	}
	if os.IsPermission(err) {
		return common.NewFileAccessDeniedError(wrapped.Error())
	}
	return common_errors.NewBadRequestError(wrapped)
}
