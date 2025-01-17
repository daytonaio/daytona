// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package telemetry

import "github.com/daytonaio/daytona/pkg/models"

type ApiKeyEventName string

const (
	ApiKeyEventLifecycleCreated        ApiKeyEventName = "api_key_lifecycle_created"
	ApiKeyEventLifecycleCreationFailed ApiKeyEventName = "api_key_lifecycle_creation_failed"
)

type ApiKeyEvent struct {
	key *models.ApiKey
	AbstractEvent
}

func NewApiKeyEvent(name ApiKeyEventName, key *models.ApiKey, err error, extras map[string]interface{}) Event {
	return ApiKeyEvent{
		key: key,
		AbstractEvent: AbstractEvent{
			name:   string(name),
			extras: extras,
			err:    err,
		},
	}
}

func (e ApiKeyEvent) Props() map[string]interface{} {
	props := e.AbstractEvent.Props()

	if e.key == nil {
		return props
	}

	props["type"] = e.key.Type
	return props
}
