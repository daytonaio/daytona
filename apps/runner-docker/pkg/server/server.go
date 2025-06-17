// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package server

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/daytonaio/runner-docker/pkg/cache"
	pb "github.com/daytonaio/runner/proto"
	"github.com/docker/docker/client"
)

type RunnerServerConfig struct {
	ApiClient          client.APIClient
	Cache              cache.IRunnerCache
	LogWriter          io.Writer
	AWSRegion          string
	AWSEndpointUrl     string
	AWSAccessKeyId     string
	AWSSecretAccessKey string
	DaemonPath         string
}

type RunnerServer struct {
	pb.UnimplementedRunnerServer
	apiClient          client.APIClient
	cache              cache.IRunnerCache
	logWriter          io.Writer
	awsRegion          string
	awsEndpointUrl     string
	awsAccessKeyId     string
	awsSecretAccessKey string
	daemonPath         string
	volumeMutexes      map[string]*sync.Mutex
	volumeMutexesMutex sync.Mutex
	proxyClient        *http.Client
}

func NewRunnerServer(config RunnerServerConfig) *RunnerServer {
	// Same transport configuration as original Gin code
	proxyTransport := &http.Transport{
		MaxIdleConns:        100,
		IdleConnTimeout:     90 * time.Second,
		MaxIdleConnsPerHost: 100,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	}

	// Same redirect policy as original Gin code
	proxyClient := &http.Client{
		Transport: proxyTransport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) > 0 {
				for key, values := range via[0].Header {
					if key != "Authorization" && key != "Cookie" {
						for _, value := range values {
							req.Header.Add(key, value)
						}
					}
				}
			}
			if len(via) >= 10 {
				return fmt.Errorf("stopped after 10 redirects")
			}
			return nil
		},
	}

	return &RunnerServer{
		apiClient:          config.ApiClient,
		cache:              config.Cache,
		logWriter:          config.LogWriter,
		awsRegion:          config.AWSRegion,
		awsEndpointUrl:     config.AWSEndpointUrl,
		awsAccessKeyId:     config.AWSAccessKeyId,
		awsSecretAccessKey: config.AWSSecretAccessKey,
		daemonPath:         config.DaemonPath,
		volumeMutexes:      make(map[string]*sync.Mutex),
		proxyClient:        proxyClient,
	}
}
