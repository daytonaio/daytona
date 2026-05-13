// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package git

import (
	"fmt"
	"net/http"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
	"github.com/daytonaio/daemon/pkg/git"
	"github.com/daytonaio/daemon/pkg/gitprovider"
	"github.com/gin-gonic/gin"
	go_git_http "github.com/go-git/go-git/v5/plumbing/transport/http"
)

// CloneRepository godoc
//
//	@Summary		Clone a Git repository
//	@Description	Clone a Git repository to the specified path
//	@Tags			git
//	@Accept			json
//	@Produce		json
//	@Param			request	body		GitCloneRequest	true	"Clone repository request"
//	@Success		200		{object}	GitCloneResponse
//	@Success		202		{object}	GitCloneResponse
//	@Router			/git/clone [post]
//
//	@id				CloneRepository
func CloneRepository(c *gin.Context) {
	var req GitCloneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(common_errors.NewInvalidBodyRequestError(fmt.Errorf("invalid request body: %w", err)))
		return
	}

	branch := ""
	if req.Branch != nil {
		branch = *req.Branch
	}

	repo := gitprovider.GitRepository{
		Url:               req.URL,
		Branch:            branch,
		Depth:             req.Depth,
		SingleBranch:      req.SingleBranch,
		NoTags:            req.NoTags,
		Sparse:            req.Sparse,
		SparsePaths:       req.SparsePaths,
		Dissociate:        req.Dissociate,
		RecurseSubmodules: req.RecurseSubmodules,
		ShallowSubmodules: req.ShallowSubmodules,
		FilterSubmodules:  req.FilterSubmodules,
		NoCheckout:        req.NoCheckout,

		BackgroundExpansion:    req.BackgroundExpansion,
		InitialSparsePaths:     req.InitialSparsePaths,
		BackgroundDeepen:       req.BackgroundDeepen,
		BackgroundUnshallow:    req.BackgroundUnshallow,
		BackgroundHydratePaths: req.BackgroundHydratePaths,
	}

	if req.ShallowSince != nil {
		repo.ShallowSince = *req.ShallowSince
	}
	if req.Filter != nil {
		repo.Filter = *req.Filter
	}
	if req.ReferencePath != nil {
		repo.ReferencePath = *req.ReferencePath
	}
	if req.BackgroundExpansion != nil && *req.BackgroundExpansion {
		if len(req.InitialSparsePaths) > 0 {
			enabled := true
			repo.Sparse = &enabled
			repo.SparsePaths = req.InitialSparsePaths
		} else if req.NoCheckout == nil {
			enabled := true
			repo.NoCheckout = &enabled
		}
	}

	if req.CommitID != nil {
		repo.Target = gitprovider.CloneTargetCommit
		repo.Sha = *req.CommitID
	}

	gitService := git.Service{
		WorkDir: req.Path,
	}

	var auth *go_git_http.BasicAuth

	// Set authentication if provided
	if req.Username != nil && req.Password != nil {
		auth = &go_git_http.BasicAuth{
			Username: *req.Username,
			Password: *req.Password,
		}
	}

	err := gitService.CloneRepository(&repo, auth)
	if err != nil {
		abortWithGitError(c, err)
		return
	}

	if req.BackgroundExpansion != nil && *req.BackgroundExpansion {
		job := cloneJobs.create(req.Path)
		go func() {
			cloneJobs.finish(job.ID, gitService.ExpandCloneInBackground(&repo, auth))
		}()

		c.JSON(http.StatusAccepted, GitCloneResponse{
			JobID:  job.ID,
			Status: job.Status,
		})
		return
	}

	c.JSON(http.StatusOK, GitCloneResponse{})
}

// GetCloneJob godoc
//
//	@Summary		Get clone expansion job status
//	@Description	Get the status of a background clone expansion job
//	@Tags			git
//	@Produce		json
//	@Param			jobId	path		string	true	"Clone expansion job ID"
//	@Success		200		{object}	GitCloneJobResponse
//	@Router			/git/clone/jobs/{jobId} [get]
//
//	@id				GetCloneJob
func GetCloneJob(c *gin.Context) {
	job, ok := cloneJobs.get(c.Param("jobId"))
	if !ok {
		abortWithGitError(c, errCloneJobNotFound)
		return
	}

	c.JSON(http.StatusOK, job.response())
}
