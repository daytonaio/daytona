// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package libvirt

import (
	"io"

	"github.com/daytonaio/runner-win/pkg/cache"
	"github.com/daytonaio/runner-win/pkg/netrules"
	log "github.com/sirupsen/logrus"
)

type LibVirtConfig struct {
	StatesCache            *cache.StatesCache
	LogWriter              io.Writer
	AWSRegion              string
	AWSEndpointUrl         string
	AWSAccessKeyId         string
	AWSSecretAccessKey     string
	DaemonPath             string
	ComputerUsePluginPath  string
	NetRulesManager        *netrules.NetRulesManager
	ResourceLimitsDisabled bool
	DaemonStartTimeoutSec  int
	SandboxStartTimeoutSec int
	UseSnapshotEntrypoint  bool
}

func NewLibVirt(config LibVirtConfig) *LibVirt {
	if config.DaemonStartTimeoutSec <= 0 {
		log.Warnf("Invalid DaemonStartTimeoutSec value: %d. Using default value: 60 seconds", config.DaemonStartTimeoutSec)
		config.DaemonStartTimeoutSec = 60
	}

	if config.SandboxStartTimeoutSec <= 0 {
		log.Warnf("Invalid SandboxStartTimeoutSec value: %d. Using default value: 30 seconds", config.SandboxStartTimeoutSec)
		config.SandboxStartTimeoutSec = 30
	}

	return &LibVirt{
		statesCache:            config.StatesCache,
		logWriter:              config.LogWriter,
		awsRegion:              config.AWSRegion,
		awsEndpointUrl:         config.AWSEndpointUrl,
		awsAccessKeyId:         config.AWSAccessKeyId,
		awsSecretAccessKey:     config.AWSSecretAccessKey,
		daemonPath:             config.DaemonPath,
		computerUsePluginPath:  config.ComputerUsePluginPath,
		netRulesManager:        config.NetRulesManager,
		resourceLimitsDisabled: config.ResourceLimitsDisabled,
		daemonStartTimeoutSec:  config.DaemonStartTimeoutSec,
		sandboxStartTimeoutSec: config.SandboxStartTimeoutSec,
		useSnapshotEntrypoint:  config.UseSnapshotEntrypoint,
	}
}

func (l *LibVirt) ApiClient() interface{} {
	log.Infoln("ApiClient")
	return nil
}

type LibVirt struct {
	statesCache            *cache.StatesCache
	logWriter              io.Writer
	awsRegion              string
	awsEndpointUrl         string
	awsAccessKeyId         string
	awsSecretAccessKey     string
	daemonPath             string
	computerUsePluginPath  string
	netRulesManager        *netrules.NetRulesManager
	resourceLimitsDisabled bool
	daemonStartTimeoutSec  int
	sandboxStartTimeoutSec int
	useSnapshotEntrypoint  bool
}
