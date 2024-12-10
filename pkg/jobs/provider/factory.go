// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"github.com/daytonaio/daytona/pkg/jobs"
	"github.com/daytonaio/daytona/pkg/models"
)

type IProviderJobFactory interface {
	Create(job models.Job) jobs.IJob
}

type ProviderJobFactory struct {
	config ProviderJobFactoryConfig
}

type ProviderJobFactoryConfig struct {
}

func NewProviderJobFactory(config ProviderJobFactoryConfig) IProviderJobFactory {
	return &ProviderJobFactory{
		config: config,
	}
}

func (f *ProviderJobFactory) Create(job models.Job) jobs.IJob {
	return &ProviderJob{
		Job: job,
	}
}
