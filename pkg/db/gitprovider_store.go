// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"gorm.io/gorm"

	. "github.com/daytonaio/daytona/pkg/db/dto"
	"github.com/daytonaio/daytona/pkg/gitprovider"
)

type GitProviderStore struct {
	db *gorm.DB
}

func (p *GitProviderStore) List() ([]*gitprovider.GitProvider, error) {
	gitProviderDTOs := []GitProviderDTO{}
	tx := p.db.Find(&gitProviderDTOs)
	if tx.Error != nil {
		return nil, tx.Error
	}

	gitProviders := []*gitprovider.GitProvider{}
	for _, gitProviderDTO := range gitProviderDTOs {
		gitProvider := ToGitProvider(gitProviderDTO)
		gitProviders = append(gitProviders, &gitProvider)
	}

	return gitProviders, nil
}

func (p *GitProviderStore) Find(idOrName string) (*gitprovider.GitProvider, error) {
	gitProviderDTO := GitProviderDTO{}
	tx := p.db.Where("id = ? OR name = ?", idOrName, idOrName).First(&gitProviderDTO)
	if tx.Error != nil {
		return nil, tx.Error
	}

	gitProvider := ToGitProvider(gitProviderDTO)

	return &gitProvider, nil
}

func (p *GitProviderStore) Save(gitProvider *gitprovider.GitProvider) error {
	gitProviderDTO := ToGitProviderDTO(*gitProvider)
	tx := p.db.Save(&gitProviderDTO)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

func (p *GitProviderStore) Delete(gitProvider *gitprovider.GitProvider) error {
	gitProviderDTO := ToGitProviderDTO(*gitProvider)
	tx := p.db.Delete(&gitProviderDTO)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}
