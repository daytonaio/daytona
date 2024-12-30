// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package toolbox

import "github.com/gin-gonic/gin"

// FsCreateFolder 			godoc
//
//	@Tags			workspace toolbox
//	@Summary		Create folder
//	@Description	Create folder inside workspace project
//	@Produce		json
//	@Param			workspaceId	path	string	true	"Workspace ID or Name"
//	@Param			projectId	path	string	true	"Project ID"
//	@Param			path		query	string	true	"Path"
//	@Param			mode		query	string	true	"Mode"
//	@Success		201
//	@Router			/workspace/{workspaceId}/{projectId}/toolbox/files/folder [post]
//
//	@id				FsCreateFolder
func FsCreateFolder(ctx *gin.Context) {
	forwardRequestToToolbox(ctx)
}

// FsDeleteFile 			godoc
//
//	@Tags			workspace toolbox
//	@Summary		Delete file
//	@Description	Delete file inside workspace project
//	@Produce		json
//	@Param			workspaceId	path	string	true	"Workspace ID or Name"
//	@Param			projectId	path	string	true	"Project ID"
//	@Param			path		query	string	true	"Path"
//	@Success		204
//	@Router			/workspace/{workspaceId}/{projectId}/toolbox/files [delete]
//
//	@id				FsDeleteFile
func FsDeleteFile(ctx *gin.Context) {
	forwardRequestToToolbox(ctx)
}

// FsDownloadFile 			godoc
//
//	@Tags			workspace toolbox
//	@Summary		Download file
//	@Description	Download file from workspace project
//	@Produce		json
//	@Param			workspaceId	path	string	true	"Workspace ID or Name"
//	@Param			projectId	path	string	true	"Project ID"
//	@Param			path		query	string	true	"Path"
//	@Success		200			{file}	file	"response contains the file"
//	@Router			/workspace/{workspaceId}/{projectId}/toolbox/files/download [get]
//
//	@id				FsDownloadFile
func FsDownloadFile(ctx *gin.Context) {
	forwardRequestToToolbox(ctx)
}

// FsFindInFiles			godoc
//
//	@Tags			workspace toolbox
//	@Summary		Search for text/pattern in files
//	@Description	Search for text/pattern inside workspace project files
//	@Produce		json
//	@Param			workspaceId	path	string	true	"Workspace ID or Name"
//	@Param			projectId	path	string	true	"Project ID"
//	@Param			path		query	string	true	"Path"
//	@Param			pattern		query	string	true	"Pattern"
//	@Success		200			{array}	Match
//	@Router			/workspace/{workspaceId}/{projectId}/toolbox/files/find [get]
//
//	@id				FsFindInFiles
func FsFindInFiles(ctx *gin.Context) {
	forwardRequestToToolbox(ctx)
}

// FsGetFileDetails			godoc
//
//	@Tags			workspace toolbox
//	@Summary		Get file info
//	@Description	Get file info inside workspace project
//	@Produce		json
//	@Param			workspaceId	path		string	true	"Workspace ID or Name"
//	@Param			projectId	path		string	true	"Project ID"
//	@Param			path		query		string	true	"Path"
//	@Success		200			{object}	FileInfo
//	@Router			/workspace/{workspaceId}/{projectId}/toolbox/files/info [get]
//
//	@id				FsGetFileDetails
func FsGetFileDetails(ctx *gin.Context) {
	forwardRequestToToolbox(ctx)
}

// FsListFiles 			godoc
//
//	@Tags			workspace toolbox
//	@Summary		List files
//	@Description	List files inside workspace project
//	@Produce		json
//	@Param			workspaceId	path	string	true	"Workspace ID or Name"
//	@Param			projectId	path	string	true	"Project ID"
//	@Param			path		query	string	false	"Path"
//	@Success		200			{array}	FileInfo
//	@Router			/workspace/{workspaceId}/{projectId}/toolbox/files [get]
//
//	@id				FsListFiles
func FsListFiles(ctx *gin.Context) {
	forwardRequestToToolbox(ctx)
}

// FsMoveFile 			godoc
//
//	@Tags			workspace toolbox
//	@Summary		Create folder
//	@Description	Create folder inside workspace project
//	@Produce		json
//	@Param			workspaceId	path	string	true	"Workspace ID or Name"
//	@Param			projectId	path	string	true	"Project ID"
//	@Param			source		query	string	true	"Source path"
//	@Param			destination	query	string	true	"Destination path"
//	@Success		200
//	@Router			/workspace/{workspaceId}/{projectId}/toolbox/files/move [post]
//
//	@id				FsMoveFile
func FsMoveFile(ctx *gin.Context) {
	forwardRequestToToolbox(ctx)
}

// FsReplaceInFiles			godoc
//
//	@Tags			workspace toolbox
//	@Summary		Repleace text/pattern in files
//	@Description	Repleace text/pattern in mutilple files inside workspace project
//	@Produce		json
//	@Param			workspaceId	path	string			true	"Workspace ID or Name"
//	@Param			projectId	path	string			true	"Project ID"
//	@Param			replace		body	ReplaceRequest	true	"ReplaceParams"
//	@Success		200			{array}	ReplaceResult
//	@Router			/workspace/{workspaceId}/{projectId}/toolbox/files/replace [post]
//
//	@id				FsReplaceInFiles
func FsReplaceInFiles(ctx *gin.Context) {
	forwardRequestToToolbox(ctx)
}

// FsSearchFiles 			godoc
//
//	@Tags			workspace toolbox
//	@Summary		Search for files
//	@Description	Search for files inside workspace project
//	@Produce		json
//	@Param			workspaceId	path		string	true	"Workspace ID or Name"
//	@Param			projectId	path		string	true	"Project ID"
//	@Param			path		query		string	true	"Path"
//	@Param			pattern		query		string	true	"Pattern"
//	@Success		200			{object}	SearchFilesResponse
//	@Router			/workspace/{workspaceId}/{projectId}/toolbox/files/search [get]
//
//	@id				FsSearchFiles
func FsSearchFiles(ctx *gin.Context) {
	forwardRequestToToolbox(ctx)
}

// FsSetFilePermissions			godoc
//
//	@Tags			workspace toolbox
//	@Summary		Set file owner/group/permissions
//	@Description	Set file owner/group/permissions inside workspace project
//	@Produce		json
//	@Param			workspaceId	path	string	true	"Workspace ID or Name"
//	@Param			projectId	path	string	true	"Project ID"
//	@Param			path		query	string	true	"Path"
//	@Param			owner		query	string	false	"Owner"
//	@Param			group		query	string	false	"Group"
//	@Param			mode		query	string	false	"Mode"
//	@Success		200
//	@Router			/workspace/{workspaceId}/{projectId}/toolbox/files/permissions [post]
//
//	@id				FsSetFilePermissions
func FsSetFilePermissions(ctx *gin.Context) {
	forwardRequestToToolbox(ctx)
}

// FsUploadFile 			godoc
//
//	@Tags			workspace toolbox
//	@Summary		Upload file
//	@Description	Upload file inside workspace project
//	@Produce		json
//	@Param			workspaceId	path		string	true	"Workspace ID or Name"
//	@Param			projectId	path		string	true	"Project ID"
//	@Param			path		query		string	true	"Path"
//	@Param			file		formData	file	true	"File"
//	@Success		200
//	@Router			/workspace/{workspaceId}/{projectId}/toolbox/files/upload [post]
//
//	@id				FsUploadFile
func FsUploadFile(ctx *gin.Context) {
	forwardRequestToToolbox(ctx)
}
