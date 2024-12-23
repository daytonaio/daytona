// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package logs

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/daytonaio/daytona/internal/util"
)

type remoteLoggerFactory struct {
	localLoggerFactory ILoggerFactory
	apiUrl             string
	apiKey             string
	apiBasePath        ApiBasePath
}

type RemoteLoggerFactoryConfig struct {
	LogsDir     string
	ApiUrl      string
	ApiKey      string
	ApiBasePath ApiBasePath
}

type ApiBasePath string

var (
	ApiBasePathWorkspace ApiBasePath = "/log/workspace"
	ApiBasePathBuild     ApiBasePath = "/log/build"
	ApiBasePathRunner    ApiBasePath = "/log/runner"
	ApiBasePathTarget    ApiBasePath = "/log/target"
)

func (r *remoteLoggerFactory) CreateLogger(id, label string, source LogSource) (Logger, error) {
	conn, _, err := util.GetWebsocketConn(context.Background(), fmt.Sprintf("%s/%s/write", r.apiBasePath, id), r.apiUrl, r.apiKey, nil)
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
