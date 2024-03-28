// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package agent

import (
	"github.com/daytonaio/daytona/pkg/agent/config"
	"github.com/daytonaio/daytona/pkg/serverapiclient"
)

type GitService interface {
	CloneRepository(project *serverapiclient.Project, authToken *string) error
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
}
