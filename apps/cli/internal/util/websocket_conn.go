/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

package util

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"

	"github.com/daytonaio/daytona/cli/config"
	"github.com/gorilla/websocket"
)

func GetWebsocketConn(ctx context.Context, path string, serverUrl string, serverApi config.ServerApi, query *string) (*websocket.Conn, *http.Response, error) {
	url, err := url.JoinPath(serverUrl, path)
	if err != nil {
		return nil, nil, err
	}

	wsUrl, err := getWebSocketUrl(url)
	if err != nil {
		return nil, nil, err
	}

	if query != nil {
		wsUrl = fmt.Sprintf("%s?%s", wsUrl, *query)
	}

	headers := http.Header{}

	if serverApi.Key != nil {
		headers.Add("Authorization", fmt.Sprintf("Bearer %s", *serverApi.Key))
	} else if serverApi.Token != nil {
		headers.Add("Authorization", fmt.Sprintf("Bearer %s", serverApi.Token.AccessToken))
	}

	return websocket.DefaultDialer.DialContext(ctx, wsUrl, headers)
}

func getWebSocketUrl(apiUrl string) (string, error) {
	hostRegex := regexp.MustCompile(`(https*)://(.*)`)

	matches := hostRegex.FindStringSubmatch(apiUrl)

	if len(matches) != 3 {
		return "", errors.New("invalid API URL")
	}

	switch matches[1] {
	case "http":
		return fmt.Sprintf("ws://%s", matches[2]), nil
	case "https":
		return fmt.Sprintf("wss://%s", matches[2]), nil
	}

	return "", errors.New("invalid API URL protocol")
}
