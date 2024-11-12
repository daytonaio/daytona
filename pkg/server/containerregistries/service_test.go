// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package containerregistries_test

import (
	"testing"

	t_containerregistries "github.com/daytonaio/daytona/internal/testing/server/containerregistries"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/server/containerregistries"
	"github.com/stretchr/testify/require"
)

func TestContainerRegistryService(t *testing.T) {
	crStore := t_containerregistries.NewInMemoryContainerRegistryStore()

	service := containerregistries.NewContainerRegistryService(containerregistries.ContainerRegistryServiceConfig{
		Store: crStore,
	})

	t.Run("CreateContainerRegistry", func(t *testing.T) {
		var crOrg = &models.ContainerRegistry{
			Server:   "example.com",
			Username: "user",
			Password: "password",
		}

		err := service.Save(crOrg)

		require.Nil(t, err)

		cr, err := service.Find("example.com")

		require.Nil(t, err)
		require.EqualValues(t, crOrg, cr)
	})

	t.Run("FindByImageName", func(t *testing.T) {
		var crOrg = &models.ContainerRegistry{
			Server:   "example.com",
			Username: "user",
			Password: "password",
		}

		err := service.Save(crOrg)

		require.Nil(t, err)

		cr, err := service.FindByImageName("example.com/image/image")

		require.Nil(t, err)
		require.EqualValues(t, crOrg, cr)
	})
}
