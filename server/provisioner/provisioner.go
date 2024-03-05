package provisioner

import (
	"github.com/daytonaio/daytona/common/types"
	provisioner_manager "github.com/daytonaio/daytona/plugins/provisioner/manager"
	"github.com/daytonaio/daytona/server/db"
	"github.com/daytonaio/daytona/server/event_bus"

	log "github.com/sirupsen/logrus"
)

func CreateWorkspace(workspace *types.Workspace) error {
	log.Info("Creating workspace")

	provisioner, err := provisioner_manager.GetProvisioner(workspace.Provisioner.Name)
	if err != nil {
		return err
	}

	_, err = (*provisioner).CreateWorkspace(workspace)
	if err != nil {
		return err
	}

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
		_, err := (*provisioner).CreateProject(project)
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
func StartWorkspace(workspace *types.Workspace) error {
	log.Info("Starting workspace")

	provisioner, err := provisioner_manager.GetProvisioner(workspace.Provisioner.Name)
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
		_, err := (*provisioner).StartProject(project)
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

	provisioner, err := provisioner_manager.GetProvisioner(workspace.Provisioner.Name)
	if err != nil {
		return err
	}

	_, err = (*provisioner).StartProject(project)
	if err != nil {
		return err
	}

	return nil
}

// WorkspacePostStart
// WorkspacePreStop
func StopWorkspace(workspace *types.Workspace) error {
	log.Info("Stopping workspace")

	provisioner, err := provisioner_manager.GetProvisioner(workspace.Provisioner.Name)
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
		_, err := (*provisioner).StopProject(project)
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

	provisioner, err := provisioner_manager.GetProvisioner(workspace.Provisioner.Name)
	if err != nil {
		return err
	}

	_, err = (*provisioner).StopProject(project)
	if err != nil {
		return err
	}

	return nil
}

// WorkspacePostStop
// WorkspacePreStop
func DestroyWorkspace(workspace *types.Workspace) error {
	log.Infof("Destroying workspace %s", workspace.Id)

	provisioner, err := provisioner_manager.GetProvisioner(workspace.Provisioner.Name)
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
		_, err := (*provisioner).DestroyProject(project)
		if err != nil {
			return err
		}
	}

	_, err = (*provisioner).DestroyWorkspace(workspace)
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
	provisioner, err := provisioner_manager.GetProvisioner(w.Provisioner.Name)
	if err != nil {
		return nil, err
	}

	return (*provisioner).GetWorkspaceInfo(w)
}
