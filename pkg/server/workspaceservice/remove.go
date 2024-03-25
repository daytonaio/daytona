// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaceservice

import (
	"errors"
	"fmt"

	"github.com/daytonaio/daytona/pkg/server/auth"
	"github.com/daytonaio/daytona/pkg/server/db"
	"github.com/daytonaio/daytona/pkg/server/provisioner"

	log "github.com/sirupsen/logrus"
)

func RemoveWorkspace(workspaceId string) error {
	w, err := db.FindWorkspaceByIdOrName(workspaceId)
	if err != nil {
		return errors.New("workspace not found")
	}

	err = provisioner.DestroyWorkspace(w)
	if err != nil {
		return err
	}

	for _, project := range w.Projects {
		err := auth.RevokeApiKey(fmt.Sprintf("%s/%s", w.Id, project.Name))
		if err != nil {
			// Should not fail the whole operation if the API key cannot be revoked
			log.Error(err)
		}
	}

	return db.DeleteWorkspace(w)
}
