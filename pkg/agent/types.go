// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package agent

import (
	"io"
	"time"

	"github.com/daytonaio/daytona/pkg/agent/config"
	"github.com/daytonaio/daytona/pkg/serverapiclient"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

type GitService interface {
	CloneRepository(project *serverapiclient.Project, auth *http.BasicAuth) error
	RepositoryExists(project *serverapiclient.Project) (bool, error)
	SetGitConfig(userData *serverapiclient.GitUser) error
}

type SshServer interface {
	Start() error
}

type TailscaleServer interface {
	Start() error
}

type Agent struct {
	Config    *config.Config
	Git       GitService
	Ssh       SshServer
	Tailscale TailscaleServer
	LogWriter io.Writer
	startTime time.Time
}
