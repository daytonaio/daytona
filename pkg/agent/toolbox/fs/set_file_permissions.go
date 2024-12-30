// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package fs

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
)

func SetFilePermissions(c *gin.Context) {
	path := c.Query("path")
	ownerParam := c.Query("owner")
	groupParam := c.Query("group")
	mode := c.Query("mode")

	if path == "" {
		c.AbortWithError(400, errors.New("path is required"))
		return
	}

	// convert to absolute path and check existence
	absPath, err := filepath.Abs(path)
	if err != nil {
		c.AbortWithError(400, errors.New("invalid path"))
		return
	}

	_, err = os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			c.AbortWithError(404, errors.New("file not found"))
			return
		}
		c.AbortWithError(400, errors.New("unable to access file"))
		return
	}

	// handle mode change
	if mode != "" {
		modeNum, err := strconv.ParseUint(mode, 8, 32)
		if err != nil {
			c.AbortWithError(400, errors.New("invalid mode format"))
			return
		}

		if err := os.Chmod(absPath, os.FileMode(modeNum)); err != nil {
			c.AbortWithError(400, fmt.Errorf("failed to change mode: %w", err))
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
						c.AbortWithError(400, errors.New("invalid user ID"))
						return
					}
				} else {
					c.AbortWithError(400, errors.New("user not found"))
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
						c.AbortWithError(400, errors.New("invalid group ID"))
						return
					}
				} else {
					c.AbortWithError(400, errors.New("group not found"))
					return
				}
			}
		}

		if err := os.Chown(absPath, uid, gid); err != nil {
			c.AbortWithError(400, fmt.Errorf("failed to change ownership: %w", err))
			return
		}
	}

	c.Status(200)
}
