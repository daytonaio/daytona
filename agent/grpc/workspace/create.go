// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace_grpc

import (
	"dagent/agent/config"
	"dagent/agent/event_bus"
	"dagent/agent/workspace"
	"dagent/credentials"
	"dagent/extensions/ssh"
	"dagent/extensions/vsc_server"
	daytona_proto "dagent/grpc/proto"
	"encoding/json"
	"errors"
	"time"

	log "github.com/sirupsen/logrus"
)

func (m *WorkspaceServer) Create(request *daytona_proto.CreateWorkspaceRequest, stream daytona_proto.Workspace_CreateServer) error {
	_, err := workspace.LoadFromDB(request.Name)
	if err == nil {
		return errors.New("workspace already exists")
	}

	c, err := config.GetConfig()
	if err != nil {
		return err
	}

	credClient := &credentials.CredentialsClient{}

	extensions := []workspace.Extension{}

	vsc_server_extension := vsc_server.VscServerExtension{}
	extensions = append(extensions, vsc_server_extension)

	sshPublicKey, err := config.GetWorkspacePublicKey()
	if err != nil {
		log.Error(err)
		return err
	}

	ssh := ssh.SshExtension{
		PublicKey: sshPublicKey,
	}
	extensions = append(extensions, ssh)

	var repositories []workspace.Repository
	for _, repo := range request.Repositories {
		repositories = append(repositories, workspace.Repository{
			Url: repo,
		})
	}

	w, err := workspace.New(workspace.WorkspaceParams{
		Name:         request.Name,
		Cwd:          c.DefaultWorkspaceDir,
		Credentials:  credClient,
		Extensions:   extensions,
		Repositories: repositories,
	})
	if err != nil {
		log.Error(err)
		return err
	}

	log.Debug(w)
	workspace.SaveToDB(w)

	unsubscribe := make(chan bool, 1)

	go func() {
		for event := range event_bus.SubscribeWithFilter(unsubscribe, func(i event_bus.Event) bool {
			if _, ok := i.Payload.(event_bus.WorkspaceEventPayload); ok {
				return i.Payload.(event_bus.WorkspaceEventPayload).WorkspaceName == w.Name
			}

			if _, ok := i.Payload.(event_bus.ProjectEventPayload); ok {
				return i.Payload.(event_bus.ProjectEventPayload).WorkspaceName == w.Name
			}

			return false
		}) {
			log.Debug(event)
			jsonPayload, err := json.Marshal(event.Payload)
			if err != nil {
				log.Error(err)
				return
			}

			err = stream.Send(&daytona_proto.CreateWorkspaceResponse{
				Event:   string(event.Name),
				Payload: string(jsonPayload),
			})
			if err != nil {
				log.Error("Event send error")
				log.Error(err)
			}
		}
	}()

	err = w.Create()
	if err != nil {
		log.Error(err)
		stream.Send(&daytona_proto.CreateWorkspaceResponse{
			Event:   "error",
			Payload: err.Error(),
		})
		return err
	}
	err = w.Start()
	if err != nil {
		log.Error(err)
		stream.Send(&daytona_proto.CreateWorkspaceResponse{
			Event:   "error",
			Payload: err.Error(),
		})
		return err
	}
	time.Sleep(100 * time.Millisecond)

	unsubscribe <- true
	return nil
}
