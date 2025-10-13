// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package proxy

import (
	"fmt"
	"net"
	"net/http"

	apiclient "github.com/daytonaio/apiclient"
	"github.com/daytonaio/otel-proxy/cmd/otel-proxy/config"
	"github.com/daytonaio/otel-proxy/internal"
	"github.com/gin-gonic/gin"

	common_cache "github.com/daytonaio/common-go/pkg/cache"
	common_errors "github.com/daytonaio/common-go/pkg/errors"
	common_proxy "github.com/daytonaio/common-go/pkg/proxy"

	log "github.com/sirupsen/logrus"
)

type Proxy struct {
	config *config.Config

	apiclient              *apiclient.APIClient
	authTokenEndpointCache common_cache.ICache[apiclient.OtelConfig]
}

func StartProxy(config *config.Config) error {
	proxy := &Proxy{
		config: config,
	}

	clientConfig := apiclient.NewConfiguration()
	clientConfig.Servers = apiclient.ServerConfigurations{
		{
			URL: config.DaytonaApiUrl,
		},
	}

	clientConfig.AddDefaultHeader("Authorization", "Bearer "+config.ApiKey)

	proxy.apiclient = apiclient.NewAPIClient(clientConfig)

	proxy.apiclient.GetConfig().HTTPClient = &http.Client{
		Transport: http.DefaultTransport,
	}

	if config.Redis != nil {
		var err error
		proxy.authTokenEndpointCache, err = common_cache.NewRedisCache[apiclient.OtelConfig](config.Redis, "otel-proxy:auth-token-endpoint:")
		if err != nil {
			return err
		}
	} else {
		proxy.authTokenEndpointCache = common_cache.NewMapCache[apiclient.OtelConfig]()
	}

	router := gin.New()
	router.Use(gin.Recovery())

	router.Use(common_errors.NewErrorMiddleware(func(ctx *gin.Context, err error) common_errors.ErrorResponse {
		return common_errors.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Message:    err.Error(),
		}
	}))

	router.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "ok", "version": internal.Version})
	})

	router.POST("/v1/metrics", func(ctx *gin.Context) {
		common_proxy.NewProxyRequestHandler(proxy.GetProxyTarget, nil)(ctx)
	})

	router.POST("/v1/traces", func(ctx *gin.Context) {
		common_proxy.NewProxyRequestHandler(proxy.GetProxyTarget, nil)(ctx)
	})

	router.POST("/v1/logs", func(ctx *gin.Context) {
		common_proxy.NewProxyRequestHandler(proxy.GetProxyTarget, nil)(ctx)
	})

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Port),
		Handler: router,
	}

	listener, err := net.Listen("tcp", httpServer.Addr)
	if err != nil {
		return err
	}

	log.Infof("OTEL Proxy server is running on port %d", config.Port)

	if config.EnableTLS {
		err = httpServer.ServeTLS(listener, config.TLSCertFile, config.TLSKeyFile)
	} else {
		err = httpServer.Serve(listener)
	}

	return err
}
