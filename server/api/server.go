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
	"time"

	"github.com/daytonaio/daytona/server/api/docs"

	log_controller "github.com/daytonaio/daytona/server/api/controllers/log"
	"github.com/daytonaio/daytona/server/api/controllers/plugin"
	"github.com/daytonaio/daytona/server/api/controllers/server"
	"github.com/daytonaio/daytona/server/api/controllers/workspace"
	"github.com/daytonaio/daytona/server/config"

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

	router = gin.Default()

	// if BaseConfig.Production {
	// 	gin.SetMode(gin.ReleaseMode)
	// 	router = gin.New()
	// 	router.Use(gin.Recovery())
	// } else {
	// 	router = gin.Default()
	// 	router.Use(cors.New(cors.Config{
	// 		AllowAllOrigins: true,
	// 	}))
	// }

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

	pluginController := router.Group("/plugin")
	{
		pluginController.POST("/provisioner/install", plugin.InstallProvisionerPlugin)
		pluginController.POST("/agent-service/install", plugin.InstallAgentServicePlugin)
		pluginController.GET("/provisioner", plugin.ListProvisionerPlugins)
		pluginController.GET("/agent-service", plugin.ListAgentServicePlugins)
		pluginController.POST("/provisioner/:provisioner/uninstall", plugin.UninstallProvisionerPlugin)
		pluginController.POST("/agent-service/:agent-service/uninstall", plugin.UninstallAgentServicePlugin)
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
