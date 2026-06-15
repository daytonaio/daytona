// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"os/exec"
	"slices"
	"testing"
)

func TestApplyEnvsMergesAcrossCalls(t *testing.T) {
	cmd := exec.Command("true")

	// First layer: wrapper-extracted vars (execute_windows buildExecCmd).
	ApplyEnvs(cmd, map[string]string{"WRAPPER_VAR": "from-wrapper", "SHARED": "wrapper"})
	// Second layer: request envs applied by the shared handler.
	ApplyEnvs(cmd, map[string]string{"REQUEST_VAR": "from-request", "SHARED": "request"})

	if !slices.Contains(cmd.Env, "WRAPPER_VAR=from-wrapper") {
		t.Fatalf("wrapper var dropped by second ApplyEnvs call; env=%v", cmd.Env)
	}
	if !slices.Contains(cmd.Env, "REQUEST_VAR=from-request") {
		t.Fatalf("request var missing; env=%v", cmd.Env)
	}
	// os/exec keeps the LAST duplicate, so request must come after wrapper
	// for request-wins precedence on shared keys.
	if slices.Index(cmd.Env, "SHARED=request") < slices.Index(cmd.Env, "SHARED=wrapper") {
		t.Fatalf("request layer does not take precedence; env=%v", cmd.Env)
	}
}

func TestApplyEnvsEmptyIsNoop(t *testing.T) {
	cmd := exec.Command("true")
	ApplyEnvs(cmd, nil)
	if cmd.Env != nil {
		t.Fatalf("empty ApplyEnvs must not materialize cmd.Env; got %d entries", len(cmd.Env))
	}
}
