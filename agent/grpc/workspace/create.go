// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace_grpc

import (
	"encoding/json"
	"errors"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/daytonaio/daytona/agent/db"
	"github.com/daytonaio/daytona/agent/event_bus"
	"github.com/daytonaio/daytona/agent/provisioner"
	daytona_proto "github.com/daytonaio/daytona/grpc/proto"
	"github.com/daytonaio/daytona/grpc/proto/types"

	log "github.com/sirupsen/logrus"
)

func (m *WorkspaceServer) Create(request *daytona_proto.CreateWorkspaceRequest, stream daytona_proto.WorkspaceService_CreateServer) error {
	_, err := db.FindWorkspace(request.Name)
	if err == nil {
		return errors.New("workspace already exists")
	}

	// c, err := config.GetConfig()
	// if err != nil {
	// 	return err
	// }

	// credClient := &credentials.CredentialsClient{}

	// extensions := []workspace.Extension{}

	// vsc_server_extension := vsc_server.VscServerExtension{}
	// extensions = append(extensions, vsc_server_extension)

	// sshPublicKey, err := config.GetWorkspacePublicKey()
	// if err != nil {
	// 	log.Error(err)
	// 	return err
	// }

	// ssh := ssh.SshExtension{
	// 	PublicKey: sshPublicKey,
	// }
	// extensions = append(extensions, ssh)

	w, err := newWorkspace(request)
	if err != nil {
		log.Error(err)
		return err
	}

	log.Debug(w)
	db.SaveWorkspace(w)

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

	err = provisioner.CreateWorkspace(w)
	if err != nil {
		log.Error(err)
		stream.Send(&daytona_proto.CreateWorkspaceResponse{
			Event:   "error",
			Payload: err.Error(),
		})
		return err
	}
	err = provisioner.StartWorkspace(w)
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

func newWorkspace(params *daytona_proto.CreateWorkspaceRequest) (*types.Workspace, error) {
	isAlphaNumeric := regexp.MustCompile(`^[a-zA-Z0-9-]+$`).MatchString
	if !isAlphaNumeric(params.Name) {
		return nil, errors.New("name is not a valid alphanumeric string")
	}

	w := &types.Workspace{
		Id:   params.Name,
		Name: params.Name,
		Provisioner: &types.WorkspaceProvisioner{
			Name:    "docker-provisioner",
			Profile: "default",
		},
	}

	w.Projects = []*types.Project{}

	for _, repo := range params.Repositories {
		project := &types.Project{
			Name: strings.ToLower(path.Base(repo)),
			Repository: &types.Repository{
				Url: repo,
			},
			WorkspaceId: w.Id,
		}
		w.Projects = append(w.Projects, project)
	}

	return w, nil
}
