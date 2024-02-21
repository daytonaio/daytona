package workspace

import (
	"errors"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/daytonaio/daytona/common/types"
	provisioner_manager "github.com/daytonaio/daytona/plugins/provisioner/manager"
	"github.com/daytonaio/daytona/server/api/controllers/workspace/dto"
	"github.com/daytonaio/daytona/server/db"
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
//	@Param			workspace	body	CreateWorkspace	true	"Create workspace"
//	@Produce		json
//	@Success		200	{object}	Workspace
//	@Router			/workspace [post]
//
//	@id				CreateWorkspace
func CreateWorkspace(ctx *gin.Context) {
	var createWorkspaceDto dto.CreateWorkspace
	err := ctx.BindJSON(&createWorkspaceDto)
	if err != nil {
		log.Error(err)
		ctx.JSON(400, gin.H{"err": err.Error()})
		return
	}

	_, err = db.FindWorkspace(createWorkspaceDto.Name)
	if err == nil {
		ctx.JSON(400, gin.H{"err": "workspace already exists"})
		return
	}

	w, err := newWorkspace(createWorkspaceDto)
	if err != nil {
		log.Error(err)
		ctx.JSON(500, gin.H{"err": err.Error()})
		return
	}

	log.Debug(w)
	db.SaveWorkspace(w)

	err = provisioner.CreateWorkspace(w)
	if err != nil {
		log.Error(err)
		ctx.JSON(500, gin.H{"err": err.Error()})
		return
	}
	err = provisioner.StartWorkspace(w)
	if err != nil {
		log.Error(err)
		ctx.JSON(500, gin.H{"err": err.Error()})
		return
	}
	time.Sleep(100 * time.Millisecond)

	ctx.JSON(200, w)
}

func newWorkspace(createWorkspaceDto dto.CreateWorkspace) (*types.Workspace, error) {
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
