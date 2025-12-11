// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package proxy

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/docker/docker/client"
	"github.com/jellydator/ttlcache/v3"
)

type ProxyCacheItem struct {
	IP string
}

type ProxyServer struct {
	addr         string
	mux          *http.ServeMux
	dockerClient *client.Client
	cache        *ttlcache.Cache[string, *ProxyCacheItem]
	server       *http.Server
	useTLS       bool
	certFile     string
	keyFile      string
	transport    *http.Transport
	log          *slog.Logger
	cacheTTL     time.Duration
	targetPort   int
	network      string
}

type ProxyServerConfig struct {
	Addr         string
	Log          *slog.Logger
	UseTLS       bool
	CertFile     string
	KeyFile      string
	CacheTTL     time.Duration
	TargetPort   int
	Network      string
	DockerClient *client.Client
}

func NewProxyServer(config ProxyServerConfig) *ProxyServer {
	cacheTTL := config.CacheTTL
	if cacheTTL == 0 {
		cacheTTL = 10 * time.Minute
	}

	targetPort := config.TargetPort
	if targetPort == 0 {
		targetPort = 2280
	}

	network := config.Network
	if network == "" {
		network = "bridge"
	}

	cache := ttlcache.New(ttlcache.WithTTL[string, *ProxyCacheItem](cacheTTL))
	go cache.Start()

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
		cache:        cache,
		addr:         config.Addr,
		transport:    transport,
		log:          log,
		useTLS:       config.UseTLS,
		certFile:     config.CertFile,
		keyFile:      config.KeyFile,
		cacheTTL:     cacheTTL,
		targetPort:   targetPort,
		network:      network,
		dockerClient: config.DockerClient,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/sandboxes/{id}/toolbox/{path...}", s.proxyRequest)

	s.mux = mux

	return s
}

func (s *ProxyServer) Start() error {
	s.server = &http.Server{
		Addr:    s.addr,
		Handler: s.mux,
	}

	s.log.Info("Starting proxy server", "addr", s.addr)
	if s.useTLS {
		return s.server.ListenAndServeTLS(s.certFile, s.keyFile)
	}

	return s.server.ListenAndServe()
}

func (s *ProxyServer) Shutdown(ctx context.Context) error {
	s.cache.Stop()

	if s.transport != nil {
		s.transport.CloseIdleConnections()
	}

	return s.server.Shutdown(ctx)
}
