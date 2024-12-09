// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package fs

import (
	"fmt"
	"log"
	"net/http"
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
		log.Printf("Error: empty path parameter received")
		c.JSON(http.StatusBadRequest, gin.H{"error": "path is required"})
		return
	}

	// convert to absolute path and check existence
	absPath, err := filepath.Abs(path)
	if err != nil {
		log.Printf("Error: failed to resolve absolute path for %s: %v", path, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid path"})
		return
	}

	_, err = os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("Error: file not found at path %s", absPath)
			c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
			return
		}
		log.Printf("Error: failed to stat file %s: %v", absPath, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to access file"})
		return
	}

	// handle mode change
	if mode != "" {
		modeNum, err := strconv.ParseUint(mode, 8, 32)
		if err != nil {
			log.Printf("Error: invalid mode format %s: %v", mode, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid mode format"})
			return
		}

		log.Printf("Changing mode of %s to %s", absPath, mode)
		if err := os.Chmod(absPath, os.FileMode(modeNum)); err != nil {
			log.Printf("Error: failed to change mode for %s to %s: %v", absPath, mode, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to change mode: %v", err)})
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
						log.Printf("Error: failed to convert UID for user %s: %v", ownerParam, err)
						c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
						return
					}
				} else {
					log.Printf("Error: user %s not found: %v", ownerParam, err)
					c.JSON(http.StatusBadRequest, gin.H{"error": "user not found"})
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
						log.Printf("Error: failed to convert GID for group %s: %v", groupParam, err)
						c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group ID"})
						return
					}
				} else {
					log.Printf("Error: group %s not found: %v", groupParam, err)
					c.JSON(http.StatusBadRequest, gin.H{"error": "group not found"})
					return
				}
			}
		}

		log.Printf("Changing ownership of %s to uid=%d,gid=%d", absPath, uid, gid)
		if err := os.Chown(absPath, uid, gid); err != nil {
			log.Printf("Error: failed to change ownership for %s to uid=%d,gid=%d: %v", absPath, uid, gid, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to change ownership: %v", err)})
			return
		}
	}

	// verify final permissions
	finalInfo, err := os.Stat(absPath)
	if err == nil {
		log.Printf("Final permissions for %s: mode=%v", absPath, finalInfo.Mode())
	}

	c.JSON(http.StatusOK, gin.H{"message": "permissions updated successfully"})
}
