package provisioner

import (
	"github.com/daytonaio/daytona/agent/event_bus"
	"github.com/daytonaio/daytona/agent/workspace"
	"github.com/daytonaio/daytona/plugin"

	log "github.com/sirupsen/logrus"
)

func CreateWorkspace(workspace workspace.Workspace) error {
	log.Info("Creating workspace")

	provisioner, err := plugin.GetProvisionerPlugin(workspace.Provisioner.Name)
	if err != nil {
		return err
	}

	(*provisioner).CreateWorkspace(workspace)

	log.Debug("Projects to initialize", workspace.Projects)

	for _, project := range workspace.Projects {
		//	todo: go routines
		event_bus.Publish(event_bus.Event{
			Name: event_bus.WorkspaceEventProjectCreating,
			Payload: event_bus.WorkspaceEventPayload{
				WorkspaceName: workspace.Name,
				ProjectName:   project.Name,
			},
		})
		err := (*provisioner).CreateProject(project)
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

	return nil

}

// WorkspacePostCreate
// WorkspacePreStart
func StartWorkspace(workspace workspace.Workspace) error {
	log.Info("Starting workspace")

	provisioner, err := plugin.GetProvisionerPlugin(workspace.Provisioner.Name)
	if err != nil {
		return err
	}

	(*provisioner).StartWorkspace(workspace)

	event_bus.Publish(event_bus.Event{
		Name: event_bus.WorkspaceEventStarting,
		Payload: event_bus.WorkspaceEventPayload{
			WorkspaceName: workspace.Name,
		},
	})

	for _, project := range workspace.Projects {
		//	todo: go routines
		err := (*provisioner).StartProject(project)
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

func StartProject(project workspace.Project) error {
	provisioner, err := plugin.GetProvisionerPlugin(project.Workspace.Provisioner.Name)
	if err != nil {
		return err
	}

	err = (*provisioner).StartProject(project)
	if err != nil {
		return err
	}

	return nil
}

// WorkspacePostStart
// WorkspacePreStop
func StopWorkspace(workspace workspace.Workspace) error {
	log.Info("Stopping workspace")

	provisioner, err := plugin.GetProvisionerPlugin(workspace.Provisioner.Name)
	if err != nil {
		return err
	}

	(*provisioner).StopWorkspace(workspace)

	event_bus.Publish(event_bus.Event{
		Name: event_bus.WorkspaceEventStopping,
		Payload: event_bus.WorkspaceEventPayload{
			WorkspaceName: workspace.Name,
		},
	})

	for _, project := range workspace.Projects {
		//	todo: go routines
		err := (*provisioner).StopProject(project)
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

func StopProject(project workspace.Project) error {
	provisioner, err := plugin.GetProvisionerPlugin(project.Workspace.Provisioner.Name)
	if err != nil {
		return err
	}

	err = (*provisioner).StopProject(project)
	if err != nil {
		return err
	}

	return nil
}

// WorkspacePostStop
// WorkspacePreStop
func DestroyWorkspace(workspace workspace.Workspace) error {
	log.Info("Destroying workspace")

	provisioner, err := plugin.GetProvisionerPlugin(workspace.Provisioner.Name)
	if err != nil {
		return err
	}

	(*provisioner).DestroyWorkspace(workspace)

	event_bus.Publish(event_bus.Event{
		Name: event_bus.WorkspaceEventRemoving,
		Payload: event_bus.WorkspaceEventPayload{
			WorkspaceName: workspace.Name,
		},
	})

	for _, project := range workspace.Projects {
		//	todo: go routines
		err := (*provisioner).DestroyProject(project)
		if err != nil {
			return err
		}
	}

	event_bus.Publish(event_bus.Event{
		Name: event_bus.WorkspaceEventRemoved,
		Payload: event_bus.WorkspaceEventPayload{
			WorkspaceName: workspace.Name,
		},
	})

	return nil
}

func GetWorkspaceInfo(w workspace.Workspace) (*workspace.WorkspaceInfo, error) {
	provisioner, err := plugin.GetProvisionerPlugin(w.Provisioner.Name)
	if err != nil {
		return nil, err
	}

	metadata, err := (*provisioner).GetWorkspaceMetadata(w)
	if err != nil {
		return nil, err
	}

	projects := []workspace.ProjectInfo{}

	for _, project := range w.Projects {
		projectInfo, err := (*provisioner).GetProjectInfo(project)
		if err != nil {
			return nil, err
		}

		projects = append(projects, *projectInfo)
	}

	return &workspace.WorkspaceInfo{
		Name:                w.Name,
		Provisioner:         w.Provisioner,
		Projects:            projects,
		ProvisionerMetadata: metadata,
	}, nil
}
