// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"

	"github.com/gorilla/websocket"
)

func GetWebsocketConn(ctx context.Context, path string, serverUrl string, apiKey string, query *string) (*websocket.Conn, *http.Response, error) {
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

	return websocket.DefaultDialer.DialContext(ctx, wsUrl, http.Header{
		"Authorization": []string{fmt.Sprintf("Bearer %s", apiKey)},
	})
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
