// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package telemetry

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/daytonaio/daytona/internal"
	"github.com/google/uuid"
)

type TelemetryService interface {
	io.Closer
	Track(event Event, clientId string) error
}

func TelemetryEnabled(ctx context.Context) bool {
	enabled, ok := ctx.Value(ENABLED_CONTEXT_KEY).(bool)
	if !ok {
		return false
	}

	return enabled
}

func ClientId(ctx context.Context) string {
	id, ok := ctx.Value(CLIENT_ID_CONTEXT_KEY).(string)
	if !ok {
		// To identify requests that had no client ID set
		return fmt.Sprintf("%s-invalid-client-id", uuid.NewString()[0:16])
	}

	return id
}

func SessionId(ctx context.Context) string {
	id, ok := ctx.Value(SESSION_ID_CONTEXT_KEY).(string)
	if !ok {
		return internal.SESSION_ID
	}

	return id
}

func ServerId(ctx context.Context) string {
	id, ok := ctx.Value(SERVER_ID_CONTEXT_KEY).(string)
	if !ok {
		// To identify requests that had no server ID set
		return fmt.Sprintf("%s-invalid-server-id", uuid.NewString()[0:16])
	}

	return id
}

func SetCommonProps(version string, source TelemetrySource, properties map[string]interface{}) {
	properties["daytona_version"] = version
	properties["source"] = source
	properties["os"] = runtime.GOOS
	properties["arch"] = runtime.GOARCH
}

func isImagePublic(imageName string) bool {
	if strings.Count(imageName, "/") < 2 {
		if strings.Count(imageName, "/") == 0 {
			return isPublic("https://hub.docker.com/_/" + imageName)
		}

		return isPublic("https://hub.docker.com/r/" + imageName)
	}

	return isPublic(imageName)
}

func isPublic(url string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	_, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
	cancel()
	return err == nil
}
