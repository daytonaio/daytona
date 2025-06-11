// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package proxy

import (
	"fmt"
	"net"
	"net/http"

	"github.com/daytonaio/daytona/daytonaapiclient"
	"github.com/daytonaio/proxy/cmd/proxy/config"
	"github.com/daytonaio/proxy/pkg/cache"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/securecookie"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
	common_proxy "github.com/daytonaio/common-go/pkg/proxy"

	log "github.com/sirupsen/logrus"
)

type RunnerInfo struct {
	ApiUrl string `json:"apiUrl"`
	ApiKey string `json:"apiKey"`
}

const DAYTONA_SANDBOX_AUTH_KEY_HEADER = "X-Daytona-Preview-Token"
const DAYTONA_SANDBOX_AUTH_KEY_QUERY_PARAM = "DAYTONA_SANDBOX_AUTH_KEY"
const DAYTONA_SANDBOX_AUTH_COOKIE_NAME = "daytona-sandbox-auth-"

type Proxy struct {
	config       *config.Config
	secureCookie *securecookie.SecureCookie

	daytonaApiClient         *daytonaapiclient.APIClient
	runnerCache              cache.ICache[RunnerInfo]
	sandboxPublicCache       cache.ICache[bool]
	sandboxAuthKeyValidCache cache.ICache[bool]
}

func StartProxy(config *config.Config) error {
	proxy := &Proxy{
		config: config,
	}

	proxy.secureCookie = securecookie.New([]byte(config.ProxyApiKey), nil)

	clientConfig := daytonaapiclient.NewConfiguration()
	clientConfig.Servers = daytonaapiclient.ServerConfigurations{
		{
			URL: config.DaytonaApiUrl,
		},
	}

	clientConfig.AddDefaultHeader("Authorization", "Bearer "+config.ProxyApiKey)

	proxy.daytonaApiClient = daytonaapiclient.NewAPIClient(clientConfig)

	proxy.daytonaApiClient.GetConfig().HTTPClient = &http.Client{
		Transport: http.DefaultTransport,
	}

	if config.Redis != nil {
		var err error
		proxy.runnerCache, err = cache.NewRedisCache[RunnerInfo](config.Redis, "proxy:sandbox-runner-info:")
		if err != nil {
			return err
		}
		proxy.sandboxPublicCache, err = cache.NewRedisCache[bool](config.Redis, "proxy:sandbox-public:")
		if err != nil {
			return err
		}
		proxy.sandboxAuthKeyValidCache, err = cache.NewRedisCache[bool](config.Redis, "proxy:sandbox-auth-key-valid:")
		if err != nil {
			return err
		}
	} else {
		proxy.runnerCache = cache.NewMapCache[RunnerInfo]()
		proxy.sandboxPublicCache = cache.NewMapCache[bool]()
		proxy.sandboxAuthKeyValidCache = cache.NewMapCache[bool]()
	}

	router := gin.New()
	router.Use(gin.Recovery())

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowCredentials = true
	router.Use(cors.New(corsConfig))

	router.Use(common_errors.NewErrorMiddleware(func(ctx *gin.Context, err error) common_errors.ErrorResponse {
		return common_errors.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Message:    err.Error(),
		}
	}))

	router.Any("/*path", func(ctx *gin.Context) {
		if ctx.Request.Host == config.ProxyDomain && ctx.Request.Method == "GET" && ctx.Request.URL.Path == "/callback" {
			proxy.AuthCallback(ctx)
			return
		}

		common_proxy.NewProxyRequestHandler(proxy.GetProxyTarget)(ctx)
	})

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.ProxyPort),
		Handler: router,
	}

	listener, err := net.Listen("tcp", httpServer.Addr)
	if err != nil {
		return err
	}

	log.Infof("Proxy server is running on port %d", config.ProxyPort)

	if config.EnableTLS {
		err = httpServer.ServeTLS(listener, config.TLSCertFile, config.TLSKeyFile)
	} else {
		err = httpServer.Serve(listener)
	}

	return err
}
