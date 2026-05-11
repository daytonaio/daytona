// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

//go:build e2e

package e2e_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSessionCodeRunPython runs `print(7*8)` and validates stdout, no error, positive duration,
// no sandbox identifiers in the response body.
func TestSessionCodeRunPython(t *testing.T) {
	cfg := LoadConfig(t)
	api := NewAPIClient(cfg)
	ic := NewSessionClient(api)

	body, status := ic.CodeRun(t, map[string]interface{}{
		"language": "python",
		"code":     "print(7*8)",
	})
	require.Equal(t, http.StatusOK, status, "code-run must return 200")

	stdout, _ := body["stdout"].(string)
	assert.Equal(t, "56\n", stdout, "stdout must contain `56\\n`")

	if errVal, ok := body["error"]; ok {
		assert.Nil(t, errVal, "error must be null on success")
	}

	durationMs, _ := body["durationMs"].(float64)
	assert.Greater(t, durationMs, float64(0), "durationMs must be > 0")

	assertNoSandboxLeak(t, body, "")
}

// TestSessionCodeRunPythonError verifies a typed error response on division by zero.
func TestSessionCodeRunPythonError(t *testing.T) {
	t.Skipf("not yet implemented: session-service-controller")

	cfg := LoadConfig(t)
	api := NewAPIClient(cfg)
	ic := NewSessionClient(api)

	body, status := ic.CodeRun(t, map[string]interface{}{
		"language": "python",
		"code":     "1/0",
	})
	require.Equal(t, http.StatusOK, status, "code-run must return 200 even when user code raises")

	errMap, _ := body["error"].(map[string]interface{})
	require.NotNil(t, errMap, "error must be present for runtime exceptions")

	name, _ := errMap["name"].(string)
	assert.Equal(t, "ZeroDivisionError", name)

	traceback, _ := errMap["traceback"].(string)
	assert.Contains(t, traceback, "division by zero")
}

// TestSessionCodeRunTypescript runs `console.log(1+1)` in TypeScript.
func TestSessionCodeRunTypescript(t *testing.T) {
	t.Skipf("not yet implemented: session-daemon-ts-host")

	cfg := LoadConfig(t)
	api := NewAPIClient(cfg)
	ic := NewSessionClient(api)

	body, status := ic.CodeRun(t, map[string]interface{}{
		"language": "typescript",
		"code":     "console.log(1+1)",
	})
	require.Equal(t, http.StatusOK, status)

	stdout, _ := body["stdout"].(string)
	assert.Equal(t, "2\n", stdout)
}

// TestSessionCodeRunEnv verifies env vars are visible in both Python (os.environ) and TS (env global).
func TestSessionCodeRunEnv(t *testing.T) {
	t.Skipf("not yet implemented: session-daemon-ts-host")

	cfg := LoadConfig(t)
	api := NewAPIClient(cfg)
	ic := NewSessionClient(api)

	t.Run("Python_OsEnviron", func(t *testing.T) {
		body, status := ic.CodeRun(t, map[string]interface{}{
			"language": "python",
			"env":      map[string]string{"FOO": "bar"},
			"code":     "import os; print(os.environ['FOO'])",
		})
		require.Equal(t, http.StatusOK, status)
		stdout, _ := body["stdout"].(string)
		assert.Equal(t, "bar\n", stdout)
	})

	t.Run("TS_EnvGlobal", func(t *testing.T) {
		body, status := ic.CodeRun(t, map[string]interface{}{
			"language": "typescript",
			"env":      map[string]string{"FOO": "bar"},
			"code":     "console.log(env.FOO)",
		})
		require.Equal(t, http.StatusOK, status)
		stdout, _ := body["stdout"].(string)
		assert.Equal(t, "bar\n", stdout)
	})
}

// TestSessionCodeRunTimeout verifies the timeout option enforces an upper bound on execution.
func TestSessionCodeRunTimeout(t *testing.T) {
	t.Skipf("not yet implemented: session-service-controller")

	cfg := LoadConfig(t)
	api := NewAPIClient(cfg)
	ic := NewSessionClient(api)

	body, status := ic.CodeRun(t, map[string]interface{}{
		"language": "python",
		"code":     "import time; time.sleep(10)",
		"timeout":  2,
	})
	require.Equal(t, http.StatusOK, status)

	errMap, _ := body["error"].(map[string]interface{})
	require.NotNil(t, errMap, "error must be present on timeout")
	name, _ := errMap["name"].(string)
	assert.Equal(t, "TimeoutError", name)

	durationMs, _ := body["durationMs"].(float64)
	assert.GreaterOrEqual(t, durationMs, float64(2000), "durationMs must be >= 2000")
	assert.Less(t, durationMs, float64(4000), "durationMs must be < 4000 (2s timeout + slack)")
}
