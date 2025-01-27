// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

//	@title			Daytona Server API
//	@version		v0.0.0-dev
//	@description	Daytona Server API

//	@host		localhost:3986
//	@schemes	http
//	@BasePath	/

//	@securityDefinitions.apikey	Bearer
//	@in							header
//	@name						Authorization
//	@description				"Type 'Bearer TOKEN' to correctly set the API Key"

//	@Security	Bearer

package api

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/daytonaio/daytona/pkg/api/docs"
	"github.com/daytonaio/daytona/pkg/api/middlewares"
	"github.com/daytonaio/daytona/pkg/frpc"
	"github.com/daytonaio/daytona/pkg/telemetry"
	"github.com/gin-contrib/cors"

	"github.com/daytonaio/daytona/pkg/api/controllers/apikey"
	"github.com/daytonaio/daytona/pkg/api/controllers/binary"
	"github.com/daytonaio/daytona/pkg/api/controllers/build"
	"github.com/daytonaio/daytona/pkg/api/controllers/containerregistry"
	"github.com/daytonaio/daytona/pkg/api/controllers/env"
	"github.com/daytonaio/daytona/pkg/api/controllers/gitprovider"
	"github.com/daytonaio/daytona/pkg/api/controllers/health"
	"github.com/daytonaio/daytona/pkg/api/controllers/job"
	log_controller "github.com/daytonaio/daytona/pkg/api/controllers/log"
	"github.com/daytonaio/daytona/pkg/api/controllers/runner"
	"github.com/daytonaio/daytona/pkg/api/controllers/runner/provider"
	"github.com/daytonaio/daytona/pkg/api/controllers/sample"
	"github.com/daytonaio/daytona/pkg/api/controllers/server"
	"github.com/daytonaio/daytona/pkg/api/controllers/target"
	"github.com/daytonaio/daytona/pkg/api/controllers/targetconfig"
	"github.com/daytonaio/daytona/pkg/api/controllers/workspace"
	"github.com/daytonaio/daytona/pkg/api/controllers/workspace/toolbox"
	"github.com/daytonaio/daytona/pkg/api/controllers/workspacetemplate"
	"github.com/daytonaio/daytona/pkg/api/controllers/workspacetemplate/prebuild"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	log "github.com/sirupsen/logrus"

	"github.com/daytonaio/daytona/internal/constants"
	daytonaServer "github.com/daytonaio/daytona/pkg/server"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type ApiServerConfig struct {
	ApiPort          int
	Version          string
	TelemetryService telemetry.TelemetryService
	Frps             *daytonaServer.FRPSConfig
	ServerId         string
}

func NewApiServer(config ApiServerConfig) *ApiServer {
	return &ApiServer{
		apiPort:          config.ApiPort,
		telemetryService: config.TelemetryService,
		version:          config.Version,
		frps:             config.Frps,
		serverId:         config.ServerId,
	}
}

type ApiServer struct {
	apiPort          int
	telemetryService telemetry.TelemetryService
	httpServer       *http.Server
	router           *gin.Engine
	version          string
	frps             *daytonaServer.FRPSConfig
	serverId         string
}

func (a *ApiServer) Start() error {
	docs.SwaggerInfo.Version = a.version
	docs.SwaggerInfo.BasePath = "/"
	docs.SwaggerInfo.Description = "Daytona Server API"
	docs.SwaggerInfo.Title = "Daytona Server API"

	_, err := net.Dial("tcp", fmt.Sprintf(":%d", a.apiPort))
	if err == nil {
		return fmt.Errorf("cannot start API server, port %d is already in use", a.apiPort)
	}

	binding.Validator = new(DefaultValidator)

	gin.SetMode(gin.ReleaseMode)

	a.router = gin.New()
	a.router.Use(gin.Recovery())
	if mode, ok := os.LookupEnv("DAYTONA_SERVER_MODE"); ok && mode == "development" {
		a.router.Use(cors.New(cors.Config{
			AllowAllOrigins: true,
		}))
		gin.SetMode(gin.DebugMode)
	}

	a.router.Use(middlewares.TelemetryMiddleware(a.telemetryService))
	a.router.Use(middlewares.LoggingMiddleware())
	a.router.Use(middlewares.SetVersionMiddleware(a.version))

	public := a.router.Group("/")
	public.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	healthController := public.Group(constants.HEALTH_CHECK_ROUTE)
	{
		healthController.GET("", health.HealthCheck)
	}

	protected := a.router.Group("/")
	protected.Use(middlewares.AuthMiddleware())

	serverController := protected.Group("/server")
	{
		serverController.GET("/config", server.GetConfig)
		serverController.PUT("/config", server.SaveConfig)
		serverController.POST("/network-key", server.CreateNetworkKey)
		serverController.GET("/logs", server.GetServerLogFiles)
	}

	binaryController := protected.Group("/binary")
	{
		binaryController.GET("/script", binary.GetDaytonaScript)
		binaryController.GET("/:version/:binaryName", binary.GetBinary)
	}

	targetController := protected.Group("/target")
	{
		targetController.GET("/:targetId", target.FindTarget)
		targetController.GET("/:targetId/state", target.GetTargetState)
		targetController.GET("", target.ListTargets)
		targetController.POST("", target.CreateTarget)
		targetController.POST("/:targetId/start", target.StartTarget)
		targetController.POST("/:targetId/stop", target.StopTarget)
		targetController.POST("/:targetId/restart", target.RestartTarget)
		targetController.PATCH("/:targetId/set-default", target.SetDefaultTarget)
		targetController.POST("/:targetId/handle-successful-creation", target.HandleSuccessfulCreation)
		targetController.POST("/:targetId/provider-metadata", target.UpdateTargetProviderMetadata)
		targetController.DELETE("/:targetId", target.DeleteTarget)
	}

	workspaceController := protected.Group("/workspace")
	{
		toolboxController := workspaceController.Group("/:workspaceId/toolbox")
		{
			toolboxController.GET("/workspace-dir", toolbox.GetWorkspaceDir)

			processController := toolboxController.Group("/process")
			{
				processController.POST("/execute", toolbox.ProcessExecuteCommand)

				sessionController := processController.Group("/session")
				{
					sessionController.GET("", toolbox.ListSessions)
					sessionController.POST("", toolbox.CreateSession)
					sessionController.POST("/:sessionId/exec", toolbox.SessionExecuteCommand)
					sessionController.DELETE("/:sessionId", toolbox.DeleteSession)
					sessionController.GET("/:sessionId/command/:commandId/logs", toolbox.GetSessionCommandLogs)
				}
			}

			fsController := toolboxController.Group("/files")
			{
				fsController.DELETE("", toolbox.FsDeleteFile)

				fsController.GET("", toolbox.FsListFiles)
				fsController.GET("/download", toolbox.FsDownloadFile)
				fsController.GET("/find", toolbox.FsFindInFiles)
				fsController.GET("/info", toolbox.FsGetFileDetails)
				fsController.GET("/search", toolbox.FsSearchFiles)

				fsController.POST("/folder", toolbox.FsCreateFolder)
				fsController.POST("/move", toolbox.FsMoveFile)
				fsController.POST("/replace", toolbox.FsReplaceInFiles)
				fsController.POST("/permissions", toolbox.FsSetFilePermissions)
				fsController.POST("/upload", toolbox.FsUploadFile)
			}

			gitController := toolboxController.Group("/git")
			{
				gitController.GET("/branches", toolbox.GitBranchList)
				gitController.GET("/history", toolbox.GitCommitHistory)
				gitController.GET("/status", toolbox.GitStatus)

				gitController.POST("/add", toolbox.GitAddFiles)
				gitController.POST("/branches", toolbox.GitCreateBranch)
				gitController.POST("/clone", toolbox.GitCloneRepository)
				gitController.POST("/commit", toolbox.GitCommitChanges)
				gitController.POST("/pull", toolbox.GitPushChanges)
				gitController.POST("/push", toolbox.GitPushChanges)
			}

			lspController := toolboxController.Group("/lsp")
			{
				lspController.GET("/document-symbols", toolbox.LspDocumentSymbols)
				lspController.GET("/workspacesymbols", toolbox.LspWorkspaceSymbols)

				lspController.POST("/completions", toolbox.LspCompletions)
				lspController.POST("/did-close", toolbox.LspDidClose)
				lspController.POST("/did-open", toolbox.LspDidOpen)
				lspController.POST("/start", toolbox.LspStart)
				lspController.POST("/stop", toolbox.LspStop)
			}
		}

		workspaceController.GET("/:workspaceId", workspace.FindWorkspace)
		workspaceController.GET("/:workspaceId/state", workspace.GetWorkspaceState)
		workspaceController.GET("", workspace.ListWorkspaces)
		workspaceController.POST("", workspace.CreateWorkspace)
		workspaceController.DELETE("/:workspaceId", workspace.DeleteWorkspace)

		workspaceController.POST("/:workspaceId/start", workspace.StartWorkspace)
		workspaceController.POST("/:workspaceId/stop", workspace.StopWorkspace)
		workspaceController.POST("/:workspaceId/restart", workspace.RestartWorkspace)
		workspaceController.POST("/:workspaceId/provider-metadata", workspace.UpdateWorkspaceProviderMetadata)
		workspaceController.POST("/:workspaceId/labels", workspace.UpdateWorkspaceLabels)
	}

	workspaceTemplateController := protected.Group("/workspace-template")
	{
		// Defining the prebuild routes first to avoid conflicts with the workspace template routes
		prebuildRoutePath := "/prebuild"
		workspaceTemplatePrebuildsGroup := workspaceTemplateController.Group(prebuildRoutePath)
		{
			workspaceTemplatePrebuildsGroup.GET("", prebuild.ListPrebuilds)
		}

		workspaceTemplateNameGroup := workspaceTemplateController.Group(":templateName")
		{
			workspaceTemplateNameGroup.PUT(prebuildRoutePath, prebuild.SavePrebuild)
			workspaceTemplateNameGroup.GET(prebuildRoutePath, prebuild.ListPrebuildsForWorkspaceTemplate)
			workspaceTemplateNameGroup.GET(prebuildRoutePath+"/:prebuildId", prebuild.FindPrebuild)
			workspaceTemplateNameGroup.DELETE(prebuildRoutePath+"/:prebuildId", prebuild.DeletePrebuild)

			workspaceTemplateNameGroup.GET("", workspacetemplate.FindWorkspaceTemplate)
			workspaceTemplateNameGroup.PATCH("/set-default", workspacetemplate.SetDefaultWorkspaceTemplate)
			workspaceTemplateNameGroup.DELETE("", workspacetemplate.DeleteWorkspaceTemplate)
		}

		workspaceTemplateController.GET("", workspacetemplate.ListWorkspaceTemplates)
		workspaceTemplateController.PUT("", workspacetemplate.SaveWorkspaceTemplate)
		workspaceTemplateController.GET("/default/:gitUrl", workspacetemplate.GetDefaultWorkspaceTemplate)
	}

	public.POST(constants.WEBHOOK_EVENT_ROUTE, prebuild.ProcessGitEvent)

	buildController := protected.Group("/build")
	{
		buildController.POST("", build.CreateBuild)
		buildController.GET("/:buildId", build.FindBuild)
		buildController.GET("", build.ListBuilds)
		buildController.GET("/successful/:repoUrl", build.ListSuccessfulBuilds)
		buildController.DELETE("", build.DeleteAllBuilds)
		buildController.DELETE("/:buildId", build.DeleteBuild)
		buildController.DELETE("/prebuild/:prebuildId", build.DeleteBuildsFromPrebuild)
	}

	targetConfigController := protected.Group("/target-config")
	{
		targetConfigController.GET("", targetconfig.ListTargetConfigs)
		targetConfigController.POST("", targetconfig.CreateTargetConfig)
		targetConfigController.DELETE("/:configId", targetconfig.DeleteTargetConfig)
	}

	logController := protected.Group("/log")
	{
		logController.GET("/server", log_controller.ReadServerLog)

		logController.GET("/target/:targetId", log_controller.ReadTargetLog)
		logController.GET("/target/:targetId/write", log_controller.WriteTargetLog)

		logController.GET("/workspace/:workspaceId", log_controller.ReadWorkspaceLog)
		logController.GET("/workspace/:workspaceId/write", log_controller.WriteWorkspaceLog)

		logController.GET("/build/:buildId", log_controller.ReadBuildLog)
		logController.GET("/build/:buildId/write", log_controller.WriteBuildLog)

		logController.GET("/runner/:runnerId", log_controller.ReadRunnerLog)
		logController.GET("/runner/:runnerId/write", log_controller.WriteRunnerLog)
	}

	gitProviderController := protected.Group("/gitprovider")
	{
		gitProviderController.GET("", gitprovider.ListGitProviders)
		gitProviderController.PUT("", gitprovider.SaveGitProvider)
		gitProviderController.DELETE("/:gitProviderId", gitprovider.DeleteGitProvider)
		gitProviderController.GET("/for-url/:url", gitprovider.ListGitProvidersForUrl)
		gitProviderController.GET("/id-for-url/:url", gitprovider.FindGitProviderIdForUrl)
		gitProviderController.GET("/:gitProviderId", gitprovider.FindGitProvider)

		gitProviderController.GET("/:gitProviderId/user", gitprovider.GetGitUser)
		gitProviderController.GET("/:gitProviderId/namespaces", gitprovider.GetNamespaces)
		gitProviderController.GET("/:gitProviderId/:namespaceId/repositories", gitprovider.GetRepositories)
		gitProviderController.GET("/:gitProviderId/:namespaceId/:repositoryId/branches", gitprovider.GetRepoBranches)
		gitProviderController.GET("/:gitProviderId/:namespaceId/:repositoryId/pull-requests", gitprovider.GetRepoPRs)
		gitProviderController.POST("/context", gitprovider.GetGitContext)
		gitProviderController.POST("/context/url", gitprovider.GetUrlFromRepository)
	}

	apiKeyController := protected.Group("/apikey")
	{
		apiKeyController.GET("", apikey.ListClientApiKeys)
		apiKeyController.POST("/:apiKeyName", apikey.CreateApiKey)
		apiKeyController.DELETE("/:apiKeyName", apikey.DeleteApiKey)
	}

	envVarController := protected.Group("/env")
	{
		envVarController.GET("", env.ListEnvironmentVariables)
		envVarController.PUT("", env.SaveEnvironmentVariable)
		envVarController.DELETE("/:key", env.DeleteEnvironmentVariable)
	}

	containerRegistryController := protected.Group("/container-registry")
	{
		containerRegistryController.GET("/:server", containerregistry.FindContainerRegistry)
	}

	jobController := protected.Group("/job")
	{
		jobController.GET("", job.ListJobs)
	}

	samplesController := protected.Group("/sample")
	{
		samplesController.GET("", sample.ListSamples)
	}

	runnerController := protected.Group("/runner")
	{
		// Defining the provider routes first to avoid conflicts with the runner routes
		providerRoutePath := "/provider"
		providersGroup := runnerController.Group(providerRoutePath)
		{
			providersGroup.GET("", provider.ListProviders)
			providersGroup.GET("/for-install", provider.ListProvidersForInstall)
		}

		runnerIdGroup := runnerController.Group(":runnerId")
		{
			runnerIdProviderGroup := runnerIdGroup.Group(providerRoutePath)
			{
				runnerIdProviderGroup.GET("", provider.GetRunnerProviders)
				runnerIdProviderGroup.POST("/:providerName/install", provider.InstallProvider)
				runnerIdProviderGroup.POST("/:providerName/uninstall", provider.UninstallProvider)
				runnerIdProviderGroup.POST("/:providerName/update", provider.UpdateProvider)
			}

			runnerIdGroup.GET("", runner.FindRunner)
			runnerIdGroup.DELETE("", runner.DeleteRunner)
		}

		runnerController.GET("", runner.ListRunners)
		runnerController.POST("", runner.CreateRunner)
	}

	workspaceGroup := protected.Group("/")
	workspaceGroup.Use(middlewares.WorkspaceAuthMiddleware())
	{
		workspaceGroup.POST(workspaceController.BasePath()+"/:workspaceId/metadata", workspace.UpdateWorkspaceMetadata)
		workspaceGroup.POST(targetController.BasePath()+"/:targetId/metadata", target.UpdateTargetMetadata)
	}

	runnerGroup := protected.Group("/")
	runnerGroup.Use(middlewares.RunnerAuthMiddleware())
	{
		runnerGroup.POST(runnerController.BasePath()+"/:runnerId/metadata", runner.UpdateRunnerMetadata)
		runnerGroup.GET(runnerController.BasePath()+"/:runnerId/jobs", runner.ListRunnerJobs)
		runnerGroup.POST(runnerController.BasePath()+"/:runnerId/jobs/:jobId/state", runner.UpdateJobState)
	}

	a.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", a.apiPort),
		Handler: a.router,
	}

	listener, err := net.Listen("tcp", a.httpServer.Addr)
	if err != nil {
		return err
	}

	errChan := make(chan error)
	go func() {
		errChan <- a.httpServer.Serve(listener)
	}()

	if a.frps == nil {
		return <-errChan
	}

	frpcHealthCheck, frpcService, err := frpc.GetService(frpc.FrpcConnectParams{
		ServerDomain: a.frps.Domain,
		ServerPort:   int(a.frps.Port),
		Name:         fmt.Sprintf("daytona-server-api-%s", a.serverId),
		Port:         int(a.apiPort),
		SubDomain:    fmt.Sprintf("api-%s", a.serverId),
	})
	if err != nil {
		return err
	}

	go func() {
		err := frpcService.Run(context.Background())
		if err != nil {
			errChan <- err
		}
	}()

	for i := 0; i < 5; i++ {
		if err = frpcHealthCheck(); err != nil {
			log.Debugf("Failed to connect to api frpc: %s", err)
			time.Sleep(2 * time.Second)
		} else {
			break
		}
	}
	if err != nil {
		return err
	}

	return <-errChan
}

func (a *ApiServer) HealthCheck() error {
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d%s", a.apiPort, constants.HEALTH_CHECK_ROUTE))
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status code %d", resp.StatusCode)
	}

	return nil
}

func (a *ApiServer) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := a.httpServer.Shutdown(ctx); err != nil {
		log.Error(err)
	}
}
