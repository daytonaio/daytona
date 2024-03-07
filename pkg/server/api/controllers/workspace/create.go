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
	"github.com/daytonaio/daytona/pkg/server/api/controllers/workspace/dto"
	"github.com/daytonaio/daytona/pkg/server/config"
	"github.com/daytonaio/daytona/pkg/server/db"
	"github.com/daytonaio/daytona/pkg/server/provisioner"
	"github.com/daytonaio/daytona/pkg/server/targets"
	"github.com/daytonaio/daytona/pkg/types"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

	_, err = db.FindWorkspaceByName(createWorkspaceDto.Name)
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
	isAlphaNumeric := regexp.MustCompile(`^[a-zA-Z0-9-]+$`).MatchString
	if !isAlphaNumeric(createWorkspaceDto.Name) {
		return nil, errors.New("name is not a valid alphanumeric string")
	}

	_, err := targets.GetTarget(createWorkspaceDto.Target)
	if err != nil {
		return nil, err
	}

	w := &types.Workspace{
		Id:     uuid.NewString(),
		Name:   createWorkspaceDto.Name,
		Target: createWorkspaceDto.Target,
	}

	w.Projects = []*types.Project{}
	serverConfig, err := config.GetConfig()
	if err != nil {
		log.Fatal(err)
	}
	userGitProviders := serverConfig.GitProviders

	for i := range createWorkspaceDto.Repositories {
		repo := createWorkspaceDto.Repositories[i]
		providerId := getGitProviderIdFromUrl(repo.Url)
		gitProvider := gitprovider.GetGitProvider(providerId, userGitProviders)

		if gitProvider != nil {
			gitUser, err := gitProvider.GetUserData()
			if err != nil {
				return nil, err
			}
			repo.GitUserData = &types.GitUserData{
				Name:  gitUser.Name,
				Email: gitUser.Email,
			}
		}

		// TODO: generate API key for project
		projectNameSlugRegex := regexp.MustCompile(`[^a-zA-Z0-9-]`)
		projectName := projectNameSlugRegex.ReplaceAllString(strings.ToLower(path.Base(repo.Url)), "-")
		project := &types.Project{
			Name:        projectName,
			Repository:  &repo,
			WorkspaceId: w.Id,
			ApiKey:      "TODO",
			Target:      createWorkspaceDto.Target,
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
