/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 *
 * Package specs provides helpers for parsing v2 job payloads defined in
 * apps/runner/specs/runner.proto.
 *
 * Generated proto types live in the gen/ sub-package (package specsgen).
 * This file exposes ParsePayload, which deserialises a job payload JSON string
 * into any proto.Message using protojson — the format that correctly maps the
 * camelCase JSON keys produced by the TypeScript API to proto field names.
 */

package specs

import (
	"fmt"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// ParsePayload deserialises the JSON job payload into a generated proto message.
// It accepts the camelCase JSON keys emitted by the TypeScript RunnerAdapterV2
// (e.g. "cpuQuota", "networkBlockAll") as well as the proto snake_case names.
// Unknown fields are silently discarded so that future proto additions are
// backwards-compatible with older runners.
func ParsePayload(payload *string, msg proto.Message) error {
	if payload == nil || *payload == "" {
		return fmt.Errorf("payload is required")
	}

	opts := protojson.UnmarshalOptions{
		DiscardUnknown: true,
		AllowPartial:   true,
	}

	if err := opts.Unmarshal([]byte(*payload), msg); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	return nil
}
