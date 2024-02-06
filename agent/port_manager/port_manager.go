package port_manager

import (
	"context"
	"fmt"

	"github.com/daytonaio/daytona/agent/event_bus"
)

type HostPort uint16
type ContainerPort uint16

type PortForward struct {
	HostPort      HostPort
	ContainerPort ContainerPort
	ctx           context.Context
	cancelFunc    context.CancelFunc
}

type PortForwards map[ContainerPort]PortForward

type WorkspacePortForward struct {
	WorkspaceName       string
	ProjectPortForwards map[string]PortForwards
}

var workspacePortForwards = make(map[string]WorkspacePortForward)

func subscribeToWorkspaceEvents() {
	unsubscribe := make(chan bool, 1)

	for event := range event_bus.SubscribeWithFilter(unsubscribe, func(i event_bus.Event) bool {
		switch i.Name {
		case event_bus.WorkspaceEventStopping:
			fallthrough
		case event_bus.WorkspaceEventStopped:
			fallthrough
		case event_bus.ProjectEventStopping:
			fallthrough
		case event_bus.ProjectEventStopped:
			fallthrough
		case event_bus.WorkspaceEventRemoving:
			fallthrough
		case event_bus.WorkspaceEventRemoved:
			return true
		}

		return false
	}) {
		if workspaceEventPayload, ok := event.Payload.(event_bus.WorkspaceEventPayload); ok {
			StopAllWorkspaceForwards(workspaceEventPayload.WorkspaceName)
		} else if projectEventPayload, ok := event.Payload.(event_bus.ProjectEventPayload); ok {
			StopAllWorkspaceProjectForwards(projectEventPayload.WorkspaceName, fmt.Sprintf("%s-%s", projectEventPayload.WorkspaceName, projectEventPayload.ProjectName))
		}
	}
}
