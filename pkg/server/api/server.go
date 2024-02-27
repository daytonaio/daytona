//	@title			Daytona Server API
//	@version		0.1.0
//	@description	Daytona Server API

//	@host		localhost:3000
//	@schemes	http
//	@BasePath	/

package api

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/daytonaio/daytona/pkg/server/api/docs"
	"github.com/daytonaio/daytona/pkg/server/api/middlewares"
	"github.com/gin-contrib/cors"

	log_controller "github.com/daytonaio/daytona/pkg/server/api/controllers/log"
	"github.com/daytonaio/daytona/pkg/server/api/controllers/provider"
	"github.com/daytonaio/daytona/pkg/server/api/controllers/server"
	"github.com/daytonaio/daytona/pkg/server/api/controllers/workspace"
	"github.com/daytonaio/daytona/pkg/server/config"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var httpServer *http.Server
var router *gin.Engine

func Start() error {
	docs.SwaggerInfo.Version = "0.1"
	docs.SwaggerInfo.BasePath = "/"
	docs.SwaggerInfo.Description = "Daytona Server API"
	docs.SwaggerInfo.Title = "Daytona Server API"

	if mode, ok := os.LookupEnv("DAYTONA_SERVER_MODE"); ok && mode == "development" {
		router = gin.Default()
		router.Use(cors.New(cors.Config{
			AllowAllOrigins: true,
		}))
	} else {
		gin.SetMode(gin.ReleaseMode)
		router = gin.New()
		router.Use(gin.Recovery())
	}
	router.Use(middlewares.LoggingMiddleware())

	config, err := config.GetConfig()
	if err != nil {
		return err
	}

	serverController := router.Group("/server")
	{
		serverController.GET("/config", server.GetConfig)
		serverController.POST("/config", server.SetConfig)
		serverController.POST("/network-key", server.GenerateNetworkKey)
	}

	workspaceController := router.Group("/workspace")
	{
		workspaceController.GET("/:workspaceId", workspace.GetWorkspaceInfo)
		workspaceController.GET("/", workspace.ListWorkspaces)
		workspaceController.POST("/", workspace.CreateWorkspace)
		workspaceController.POST("/:workspaceId/start", workspace.StartWorkspace)
		workspaceController.POST("/:workspaceId/stop", workspace.StopWorkspace)
		workspaceController.DELETE("/:workspaceId", workspace.RemoveWorkspace)
		workspaceController.POST("/:workspaceId/:projectId/start", workspace.StartProject)
		workspaceController.POST("/:workspaceId/:projectId/stop", workspace.StopProject)
	}

	providerController := router.Group("/provider")
	{
		providerController.POST("/install", provider.InstallProvider)
		providerController.GET("/", provider.ListProviders)
		providerController.POST("/:provider/uninstall", provider.UninstallProvider)
	}

	logController := router.Group("/log")
	{
		logController.GET("/ws", log_controller.ReadServerLog)
	}

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", config.ApiPort),
		Handler: router,
	}

	log.Infof("Starting api server on port %d", config.ApiPort)

	return httpServer.ListenAndServe()
}

func Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Error(err)
	}
}
