// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package proxy

import (
	"errors"
	"fmt"
	"maps"
	"net"
	"net/http"
	"slices"

	apiclient "github.com/daytonaio/apiclient"
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

	apiclient                *apiclient.APIClient
	runnerCache              cache.ICache[RunnerInfo]
	sandboxPublicCache       cache.ICache[bool]
	sandboxAuthKeyValidCache cache.ICache[bool]
}

func StartProxy(config *config.Config) error {
	proxy := &Proxy{
		config: config,
	}

	proxy.secureCookie = securecookie.New([]byte(config.ProxyApiKey), nil)

	clientConfig := apiclient.NewConfiguration()
	clientConfig.Servers = apiclient.ServerConfigurations{
		{
			URL: config.DaytonaApiUrl,
		},
	}

	clientConfig.AddDefaultHeader("Authorization", "Bearer "+config.ProxyApiKey)

	proxy.apiclient = apiclient.NewAPIClient(clientConfig)

	proxy.apiclient.GetConfig().HTTPClient = &http.Client{
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

	router.Use(func(ctx *gin.Context) {
		if ctx.Request.Header.Get("X-Daytona-Disable-CORS") == "true" {
			ctx.Request.Header.Del("X-Daytona-Disable-CORS")
			return
		}

		corsConfig := cors.DefaultConfig()
		corsConfig.AllowOriginFunc = func(origin string) bool {
			return true
		}
		corsConfig.AllowCredentials = true
		corsConfig.AllowHeaders = slices.Collect(maps.Keys(ctx.Request.Header))
		corsConfig.AllowHeaders = append(corsConfig.AllowHeaders, ctx.Request.Header.Values("Access-Control-Request-Headers")...)

		cors.New(corsConfig)(ctx)
	})

	router.Use(common_errors.NewErrorMiddleware(func(ctx *gin.Context, err error) common_errors.ErrorResponse {
		return common_errors.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Message:    err.Error(),
		}
	}))

	router.Any("/*path", func(ctx *gin.Context) {
		_, _, err := proxy.parseHost(ctx.Request.Host)
		// if the host is not valid, we don't proxy the request
		if err != nil {
			switch ctx.Request.Method {
			case "GET":
				switch ctx.Request.URL.Path {
				case "/callback":
					proxy.AuthCallback(ctx)
					return
				case "/health":
					ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
					return
				}
			}

			ctx.Error(common_errors.NewNotFoundError(errors.New("not found")))
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
