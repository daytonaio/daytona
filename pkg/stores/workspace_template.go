// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package stores

import (
	"context"
	"errors"

	"github.com/daytonaio/daytona/pkg/models"
)

type WorkspaceTemplateFilter struct {
	Name                *string
	Url                 *string
	Default             *bool
	PrebuildId          *string
	GitProviderConfigId *string
}

type PrebuildFilter struct {
	WorkspaceTemplateName *string
	Id                    *string
	Branch                *string
	CommitInterval        *int
	TriggerFiles          *[]string
}

type WorkspaceTemplateStore interface {
	IStore
	List(ctx context.Context, filter *WorkspaceTemplateFilter) ([]*models.WorkspaceTemplate, error)
	Find(ctx context.Context, filter *WorkspaceTemplateFilter) (*models.WorkspaceTemplate, error)
	Save(ctx context.Context, workspaceTemplate *models.WorkspaceTemplate) error
	Delete(ctx context.Context, workspaceTemplate *models.WorkspaceTemplate) error
}

var (
	ErrWorkspaceTemplateNotFound = errors.New("workspace template not found")
	ErrPrebuildNotFound          = errors.New("prebuild not found")
)

func IsWorkspaceTemplateNotFound(err error) bool {
	return err.Error() == ErrWorkspaceTemplateNotFound.Error()
}

func IsPrebuildNotFound(err error) bool {
	return err.Error() == ErrPrebuildNotFound.Error()
}
