// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package fs

import (
	"errors"
	iofs "io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/daytonaio/daemon/internal/util"
	"github.com/gin-gonic/gin"
)

// ListFiles godoc
//
//	@Summary		List files and directories
//	@Description	List files and directories in the specified path. Use the optional depth
//	@Description	parameter to list recursively: depth=1 (default) lists the directory's
//	@Description	entries, depth=2 also includes their children, and so on.
//	@Tags			file-system
//	@Produce		json
//	@Param			path	query	string	false	"Directory path to list (defaults to working directory)"
//	@Param			depth	query	int		false	"How many levels deep to list (default: 1, must be >= 1)"
//	@Success		200		{array}	FileInfo
//	@Router			/files [get]
//
//	@id				ListFiles
func ListFiles(c *gin.Context) {
	root := c.Query("path")
	if root == "" {
		root = "."
	}

	depth := 1
	if depthStr := c.Query("depth"); depthStr != "" {
		parsed, err := strconv.Atoi(depthStr)
		if err != nil || parsed < 1 {
			c.AbortWithError(http.StatusBadRequest, errors.New("depth must be an integer >= 1"))
			return
		}
		depth = parsed
	}

	stripPath := util.ClientRejectsUnknownResponseFields(c.Request.Header)

	// depth=1 uses the original os.ReadDir code path to avoid behavioural
	// regressions for the default (most common) case.
	if depth == 1 {
		listFilesShallow(c, root, stripPath)
		return
	}
	listFilesRecursive(c, root, depth, stripPath)
}

func listFilesShallow(c *gin.Context, path string, stripPath bool) {
	files, err := os.ReadDir(path)
	if err != nil {
		abortWithFsError(c, err)
		return
	}

	fileInfos := make([]FileInfo, 0)
	for _, file := range files {
		fullPath := filepath.Join(path, file.Name())
		info, err := getFileInfo(fullPath)
		if err != nil {
			continue
		}
		if !stripPath {
			info.Path = fullPath
		}
		fileInfos = append(fileInfos, info)
	}

	c.JSON(http.StatusOK, fileInfos)
}

// listFilesRecursive returns a flat listing up to depth levels below root;
// unreadable subtrees are skipped and symlinks are not followed.
func listFilesRecursive(c *gin.Context, root string, depth int, stripPath bool) {
	if _, err := os.ReadDir(root); err != nil {
		abortWithFsError(c, err)
		return
	}

	// filepath.WalkDir lstats its root and refuses to descend into a symlinked
	// root directory, while the depth=1 os.ReadDir path follows it. Resolve a
	// symlinked root so both depths list the same tree; reported paths stay
	// under the requested root.
	walkRoot := root
	if info, err := os.Lstat(root); err == nil && info.Mode()&os.ModeSymlink != 0 {
		if resolved, err := filepath.EvalSymlinks(root); err == nil {
			walkRoot = resolved
		}
	}

	fileInfos := make([]FileInfo, 0)
	_ = filepath.WalkDir(walkRoot, func(entryPath string, d iofs.DirEntry, err error) error {
		if err != nil {
			if d != nil && d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entryPath == walkRoot {
			return nil
		}

		rel, relErr := filepath.Rel(walkRoot, entryPath)
		if relErr != nil {
			return nil
		}
		entryDepth := strings.Count(rel, string(os.PathSeparator)) + 1
		if entryDepth > depth {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if info, infoErr := getFileInfo(entryPath); infoErr == nil {
			if !stripPath {
				info.Path = filepath.Join(root, rel)
			}
			fileInfos = append(fileInfos, info)
		}

		if d.IsDir() && entryDepth >= depth {
			return filepath.SkipDir
		}
		return nil
	})

	c.JSON(http.StatusOK, fileInfos)
}

func abortWithFsError(c *gin.Context, err error) {
	if os.IsNotExist(err) {
		c.AbortWithError(http.StatusNotFound, err)
		return
	}
	if os.IsPermission(err) {
		c.AbortWithError(http.StatusForbidden, err)
		return
	}
	c.AbortWithError(http.StatusBadRequest, err)
}
