// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provisioner

import (
	"fmt"
	"io"
	"os"

	"github.com/daytonaio/daytona/internal/util"
	provider_types "github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/provider/manager"
	"github.com/daytonaio/daytona/pkg/server/config"
	"github.com/daytonaio/daytona/pkg/server/event_bus"
	"github.com/daytonaio/daytona/pkg/server/targets"
	"github.com/daytonaio/daytona/pkg/types"
	log "github.com/sirupsen/logrus"
)

func CreateWorkspace(workspace *types.Workspace) error {
	target, err := targets.GetTarget(workspace.Target)
	if err != nil {
		return err
	}

	workspaceLogFilePath, err := config.GetWorkspaceLogFilePath(workspace.Id)
	if err != nil {
		return err
	}

	workspaceLogFile, err := os.OpenFile(workspaceLogFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer workspaceLogFile.Close()

	wsLogWriter := io.MultiWriter(&util.InfoLogWriter{}, io.Writer(workspaceLogFile))

	wsLogWriter.Write([]byte("Creating workspace\n"))

	provider, err := manager.GetProvider(target.ProviderInfo.Name)
	if err != nil {
		return err
	}

	_, err = (*provider).CreateWorkspace(&provider_types.WorkspaceRequest{
		TargetOptions: target.Options,
		Workspace:     workspace,
	})
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

		projectLogWriter := io.MultiWriter(wsLogWriter, io.Writer(projectLogFile))
		projectLogWriter.Write([]byte(fmt.Sprintf("Creating project %s\n", project.Name)))

		//	todo: go routines
		event_bus.Publish(event_bus.Event{
			Name: event_bus.WorkspaceEventProjectCreating,
			Payload: event_bus.WorkspaceEventPayload{
				WorkspaceName: workspace.Name,
				ProjectName:   project.Name,
			},
		})
		_, err = (*provider).CreateProject(&provider_types.ProjectRequest{
			TargetOptions: target.Options,
			Project:       project,
		})
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

		projectLogWriter.Write([]byte(fmt.Sprintf("Project %s created\n", project.Name)))
	}

	event_bus.Publish(event_bus.Event{
		Name: event_bus.WorkspaceEventCreated,
		Payload: event_bus.WorkspaceEventPayload{
			WorkspaceName: workspace.Name,
		},
	})

	wsLogWriter.Write([]byte("Workspace creation completed\n"))

	return nil
}

func StartWorkspace(workspace *types.Workspace) error {
	target, err := targets.GetTarget(workspace.Target)
	if err != nil {
		return err
	}

	workspaceLogFilePath, err := config.GetWorkspaceLogFilePath(workspace.Id)
	if err != nil {
		return err
	}

	workspaceLogFile, err := os.OpenFile(workspaceLogFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer workspaceLogFile.Close()

	wsLogWriter := io.MultiWriter(&util.InfoLogWriter{}, io.Writer(workspaceLogFile))

	wsLogWriter.Write([]byte("Starting workspace\n"))

	provider, err := manager.GetProvider(target.ProviderInfo.Name)
	if err != nil {
		return err
	}

	(*provider).StartWorkspace(&provider_types.WorkspaceRequest{
		TargetOptions: target.Options,
		Workspace:     workspace,
	})

	event_bus.Publish(event_bus.Event{
		Name: event_bus.WorkspaceEventStarting,
		Payload: event_bus.WorkspaceEventPayload{
			WorkspaceName: workspace.Name,
		},
	})

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

		projectLogWriter := io.MultiWriter(wsLogWriter, io.Writer(projectLogFile))
		projectLogWriter.Write([]byte(fmt.Sprintf("Starting project %s\n", project.Name)))

		//	todo: go routines
		_, err = (*provider).StartProject(&provider_types.ProjectRequest{
			TargetOptions: target.Options,
			Project:       project,
		})
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
	target, err := targets.GetTarget(project.Target)
	if err != nil {
		return err
	}

	provider, err := manager.GetProvider(target.ProviderInfo.Name)
	if err != nil {
		return err
	}

	_, err = (*provider).StartProject(&provider_types.ProjectRequest{
		TargetOptions: target.Options,
		Project:       project,
	})
	if err != nil {
		return err
	}

	return nil
}

func StopWorkspace(workspace *types.Workspace) error {
	log.Info("Stopping workspace")

	target, err := targets.GetTarget(workspace.Target)
	if err != nil {
		return err
	}

	provider, err := manager.GetProvider(target.ProviderInfo.Name)
	if err != nil {
		return err
	}

	(*provider).StopWorkspace(&provider_types.WorkspaceRequest{
		TargetOptions: target.Options,
		Workspace:     workspace,
	})

	event_bus.Publish(event_bus.Event{
		Name: event_bus.WorkspaceEventStopping,
		Payload: event_bus.WorkspaceEventPayload{
			WorkspaceName: workspace.Name,
		},
	})

	for _, project := range workspace.Projects {
		//	todo: go routines
		_, err := (*provider).StopProject(&provider_types.ProjectRequest{
			TargetOptions: target.Options,
			Project:       project,
		})
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
	target, err := targets.GetTarget(project.Target)
	if err != nil {
		return err
	}

	provider, err := manager.GetProvider(target.ProviderInfo.Name)
	if err != nil {
		return err
	}

	_, err = (*provider).StopProject(&provider_types.ProjectRequest{
		TargetOptions: target.Options,
		Project:       project,
	})
	if err != nil {
		return err
	}

	return nil
}

func DestroyWorkspace(workspace *types.Workspace) error {
	log.Infof("Destroying workspace %s", workspace.Id)

	target, err := targets.GetTarget(workspace.Target)
	if err != nil {
		return err
	}

	provider, err := manager.GetProvider(target.ProviderInfo.Name)
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
		_, err := (*provider).DestroyProject(&provider_types.ProjectRequest{
			TargetOptions: target.Options,
			Project:       project,
		})
		if err != nil {
			return err
		}
	}

	_, err = (*provider).DestroyWorkspace(&provider_types.WorkspaceRequest{
		TargetOptions: target.Options,
		Workspace:     workspace,
	})
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

func GetWorkspaceInfo(workspace *types.Workspace) (*types.WorkspaceInfo, error) {
	target, err := targets.GetTarget(workspace.Target)
	if err != nil {
		return nil, err
	}

	provider, err := manager.GetProvider(target.ProviderInfo.Name)
	if err != nil {
		return nil, err
	}

	return (*provider).GetWorkspaceInfo(&provider_types.WorkspaceRequest{
		TargetOptions: target.Options,
		Workspace:     workspace,
	})
}
