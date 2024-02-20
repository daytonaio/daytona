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

	"github.com/daytonaio/daytona/server/api/controllers/server"
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
