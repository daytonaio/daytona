package workspace

import (
	"errors"
	"fmt"
	"net/http"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/provider/manager"
	"github.com/daytonaio/daytona/pkg/server/api/controllers/workspace/dto"
	"github.com/daytonaio/daytona/pkg/server/config"
	"github.com/daytonaio/daytona/pkg/server/db"
	"github.com/daytonaio/daytona/pkg/server/provisioner"
	"github.com/daytonaio/daytona/pkg/types"
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
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %s", err.Error()))
		return
	}

	_, err = db.FindWorkspace(createWorkspaceDto.Name)
	if err == nil {
		ctx.AbortWithError(http.StatusConflict, errors.New("workspace already exists"))
		return
	}

	w, err := newWorkspace(createWorkspaceDto)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to initialize workspace: %s", err.Error()))
		return
	}

	log.Debug(w)
	db.SaveWorkspace(w)

	err = provisioner.CreateWorkspace(w)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to create workspace: %s", err.Error()))
		return
	}
	err = provisioner.StartWorkspace(w)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to start workspace: %s", err.Error()))
		return
	}
	time.Sleep(100 * time.Millisecond)

	ctx.JSON(200, w)
}

func newWorkspace(createWorkspaceDto dto.CreateWorkspace) (*types.Workspace, error) {
	fmt.Println(createWorkspaceDto.Name)
	fmt.Println(createWorkspaceDto.Name)
	fmt.Println(createWorkspaceDto.Name)
	fmt.Println(createWorkspaceDto.Name)
	fmt.Println(createWorkspaceDto.Name)
	fmt.Println(createWorkspaceDto.Name)
	fmt.Println(createWorkspaceDto.Name)
	fmt.Println(createWorkspaceDto.Name)
	fmt.Println(createWorkspaceDto.Name)
	fmt.Println(createWorkspaceDto.Name)
	fmt.Println(createWorkspaceDto.Name)
	isAlphaNumeric := regexp.MustCompile(`^[a-zA-Z0-9-]+$`).MatchString
	if !isAlphaNumeric(createWorkspaceDto.Name) {
		return nil, errors.New("name is not a valid alphanumeric string")
	}

	_, err := manager.GetProvider(createWorkspaceDto.Provider)
	if err != nil {
		return nil, err
	}

	w := &types.Workspace{
		Id:   createWorkspaceDto.Name,
		Name: createWorkspaceDto.Name,
		Provider: &types.WorkspaceProvider{
			Name: createWorkspaceDto.Provider,
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
		gitProvider := gitprovider.GetGitProvider(providerId, userGitProviders)

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
