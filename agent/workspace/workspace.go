// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"context"
	"dagent/agent/event_bus"
	"dagent/credentials"
	"errors"
	"fmt"
	"regexp"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
)

type Extension interface {
	Name() string
	PreInit(project Project) error
	Init(project Project) error
	Start(project Project) error
	Info(project Project) string
	LivenessProbe(project Project) (bool, error)
	LivenessProbeTimeout() int
}

// DO NOT INSTANTIATE THIS STRUCT DIRECTLY!
type Workspace struct {
	Name        string `gorm:"primaryKey"`
	Cwd         string
	Credentials credentials.CredentialProvider `gorm:"-"`
	Extensions  []Extension                    `gorm:"-"`
	Projects    []Project                      `gorm:"serializer:json"`
}

type Repository struct {
	Url      string   `json:"url"`
	Branch   *string  `default:"main" json:"branch,omitempty"`
	SHA      *string  `json:"sha,omitempty"`
	Owner    *string  `json:"owner,omitempty"`
	PrNumber *float32 `json:"prNumber,omitempty"`
	Source   *string  `json:"source,omitempty"`
	Path     *string  `json:"path,omitempty"`
}

type WorkspaceParams struct {
	Name         string
	Cwd          string
	Credentials  credentials.CredentialProvider
	Extensions   []Extension
	Repositories []Repository
}

type WorkspaceInfo struct {
	Name     string
	Projects []ProjectInfo
	Cwd      string
}

func New(params WorkspaceParams) (*Workspace, error) {
	isAlphaNumeric := regexp.MustCompile(`^[a-zA-Z0-9-]+$`).MatchString
	if !isAlphaNumeric(params.Name) {
		return nil, errors.New("name is not a valid alphanumeric string")
	}

	w := Workspace{
		Name:        params.Name,
		Cwd:         params.Cwd,
		Credentials: params.Credentials,
		Extensions:  params.Extensions,
	}
	w.Projects = []Project{}

	for _, repo := range params.Repositories {
		project := Project{
			Repository: repo,
			Workspace:  &w,
		}
		w.Projects = append(w.Projects, project)
	}

	return &w, nil
}

func (w Workspace) Create() error {
	log.Info("Creating workspace")

	event_bus.Publish(event_bus.Event{
		Name: event_bus.WorkspaceEventCreatingNetwork,
		Payload: event_bus.WorkspaceEventPayload{
			WorkspaceName: w.Name,
		},
	})

	err := w.createNetwork()
	if err != nil {
		return err
	}

	event_bus.Publish(event_bus.Event{
		Name: event_bus.WorkspaceEventNetworkCreated,
		Payload: event_bus.WorkspaceEventPayload{
			WorkspaceName: w.Name,
		},
	})

	log.Debug("Projects to initialize", w.Projects)

	for _, project := range w.Projects {
		//	todo: go routines
		event_bus.Publish(event_bus.Event{
			Name: event_bus.WorkspaceEventProjectInitializing,
			Payload: event_bus.WorkspaceEventPayload{
				WorkspaceName: w.Name,
				ProjectName:   project.GetName(),
			},
		})
		err := project.initialize()
		if err != nil {
			return err
		}
		event_bus.Publish(event_bus.Event{
			Name: event_bus.WorkspaceEventProjectInitialized,
			Payload: event_bus.WorkspaceEventPayload{
				WorkspaceName: w.Name,
				ProjectName:   project.GetName(),
			},
		})
	}

	event_bus.Publish(event_bus.Event{
		// TODO: possible error - might need WorkspacEventCreated instead of Creating - check with Vedran
		Name: event_bus.WorkspaceEventCreating,
		Payload: event_bus.WorkspaceEventPayload{
			WorkspaceName: w.Name,
		},
	})

	return nil
}

func (w Workspace) Start() error {
	log.Info("Starting workspace")

	event_bus.Publish(event_bus.Event{
		Name: event_bus.WorkspaceEventStarting,
		Payload: event_bus.WorkspaceEventPayload{
			WorkspaceName: w.Name,
		},
	})

	for _, project := range w.Projects {
		//	todo: go routines
		err := project.Start()
		if err != nil {
			return err
		}
	}

	event_bus.Publish(event_bus.Event{
		Name: event_bus.WorkspaceEventStarted,
		Payload: event_bus.WorkspaceEventPayload{
			WorkspaceName: w.Name,
		},
	})

	return nil
}

func (w Workspace) StartProject(projectName string) error {
	log.Info(fmt.Sprintf("Starting project %s", projectName))

	for _, project := range w.Projects {
		if project.GetName() == projectName {
			//	todo: go routines
			err := project.Start()
			if err != nil {
				return err
			}
		}
	}

	log.Info(fmt.Sprintf("Started project %s", projectName))

	return nil
}

func (w Workspace) Stop() error {
	log.Info("Stopping workspace")

	event_bus.Publish(event_bus.Event{
		Name: event_bus.WorkspaceEventStopping,
		Payload: event_bus.WorkspaceEventPayload{
			WorkspaceName: w.Name,
		},
	})

	for _, project := range w.Projects {
		//	todo: go routines
		err := project.Stop()
		if err != nil {
			return err
		}
	}

	event_bus.Publish(event_bus.Event{
		Name: event_bus.WorkspaceEventStopped,
		Payload: event_bus.WorkspaceEventPayload{
			WorkspaceName: w.Name,
		},
	})

	return nil
}

func (w Workspace) StopProject(projectName string) error {
	log.Info(fmt.Sprintf("Stopping project %s", projectName))

	for _, project := range w.Projects {
		if project.GetName() == projectName {
			//	todo: go routines
			err := project.Stop()
			if err != nil {
				return err
			}
		}
	}

	log.Info(fmt.Sprintf("Stopped project %s", projectName))

	return nil
}

func (w Workspace) Remove(force bool) error {
	log.Info("Removing workspace")

	event_bus.Publish(event_bus.Event{
		Name: event_bus.WorkspaceEventRemoving,
		Payload: event_bus.WorkspaceEventPayload{
			WorkspaceName: w.Name,
		},
	})

	for _, project := range w.Projects {
		//	todo: go routines
		err := project.Remove(force)
		if err != nil {
			return err
		}
	}

	w.removeNetwork()

	event_bus.Publish(event_bus.Event{
		Name: event_bus.WorkspaceEventRemoved,
		Payload: event_bus.WorkspaceEventPayload{
			WorkspaceName: w.Name,
		},
	})

	return nil
}

func (w Workspace) Info() (*WorkspaceInfo, error) {
	projects := []ProjectInfo{}

	for _, project := range w.Projects {
		projectInfo, err := project.Info()
		if err != nil {
			return nil, err
		}

		projects = append(projects, *projectInfo)
	}

	return &WorkspaceInfo{
		Name:     w.Name,
		Projects: projects,
		Cwd:      w.Cwd,
	}, nil
}

func (w Workspace) createNetwork() error {
	log.Debug("Initializing network")
	ctx := context.Background()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	networks, err := cli.NetworkList(ctx, types.NetworkListOptions{})
	if err != nil {
		return err
	}

	for _, network := range networks {
		if network.Name == w.Name {
			log.WithFields(log.Fields{
				"workspace": w.Name,
			}).Debug("Network already exists")
			return nil
		}
	}

	_, err = cli.NetworkCreate(ctx, w.Name, types.NetworkCreate{
		Attachable: true,
	})
	if err != nil {
		return err
	}

	log.Debug("Network initialized")

	return nil
}

func (w Workspace) removeNetwork() error {
	log.Debug("Removing network")
	ctx := context.Background()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	networks, err := cli.NetworkList(ctx, types.NetworkListOptions{})
	if err != nil {
		return err
	}

	for _, network := range networks {
		if network.Name == w.Name {
			err := cli.NetworkRemove(ctx, network.ID)
			if err != nil {
				return err
			}
		}
	}

	log.Debug("Network removed")

	return nil
}

func (w Workspace) GetProject(projectName string) (Project, error) {
	for _, p := range w.Projects {
		if p.GetName() == projectName {
			return p, nil
		}
	}

	return Project{}, errors.New("project not found")
}
