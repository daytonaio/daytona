package workspace

import (
	"encoding/json"
	"errors"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/daytonaio/daytona/common/types"
	provisioner_manager "github.com/daytonaio/daytona/plugins/provisioner/manager"
	workspace_dto "github.com/daytonaio/daytona/server/api/controllers/workspace/dto"
	"github.com/daytonaio/daytona/server/config"
	"github.com/daytonaio/daytona/server/db"
	"github.com/daytonaio/daytona/server/event_bus"
	"github.com/daytonaio/daytona/server/headscale"
	"github.com/daytonaio/daytona/server/provisioner"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// CreateWorkspace 			godoc
//
//	@Tags			workspace
//	@Summary		Create a workspace
//	@Description	Create a workspace
//	@Produce		json
//	@Success		200	{object}	types.ServerConfig
//	@Router			/workspace/create [post]
//
//	@id				CreateWorkspace
func CreateWorkspace(ctx *gin.Context) {
	var createWorkspaceDto workspace_dto.CreateWorkspaceDTO
	err := ctx.BindJSON(&createWorkspaceDto)

	_, err = db.FindWorkspace(createWorkspaceDto.Name)
	if err != nil {
		ctx.JSON(500, gin.H{"workspace already exists": err.Error()})
		return
	}

	w, err := newWorkspace(createWorkspaceDto)
	if err != nil {
		ctx.JSON(500, gin.H{"err": err.Error()})
		return
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
				ctx.JSON(500, gin.H{"err": err.Error()})
				return
			}

			err = stream.Send(&workspace_dto.WorkspaceCreationDTO{
				Event:   string(event.Name),
				Payload: string(jsonPayload),
			})
			if err != nil {
				ctx.JSON(500, gin.H{"Event send error": err.Error()})
			}
		}
	}()

	err = provisioner.CreateWorkspace(w)
	if err != nil {
		log.Error(err)
		stream.Send(&workspace_dto.WorkspaceCreationDTO{
			Event:   "error",
			Payload: err.Error(),
		})
		ctx.JSON(500, gin.H{"err": err.Error()})
	}
	err = provisioner.StartWorkspace(w)
	if err != nil {
		log.Error(err)
		stream.Send(&workspace_dto.WorkspaceCreationDTO{
			Event:   "error",
			Payload: err.Error(),
		})
		ctx.JSON(500, gin.H{"err": err.Error()})
	}
	time.Sleep(100 * time.Millisecond)

	unsubscribe <- true

	// ctx.JSON(200, )
	// return nil
	ctx.JSON(200, nil)
}

func newWorkspace(createWorkspaceDto workspace_dto.CreateWorkspaceDTO) (*types.Workspace, error) {
	isAlphaNumeric := regexp.MustCompile(`^[a-zA-Z0-9-]+$`).MatchString
	if !isAlphaNumeric(createWorkspaceDto.Name) {
		return nil, errors.New("name is not a valid alphanumeric string")
	}

	_, err := provisioner_manager.GetProvisioner(createWorkspaceDto.Provisioner)
	if err != nil {
		return nil, err
	}

	w := &types.Workspace{
		Id:   createWorkspaceDto.Name,
		Name: createWorkspaceDto.Name,
		Provisioner: &types.WorkspaceProvisioner{
			Name: createWorkspaceDto.Provisioner,
			// TODO: Add profile support
			Profile: "default",
		},
	}

	w.Projects = []*types.Project{}

	for _, repo := range createWorkspaceDto.Repositories {
		authKey, err := headscale.CreateAuthKey()
		if err != nil {
			return nil, err
		}

		project := &types.Project{
			Name: strings.ToLower(path.Base(repo)),
			Repository: &types.Repository{
				Url: repo,
			},
			WorkspaceId: w.Id,
			AuthKey:     authKey,
		}
		w.Projects = append(w.Projects, project)
	}

	return w, nil
}

// SetConfig 			godoc
//
//	@Tags			server
//	@Summary		Set the server configuration
//	@Description	Set the server configuration
//	@Accept			json
//	@Produce		json
//	@Param			config	body		types.ServerConfig	true	"Server configuration"
//	@Success		200		{object}	types.ServerConfig
//	@Router			/server/config [post]
//
//	@id				SetConfig
func SetConfig(ctx *gin.Context) {
	var c types.ServerConfig
	err := ctx.BindJSON(&c)
	if err != nil {
		ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	err = config.Save(&c)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(200, c)
}
