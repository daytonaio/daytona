// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provisioner

import (
	"fmt"
	"io"
	"os"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/provider/manager"
	"github.com/daytonaio/daytona/pkg/server/config"
	"github.com/daytonaio/daytona/pkg/server/db"
	"github.com/daytonaio/daytona/pkg/server/event_bus"
	"github.com/daytonaio/daytona/pkg/types"
	log "github.com/sirupsen/logrus"
)

func CreateWorkspace(workspace *types.Workspace) error {
	workspaceLogFilePath, err := config.GetWorkspaceLogFilePath(workspace.Id)
	if err != nil {
		return err
	}

	workspaceLogFile, err := os.OpenFile(workspaceLogFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer workspaceLogFile.Close()

	wsMultiWriter := io.MultiWriter(&util.InfoLogWriter{}, io.Writer(workspaceLogFile))

	wsMultiWriter.Write([]byte("Creating workspace\n"))
	// log.Info("Creating workspace")

	provider, err := manager.GetProvider(workspace.Provider.Name)
	if err != nil {
		return err
	}

	_, err = (*provider).CreateWorkspace(workspace)
	if err != nil {
		return err
	}

	log.Debug("Projects to initialize", workspace.Projects)

	for _, project := range workspace.Projects {

		projectLogFilePath, err := config.GetProjectLogFilePath(workspace.Id, project.Name)
		if err != nil {
			return err
		}

		projectLogFile, err := os.OpenFile(projectLogFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return err
		}
		defer projectLogFile.Close()

		projectMultiWriter := io.MultiWriter(&util.InfoLogWriter{}, io.Writer(workspaceLogFile), io.Writer(projectLogFile))
		projectMultiWriter.Write([]byte(fmt.Sprintf("Creating project %s\n", project.Name)))

		//	todo: go routines
		event_bus.Publish(event_bus.Event{
			Name: event_bus.WorkspaceEventProjectCreating,
			Payload: event_bus.WorkspaceEventPayload{
				WorkspaceName: workspace.Name,
				ProjectName:   project.Name,
			},
		})
		_, err = (*provider).CreateProject(project)
		if err != nil {
			return err
		}
		event_bus.Publish(event_bus.Event{
			Name: event_bus.WorkspaceEventProjectCreated,
			Payload: event_bus.WorkspaceEventPayload{
				WorkspaceName: workspace.Name,
				ProjectName:   project.Name,
			},
		})
	}

	event_bus.Publish(event_bus.Event{
		Name: event_bus.WorkspaceEventCreated,
		Payload: event_bus.WorkspaceEventPayload{
			WorkspaceName: workspace.Name,
		},
	})

	wsMultiWriter.Write([]byte("Workspace creation completed\n"))

	return nil
}

// WorkspacePostCreate
// WorkspacePreStart
func StartWorkspace(workspace *types.Workspace) error {
	log.Info("Starting workspace")

	provider, err := manager.GetProvider(workspace.Provider.Name)
	if err != nil {
		return err
	}

	(*provider).StartWorkspace(workspace)

	event_bus.Publish(event_bus.Event{
		Name: event_bus.WorkspaceEventStarting,
		Payload: event_bus.WorkspaceEventPayload{
			WorkspaceName: workspace.Name,
		},
	})

	for _, project := range workspace.Projects {
		//	todo: go routines
		_, err := (*provider).StartProject(project)
		if err != nil {
			return err
		}
	}

	event_bus.Publish(event_bus.Event{
		Name: event_bus.WorkspaceEventStarted,
		Payload: event_bus.WorkspaceEventPayload{
			WorkspaceName: workspace.Name,
		},
	})
	return nil
}

func StartProject(project *types.Project) error {
	workspace, err := db.FindWorkspace(project.WorkspaceId)
	if err != nil {
		return err
	}

	provider, err := manager.GetProvider(workspace.Provider.Name)
	if err != nil {
		return err
	}

	_, err = (*provider).StartProject(project)
	if err != nil {
		return err
	}

	return nil
}

// WorkspacePostStart
// WorkspacePreStop
func StopWorkspace(workspace *types.Workspace) error {
	log.Info("Stopping workspace")

	provider, err := manager.GetProvider(workspace.Provider.Name)
	if err != nil {
		return err
	}

	(*provider).StopWorkspace(workspace)

	event_bus.Publish(event_bus.Event{
		Name: event_bus.WorkspaceEventStopping,
		Payload: event_bus.WorkspaceEventPayload{
			WorkspaceName: workspace.Name,
		},
	})

	for _, project := range workspace.Projects {
		//	todo: go routines
		_, err := (*provider).StopProject(project)
		if err != nil {
			return err
		}
	}

	event_bus.Publish(event_bus.Event{
		Name: event_bus.WorkspaceEventStopped,
		Payload: event_bus.WorkspaceEventPayload{
			WorkspaceName: workspace.Name,
		},
	})

	return nil
}

func StopProject(project *types.Project) error {
	workspace, err := db.FindWorkspace(project.WorkspaceId)
	if err != nil {
		return err
	}

	provider, err := manager.GetProvider(workspace.Provider.Name)
	if err != nil {
		return err
	}

	_, err = (*provider).StopProject(project)
	if err != nil {
		return err
	}

	return nil
}

// WorkspacePostStop
// WorkspacePreStop
func DestroyWorkspace(workspace *types.Workspace) error {
	log.Infof("Destroying workspace %s", workspace.Id)

	provider, err := manager.GetProvider(workspace.Provider.Name)
	if err != nil {
		return err
	}

	event_bus.Publish(event_bus.Event{
		Name: event_bus.WorkspaceEventRemoving,
		Payload: event_bus.WorkspaceEventPayload{
			WorkspaceName: workspace.Name,
		},
	})

	for _, project := range workspace.Projects {
		//	todo: go routines
		_, err := (*provider).DestroyProject(project)
		if err != nil {
			return err
		}
	}

	_, err = (*provider).DestroyWorkspace(workspace)
	if err != nil {
		return err
	}

	event_bus.Publish(event_bus.Event{
		Name: event_bus.WorkspaceEventRemoved,
		Payload: event_bus.WorkspaceEventPayload{
			WorkspaceName: workspace.Name,
		},
	})

	log.Infof("Workspace %s destroyed", workspace.Id)

	return nil
}

func GetWorkspaceInfo(w *types.Workspace) (*types.WorkspaceInfo, error) {
	provider, err := manager.GetProvider(w.Provider.Name)
	if err != nil {
		return nil, err
	}

	return (*provider).GetWorkspaceInfo(w)
}
