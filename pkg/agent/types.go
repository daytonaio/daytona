// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package agent

import (
	"io"
	"time"

	"github.com/daytonaio/daytona/pkg/agent/config"
	"github.com/daytonaio/daytona/pkg/git"
)

type SshServer interface {
	Start() error
}

type TailscaleServer interface {
	Start() error
}

type Agent struct {
	Config           *config.Config
	Git              git.IGitService
	Ssh              SshServer
	Tailscale        TailscaleServer
	LogWriter        io.Writer
	TelemetryEnabled bool
	startTime        time.Time
}
