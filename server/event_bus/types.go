// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package event_bus

import "encoding/json"

type WorkspaceEventPayload struct {
	WorkspaceName string `json:"workspaceName"`
	ProjectName   string `json:"projectName"`
}

const (
	WorkspaceEventCreatingNetwork EventName = "creating-network"
	WorkspaceEventNetworkCreated  EventName = "network-created"
	WorkspaceEventProjectCreating EventName = "project-creating"
	WorkspaceEventProjectCreated  EventName = "project-created"
	WorkspaceEventCreating        EventName = "creating"
	WorkspaceEventCreated         EventName = "created"
	WorkspaceEventStarting        EventName = "starting"
	WorkspaceEventStarted         EventName = "started"
	WorkspaceEventStopping        EventName = "stopping"
	WorkspaceEventStopped         EventName = "stopped"
	WorkspaceEventRemoving        EventName = "removing"
	WorkspaceEventRemoved         EventName = "removed"
)

type ProjectEventPayload struct {
	ProjectName   string `json:"projectName"`
	WorkspaceName string `json:"workspaceName"`
	ExtensionName string `json:"extensionName"`
}

const (
	ProjectEventCloningRepo           EventName = "cloning-repo"
	ProjectEventRepoCloned            EventName = "repo-cloned"
	ProjectEventInitializing          EventName = "project-initializing"
	ProjectEventInitialized           EventName = "project-initialized"
	ProjectEventPreparingExtension    EventName = "preparing-extension"
	ProjectEventInitializingExtension EventName = "initializing-extension"
	ProjectEventStartingExtension     EventName = "starting-extension"
	ProjectEventStarting              EventName = "project-starting"
	ProjectEventStarted               EventName = "project-started"
	ProjectEventStopping              EventName = "project-stopping"
	ProjectEventStopped               EventName = "project-stopped"
	ProjectEventRemoving              EventName = "project-removing"
	ProjectEventRemoved               EventName = "project-removed"
)

func UnmarshallWorkspaceEventPayload(payload string) (*WorkspaceEventPayload, error) {
	workspaceEventPayload := &WorkspaceEventPayload{}
	err := json.Unmarshal([]byte(payload), workspaceEventPayload)
	if err != nil {
		return nil, err
	}

	return workspaceEventPayload, nil
}

func UnmarshallProjectEventPayload(payload string) (*ProjectEventPayload, error) {
	projectEventPayload := &ProjectEventPayload{}
	err := json.Unmarshal([]byte(payload), projectEventPayload)
	if err != nil {
		return nil, err
	}

	return projectEventPayload, nil
}
