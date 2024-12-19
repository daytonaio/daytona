// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"encoding/json"
	"time"
)

const (
	MaxReconnectAttempts = 5
	InitialRetryDelay    = 1 * time.Second
	MaxRetryDelay        = 30 * time.Second
)

type LogPosition struct {
	Offset    int64     `json:"offset"`
	Timestamp time.Time `json:"timestamp"`
	LastLine  string    `json:"last_line,omitempty"`
}

func (p *LogPosition) Marshal() string {
	data, _ := json.Marshal(p)
	return string(data)
}

func UnmarshalPosition(data string) (*LogPosition, error) {
	var pos LogPosition
	if err := json.Unmarshal([]byte(data), &pos); err != nil {
		return nil, err
	}
	return &pos, nil
}
