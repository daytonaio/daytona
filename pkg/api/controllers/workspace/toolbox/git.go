// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package toolbox

import "github.com/gin-gonic/gin"

// GitAddFiles			godoc
//
//	@Tags			workspace toolbox
//	@Summary		Add files
//	@Description	Add files to git commit
//	@Produce		json
//	@Param			workspaceId	path	string			true	"Workspace ID or Name"
//	@Param			projectId	path	string			true	"Project ID"
//	@Param			params		body	GitAddRequest	true	"GitAddRequest"
//	@Success		200
//	@Router			/workspace/{workspaceId}/{projectId}/toolbox/git/add [post]
//
//	@id				GitAddFiles
func GitAddFiles(ctx *gin.Context) {
	forwardRequestToToolbox(ctx)
}

// GitCloneRepository			godoc
//
//	@Tags			workspace toolbox
//	@Summary		Clone git repository
//	@Description	Clone git repository inside workspace project
//	@Produce		json
//	@Param			workspaceId	path	string			true	"Workspace ID or Name"
//	@Param			projectId	path	string			true	"Project ID"
//	@Param			params		body	GitCloneRequest	true	"GitCloneRequest"
//	@Success		200
//	@Router			/workspace/{workspaceId}/{projectId}/toolbox/git/clone [post]
//
//	@id				GitCloneRepository
func GitCloneRepository(ctx *gin.Context) {
	forwardRequestToToolbox(ctx)
}

// GitCommitChanges			godoc
//
//	@Tags			workspace toolbox
//	@Summary		Commit changes
//	@Description	Commit changes to git repository inside workspace project
//	@Produce		json
//	@Param			workspaceId	path		string				true	"Workspace ID or Name"
//	@Param			projectId	path		string				true	"Project ID"
//	@Param			params		body		GitCommitRequest	true	"GitCommitRequest"
//	@Success		200			{object}	GitCommitResponse
//	@Router			/workspace/{workspaceId}/{projectId}/toolbox/git/commit [post]
//
//	@id				GitCommitChanges
func GitCommitChanges(ctx *gin.Context) {
	forwardRequestToToolbox(ctx)
}

// GitCreateBranch			godoc
//
//	@Tags			workspace toolbox
//	@Summary		Create branch
//	@Description	Create branch on git repository inside workspace project
//	@Produce		json
//	@Param			workspaceId	path	string				true	"Workspace ID or Name"
//	@Param			projectId	path	string				true	"Project ID"
//	@Param			params		body	GitBranchRequest	true	"GitBranchRequest"
//	@Success		201
//	@Router			/workspace/{workspaceId}/{projectId}/toolbox/git/branches [post]
//
//	@id				GitCreateBranch
func GitCreateBranch(ctx *gin.Context) {
	forwardRequestToToolbox(ctx)
}

// GitCommitHistory			godoc
//
//	@Tags			workspace toolbox
//	@Summary		Get commit history
//	@Description	Get commit history from git repository inside workspace project
//	@Produce		json
//	@Param			workspaceId	path	string	true	"Workspace ID or Name"
//	@Param			projectId	path	string	true	"Project ID"
//	@Param			path		query	string	true	"Path to git repository"
//	@Success		200			{array}	GitCommitInfo
//	@Router			/workspace/{workspaceId}/{projectId}/toolbox/git/history [get]
//
//	@id				GitCommitHistory
func GitCommitHistory(ctx *gin.Context) {
	forwardRequestToToolbox(ctx)
}

// GitBranchList			godoc
//
//	@Tags			workspace toolbox
//	@Summary		Get branch list
//	@Description	Get branch list from git repository inside workspace project
//	@Produce		json
//	@Param			workspaceId	path		string	true	"Workspace ID or Name"
//	@Param			projectId	path		string	true	"Project ID"
//	@Param			path		query		string	true	"Path to git repository"
//	@Success		200			{object}	ListBranchResponse
//	@Router			/workspace/{workspaceId}/{projectId}/toolbox/git/branches [get]
//
//	@id				GitBranchList
func GitBranchList(ctx *gin.Context) {
	forwardRequestToToolbox(ctx)
}

// GitPullChanges			godoc
//
//	@Tags			workspace toolbox
//	@Summary		Pull changes
//	@Description	Pull changes from remote to git repository inside workspace project
//	@Produce		json
//	@Param			workspaceId	path	string			true	"Workspace ID or Name"
//	@Param			projectId	path	string			true	"Project ID"
//	@Param			params		body	GitRepoRequest	true	"Git pull request"
//	@Success		200
//	@Router			/workspace/{workspaceId}/{projectId}/toolbox/git/pull [post]
//
//	@id				GitPullChanges
func GitPullChanges(ctx *gin.Context) {
	forwardRequestToToolbox(ctx)
}

// GitPushChanges			godoc
//
//	@Tags			workspace toolbox
//	@Summary		Push changes
//	@Description	Push changes to remote from git repository inside workspace project
//	@Produce		json
//	@Param			workspaceId	path	string			true	"Workspace ID or Name"
//	@Param			projectId	path	string			true	"Project ID"
//	@Param			params		body	GitRepoRequest	true	"Git push request"
//	@Success		200
//	@Router			/workspace/{workspaceId}/{projectId}/toolbox/git/push [post]
//
//	@id				GitPushChanges
func GitPushChanges(ctx *gin.Context) {
	forwardRequestToToolbox(ctx)
}

// GitStatus			godoc
//
//	@Tags			workspace toolbox
//	@Summary		Get git status
//	@Description	Get status from git repository inside workspace project
//	@Produce		json
//	@Param			workspaceId	path		string	true	"Workspace ID or Name"
//	@Param			projectId	path		string	true	"Project ID"
//	@Param			path		query		string	true	"Path to git repository"
//	@Success		200			{object}	GitStatus
//	@Router			/workspace/{workspaceId}/{projectId}/toolbox/git/status [get]
//
//	@id				GitGitStatus
func GitStatus(ctx *gin.Context) {
	forwardRequestToToolbox(ctx)
}
