// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package remotelogs

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/logs"
)

type remoteLoggerFactory struct {
	localLoggerFactory logs.ILoggerFactory
	serverUrl          string
	apiKey             string
	basePath           string
}

type RemoteLoggerFactoryConfig struct {
	LogsDir      string
	ServerUrl    string
	ServerApiKey string
	BasePath     string
}

func NewRemoteLoggerFactory(config RemoteLoggerFactoryConfig) logs.ILoggerFactory {
	loggerFactoryImpl := &remoteLoggerFactory{
		localLoggerFactory: logs.NewLoggerFactory(config.LogsDir),
		serverUrl:          config.ServerUrl,
		apiKey:             config.ServerApiKey,
		basePath:           config.BasePath,
	}

	return loggerFactoryImpl
}

func (r *remoteLoggerFactory) CreateLogger(id, label string, source logs.LogSource) (logs.Logger, error) {
	conn, _, err := util.GetWebsocketConn(context.Background(), fmt.Sprintf("%s/%s/write", r.basePath, id), r.serverUrl, r.apiKey, nil)
	if err != nil {
		return nil, err
	}

	localLogger, err := r.localLoggerFactory.CreateLogger(id, label, source)
	if err != nil {
		return nil, err
	}

	return &RemoteLogger{
		localLogger: localLogger,
		conn:        conn,
	}, nil
}

func (l *remoteLoggerFactory) CreateLogReader(id string) (io.Reader, error) {
	return nil, errors.New("not implemented")
}

func (l *remoteLoggerFactory) CreateLogWriter(id string) (io.WriteCloser, error) {
	return nil, errors.New("not implemented")
}
