// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package proxy

import (
	"errors"
	"fmt"
	"maps"
	"net"
	"net/http"
	"net/url"
	"slices"
	"strings"

	apiclient "github.com/daytonaio/apiclient"
	"github.com/daytonaio/proxy/cmd/proxy/config"
	"github.com/daytonaio/proxy/internal"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/securecookie"

	common_cache "github.com/daytonaio/common-go/pkg/cache"
	common_errors "github.com/daytonaio/common-go/pkg/errors"
	common_proxy "github.com/daytonaio/common-go/pkg/proxy"

	log "github.com/sirupsen/logrus"
)

type RunnerInfo struct {
	ApiUrl string `json:"apiUrl"`
	ApiKey string `json:"apiKey"`
}

const SANDBOX_AUTH_KEY_HEADER = "X-Daytona-Preview-Token"
const SANDBOX_AUTH_KEY_QUERY_PARAM = "DAYTONA_SANDBOX_AUTH_KEY"
const SANDBOX_AUTH_COOKIE_NAME = "daytona-sandbox-auth-"
const SKIP_LAST_ACTIVITY_UPDATE_HEADER = "X-Daytona-Skip-Last-Activity-Update"
const TERMINAL_PORT = "22222"
const TOOLBOX_PORT = "2280"

type Proxy struct {
	config       *config.Config
	secureCookie *securecookie.SecureCookie
	cookieDomain string

	apiclient                      *apiclient.APIClient
	runnerCache                    common_cache.ICache[RunnerInfo]
	sandboxPublicCache             common_cache.ICache[bool]
	sandboxAuthKeyValidCache       common_cache.ICache[bool]
	sandboxLastActivityUpdateCache common_cache.ICache[bool]
}

func StartProxy(config *config.Config) error {
	proxy := &Proxy{
		config: config,
	}

	proxy.secureCookie = securecookie.New([]byte(config.ProxyApiKey), nil)
	cookieDomain := config.ProxyDomain
	cookieDomain = strings.Split(cookieDomain, ":")[0]
	cookieDomain = fmt.Sprintf(".%s", cookieDomain)
	proxy.cookieDomain = cookieDomain

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
		proxy.runnerCache, err = common_cache.NewRedisCache[RunnerInfo](config.Redis, "proxy:sandbox-runner-info:")
		if err != nil {
			return err
		}
		proxy.sandboxPublicCache, err = common_cache.NewRedisCache[bool](config.Redis, "proxy:sandbox-public:")
		if err != nil {
			return err
		}
		proxy.sandboxAuthKeyValidCache, err = common_cache.NewRedisCache[bool](config.Redis, "proxy:sandbox-auth-key-valid:")
		if err != nil {
			return err
		}
		proxy.sandboxLastActivityUpdateCache, err = common_cache.NewRedisCache[bool](config.Redis, "proxy:sandbox-last-activity-update:")
		if err != nil {
			return err
		}
	} else {
		proxy.runnerCache = common_cache.NewMapCache[RunnerInfo]()
		proxy.sandboxPublicCache = common_cache.NewMapCache[bool]()
		proxy.sandboxAuthKeyValidCache = common_cache.NewMapCache[bool]()
		proxy.sandboxLastActivityUpdateCache = common_cache.NewMapCache[bool]()
	}

	router := gin.New()
	router.Use(gin.Recovery())

	router.Use(common_errors.NewErrorMiddleware(func(ctx *gin.Context, err error) common_errors.ErrorResponse {
		return common_errors.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Message:    err.Error(),
		}
	}))

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

	router.Use(proxy.browserWarningMiddleware())

	router.Any("/*path", func(ctx *gin.Context) {
		if ctx.Request.Method == "POST" && ctx.Request.URL.Path == ACCEPT_PREVIEW_PAGE_WARNING_PATH {
			handleAcceptProxyWarning(ctx, config.ProxyProtocol == "https")
			return
		}

		targetPort, _, err := proxy.parseHost(ctx.Request.Host)
		// if the host is not valid, we don't proxy the request
		if err != nil {
			switch ctx.Request.Method {
			case "GET":
				switch ctx.Request.URL.Path {
				case "/callback":
					proxy.AuthCallback(ctx)
					return
				case "/health":
					ctx.JSON(http.StatusOK, gin.H{"status": "ok", "version": internal.Version})
					return
				}
			}

			if strings.HasPrefix(ctx.Request.URL.Path, "/toolbox/") {
				_, sandboxID, _, err := proxy.parseToolboxSubpath(ctx.Request.URL.Path)
				if err != nil {
					ctx.Error(common_errors.NewNotFoundError(errors.New("not found")))
					return
				}

				prefix := fmt.Sprintf("/toolbox/%s", sandboxID)

				getProxyTarget := func(ctx *gin.Context) (*url.URL, map[string]string, error) {
					return proxy.GetProxyTarget(ctx, true)
				}

				modifyResponse := func(res *http.Response) error {
					if res.StatusCode >= 300 && res.StatusCode < 400 {
						if loc := res.Header.Get("Location"); !strings.HasPrefix(loc, prefix) {
							res.Header.Set("Location", prefix+loc)
						}
					}
					return nil
				}

				common_proxy.NewProxyRequestHandler(getProxyTarget, modifyResponse)(ctx)
				return
			}

			ctx.Error(common_errors.NewNotFoundError(errors.New("not found")))
			return
		}

		// If toolbox only mode is enabled, only allow requests to the toolbox port
		if targetPort != TOOLBOX_PORT && proxy.config.ToolboxOnlyMode {
			ctx.Error(common_errors.NewNotFoundError(errors.New("not found")))
			return
		}

		getProxyTarget := func(ctx *gin.Context) (*url.URL, map[string]string, error) {
			return proxy.GetProxyTarget(ctx, false)
		}

		common_proxy.NewProxyRequestHandler(getProxyTarget, nil)(ctx)
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
