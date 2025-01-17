// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package telemetry

type Event interface {
	Name() string
	Props() map[string]interface{}
}

type AbstractEvent struct {
	name   string
	extras map[string]interface{}
	err    error
}

func (e AbstractEvent) Name() string {
	return e.name
}

func (e AbstractEvent) Props() map[string]interface{} {
	props := map[string]interface{}{}
	if e.err != nil {
		props["error"] = e.err.Error()
	}

	for k, v := range e.extras {
		props[k] = v
	}

	return props
}
