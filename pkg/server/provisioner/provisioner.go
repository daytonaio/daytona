// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provisioner

import (
	"fmt"
	"io"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/logger"
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

	logsDir, err := config.GetWorkspaceLogsDir()
	if err != nil {
		return err
	}

	workspaceLogger := logger.GetWorkspaceLogger(logsDir, workspace.Id)
	defer workspaceLogger.Close()

	wsLogWriter := io.MultiWriter(&util.InfoLogWriter{}, workspaceLogger)

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
		projectLogger := logger.GetProjectLogger(logsDir, workspace.Id, project.Name)
		defer projectLogger.Close()

		projectLogWriter := io.MultiWriter(wsLogWriter, projectLogger)
		projectLogWriter.Write([]byte(fmt.Sprintf("Creating project %s\n", project.Name)))

		//	todo: go routines
		err = event_bus.Publish(event_bus.Event{
			Name: event_bus.WorkspaceEventProjectCreating,
			Payload: event_bus.WorkspaceEventPayload{
				WorkspaceName: workspace.Name,
				ProjectName:   project.Name,
			},
		})
		if err != nil {
			log.Error(err)
		}
		_, err = (*provider).CreateProject(&provider_types.ProjectRequest{
			TargetOptions: target.Options,
			Project:       project,
		})
		if err != nil {
			return err
		}
		err = event_bus.Publish(event_bus.Event{
			Name: event_bus.WorkspaceEventProjectCreated,
			Payload: event_bus.WorkspaceEventPayload{
				WorkspaceName: workspace.Name,
				ProjectName:   project.Name,
			},
		})
		if err != nil {
			log.Error(err)
		}

		projectLogWriter.Write([]byte(fmt.Sprintf("Project %s created\n", project.Name)))
	}

	err = event_bus.Publish(event_bus.Event{
		Name: event_bus.WorkspaceEventCreated,
		Payload: event_bus.WorkspaceEventPayload{
			WorkspaceName: workspace.Name,
		},
	})
	if err != nil {
		log.Error(err)
	}

	wsLogWriter.Write([]byte("Workspace creation completed\n"))

	return nil
}

func StartWorkspace(workspace *types.Workspace) error {
	logsDir, err := config.GetWorkspaceLogsDir()
	if err != nil {
		return err
	}

	target, err := targets.GetTarget(workspace.Target)
	if err != nil {
		return err
	}

	workspaceLogger := logger.GetWorkspaceLogger(logsDir, workspace.Id)
	defer workspaceLogger.Close()

	wsLogWriter := io.MultiWriter(&util.InfoLogWriter{}, workspaceLogger)

	wsLogWriter.Write([]byte("Starting workspace\n"))

	provider, err := manager.GetProvider(target.ProviderInfo.Name)
	if err != nil {
		return err
	}

	_, err = (*provider).StartWorkspace(&provider_types.WorkspaceRequest{
		TargetOptions: target.Options,
		Workspace:     workspace,
	})
	if err != nil {
		return err
	}

	err = event_bus.Publish(event_bus.Event{
		Name: event_bus.WorkspaceEventStarting,
		Payload: event_bus.WorkspaceEventPayload{
			WorkspaceName: workspace.Name,
		},
	})
	if err != nil {
		log.Error(err)
	}

	for _, project := range workspace.Projects {
		projectLogger := logger.GetProjectLogger(logsDir, workspace.Id, project.Name)
		defer projectLogger.Close()

		projectLogWriter := io.MultiWriter(wsLogWriter, projectLogger)
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

	err = event_bus.Publish(event_bus.Event{
		Name: event_bus.WorkspaceEventStarted,
		Payload: event_bus.WorkspaceEventPayload{
			WorkspaceName: workspace.Name,
		},
	})
	if err != nil {
		log.Error(err)
	}
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

	_, err = (*provider).StopWorkspace(&provider_types.WorkspaceRequest{
		TargetOptions: target.Options,
		Workspace:     workspace,
	})
	if err != nil {
		return err
	}

	err = event_bus.Publish(event_bus.Event{
		Name: event_bus.WorkspaceEventStopping,
		Payload: event_bus.WorkspaceEventPayload{
			WorkspaceName: workspace.Name,
		},
	})
	if err != nil {
		log.Error(err)
	}

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

	err = event_bus.Publish(event_bus.Event{
		Name: event_bus.WorkspaceEventStopped,
		Payload: event_bus.WorkspaceEventPayload{
			WorkspaceName: workspace.Name,
		},
	})
	if err != nil {
		log.Error(err)
	}

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

	err = event_bus.Publish(event_bus.Event{
		Name: event_bus.WorkspaceEventRemoving,
		Payload: event_bus.WorkspaceEventPayload{
			WorkspaceName: workspace.Name,
		},
	})
	if err != nil {
		log.Error(err)
	}

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

	err = event_bus.Publish(event_bus.Event{
		Name: event_bus.WorkspaceEventRemoved,
		Payload: event_bus.WorkspaceEventPayload{
			WorkspaceName: workspace.Name,
		},
	})
	if err != nil {
		log.Error(err)
	}

	err = config.DeleteWorkspaceLogs(workspace.Id)
	if err != nil {
		return err
	}

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
