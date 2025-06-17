// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package proxy

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"

	"github.com/docker/docker/client"
	"github.com/jellydator/ttlcache/v3"

	"log/slog"
)

type ProxyCacheItem struct {
	IP string
}

type ProxyServer struct {
	Addr         string
	dockerClient *client.Client
	mux          *http.ServeMux
	cache        *ttlcache.Cache[string, *ProxyCacheItem]
	server       *http.Server
	useTLS       bool
	certFile     string
	keyFile      string
	transport    *http.Transport
	log          *slog.Logger
}

type ProxyServerConfig struct {
	DockerClient *client.Client
	Addr         string
	Log          *slog.Logger
	UseTLS       bool
	CertFile     string
	KeyFile      string
}

func New(config ProxyServerConfig) *ProxyServer {
	cache := ttlcache.New(ttlcache.WithTTL[string, *ProxyCacheItem](10 * time.Minute))
	go cache.Start()

	// Create a custom transport with connection pooling and keep-alive
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     60 * time.Second,
	}

	log := config.Log.With("service", "proxy")
	s := &ProxyServer{
		dockerClient: config.DockerClient,
		cache:        cache,
		Addr:         config.Addr,
		transport:    transport,
		log:          log,
		useTLS:       config.UseTLS,
		certFile:     config.CertFile,
		keyFile:      config.KeyFile,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/sandboxes/{id}/toolbox/{path...}", withBearerAuth(s.proxyRequest, os.Getenv("TOKEN")))

	s.mux = mux

	return s
}

func (s *ProxyServer) Start() error {
	s.server = &http.Server{
		Addr:    s.Addr,
		Handler: s.mux,
	}

	if s.useTLS {
		return s.server.ListenAndServeTLS(s.certFile, s.keyFile)
	}

	return s.server.ListenAndServe()
}

// Graceful shutdown
func (s *ProxyServer) Shutdown(ctx context.Context) error {
	// Stop the cache goroutine
	s.cache.Stop()

	// Close the transport to clean up idle connections
	if s.transport != nil {
		s.transport.CloseIdleConnections()
	}

	// Shutdown the server with context timeout
	return s.server.Shutdown(ctx)
}

func (s *ProxyServer) proxyRequest(w http.ResponseWriter, r *http.Request) {
	sandboxId := r.PathValue("id")
	path := r.PathValue("path")

	target, err := s.getProxyTarget(r.Context(), sandboxId)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Create reverse proxy
	proxy := &httputil.ReverseProxy{
		Transport: s.transport,
		Director: func(req *http.Request) {
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			req.URL.Path = "/" + path

			// Preserve query parameters from original request
			if r.URL.RawQuery != "" {
				req.URL.RawQuery = r.URL.RawQuery
			}

			// Copy all headers from the original request
			for k, v := range r.Header {
				req.Header[k] = v
			}

			// Set the Host to match the target
			req.Host = target.Host

			// Ensure keep-alive is enabled
			req.Header.Set("Connection", "keep-alive")
		},
	}

	// Serve the request
	proxy.ServeHTTP(w, r)
}

func (s *ProxyServer) getProxyTarget(ctx context.Context, sandboxId string) (*url.URL, error) {
	// Check if ip address is in cache
	cacheItem := s.cache.Get(sandboxId)
	containerIP := ""

	// If not found in cache, check docker client
	if cacheItem != nil {
		containerIP = cacheItem.Value().IP
		// extend the cache item
		s.cache.Set(sandboxId, &ProxyCacheItem{IP: containerIP}, 10*time.Minute)
	} else {
		// Get container details
		container, err := s.dockerClient.ContainerInspect(ctx, sandboxId)
		if err != nil {
			return nil, fmt.Errorf("sandbox container not found: %w", err)
		}

		for _, network := range container.NetworkSettings.Networks {
			containerIP = network.IPAddress
			break
		}

		if containerIP == "" {
			return nil, errors.New("container has no IP address, it might not be running")
		}

		// store ip address in cache
		s.cache.Set(sandboxId, &ProxyCacheItem{IP: containerIP}, 10*time.Minute)
	}

	// Build the target URL
	targetURL := fmt.Sprintf("http://%s:2280", containerIP)
	target, err := url.Parse(targetURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse target URL: %w", err)
	}

	return target, nil
}
