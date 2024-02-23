package workspace

import (
	"errors"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/daytonaio/daytona/common/types"
	"github.com/daytonaio/daytona/pkg/git_provider"
	provisioner_manager "github.com/daytonaio/daytona/plugins/provisioner/manager"
	"github.com/daytonaio/daytona/server/api/controllers/workspace/dto"
	"github.com/daytonaio/daytona/server/config"
	"github.com/daytonaio/daytona/server/db"
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
	serverConfig, err := config.GetConfig()
	if err != nil {
		log.Fatal(err)
	}
	userGitProviders := serverConfig.GitProviders

	for _, repo := range createWorkspaceDto.Repositories {
		var gitUserData *types.GitUserData
		providerId := getGitProviderIdFromUrl(repo)

		var providerCredentialsExist bool
		for _, gitProvider := range userGitProviders {
			if gitProvider.Id == providerId {
				providerCredentialsExist = true
			}
		}

		if providerCredentialsExist {
			gitProvider, err := git_provider.GetGitProviderServer(providerId, userGitProviders)
			if err != nil {
				return nil, err
			}
			if gitProvider != nil {
				gitUser, err := gitProvider.GetUserData()
				if err != nil {
					return nil, err
				}
				gitUserData = &types.GitUserData{
					Name:  gitUser.Name,
					Email: gitUser.Email,
				}
			}
		}

		// TODO: generate API key for project
		project := &types.Project{
			Name: strings.ToLower(path.Base(repo)),
			Repository: &types.Repository{
				Url:         repo,
				GitUserData: gitUserData,
			},
			WorkspaceId: w.Id,
			ApiKey:      "TODO",
		}
		w.Projects = append(w.Projects, project)
	}

	return w, nil
}

func getGitProviderIdFromUrl(url string) string {
	if strings.Contains(url, "github.com") {
		return "github"
	} else if strings.Contains(url, "gitlab.com") {
		return "gitlab"
	} else if strings.Contains(url, "bitbucket.org") {
		return "bitbucket"
	} else {
		return ""
	}
}
