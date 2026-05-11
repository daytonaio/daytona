// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

//go:build e2e

package e2e_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSessionTsImportZod proves the bundler resolves curated node_modules and
// Session.compileModule works.
func TestSessionTsImportZod(t *testing.T) {
	t.Skipf("not yet implemented: session-daemon-ts-host")

	cfg := LoadConfig(t)
	api := NewAPIClient(cfg)
	ic := NewSessionClient(api)

	body, status := ic.CodeRun(t, map[string]interface{}{
		"language": "typescript",
		"code":     `import { z } from 'zod';\nconsole.log(z.string().parse('hi'));`,
	})
	require.Equal(t, http.StatusOK, status)
	stdout, _ := body["stdout"].(string)
	assert.Contains(t, stdout, "hi", "zod parse must succeed and print 'hi'")
}

// TestSessionTsFetch proves the host-side fetch Reference bridge works.
func TestSessionTsFetch(t *testing.T) {
	t.Skipf("not yet implemented: session-daemon-ts-host")

	cfg := LoadConfig(t)
	api := NewAPIClient(cfg)
	ic := NewSessionClient(api)

	body, status := ic.CodeRun(t, map[string]interface{}{
		"language": "typescript",
		"code":     `const r = await fetch('https://httpbin.org/get').then((x: any) => x.json()); console.log(typeof r.url);`,
		"timeout":  30,
	})
	require.Equal(t, http.StatusOK, status)
	stdout, _ := body["stdout"].(string)
	assert.Contains(t, stdout, "string", "fetch().then(json) must yield an object whose .url is a string")
}

// TestSessionTsNativeModuleRejected verifies imports of Node-native modules raise NativeModuleError.
func TestSessionTsNativeModuleRejected(t *testing.T) {
	t.Skipf("not yet implemented: session-daemon-ts-host")

	cfg := LoadConfig(t)
	api := NewAPIClient(cfg)
	ic := NewSessionClient(api)

	body, status := ic.CodeRun(t, map[string]interface{}{
		"language": "typescript",
		"code":     `import * as fs from 'fs'; console.log(typeof fs.readFileSync);`,
	})
	require.Equal(t, http.StatusOK, status, "user-side errors return 200 with error payload")
	errMap, _ := body["error"].(map[string]interface{})
	require.NotNil(t, errMap)
	name, _ := errMap["name"].(string)
	assert.Equal(t, "NativeModuleError", name, "fs is Node-native, must be rejected")
	value, _ := errMap["value"].(string)
	assert.Contains(t, value, "fs", "error message must name the offending package")
}

// TestSessionTsProcessNotExposed verifies process / Buffer are not defined but env is accessible.
func TestSessionTsProcessNotExposed(t *testing.T) {
	t.Skipf("not yet implemented: session-daemon-ts-host")

	cfg := LoadConfig(t)
	api := NewAPIClient(cfg)
	ic := NewSessionClient(api)

	body, status := ic.CodeRun(t, map[string]interface{}{
		"language": "typescript",
		"env":      map[string]string{"FOO": "bar"},
		"code": `console.log(typeof process);
console.log(typeof Buffer);
console.log(env.FOO);`,
	})
	require.Equal(t, http.StatusOK, status)
	stdout, _ := body["stdout"].(string)
	assert.Contains(t, stdout, "undefined")
	assert.Contains(t, stdout, "bar")
}

// TestSessionTsBundleCacheReuse verifies that two contexts in the same sandbox sharing an
// import resolve the second context's first import faster (bundle cache hit).
func TestSessionTsBundleCacheReuse(t *testing.T) {
	t.Skipf("not yet implemented: session-daemon-ts-host")

	cfg := LoadConfig(t)
	api := NewAPIClient(cfg)
	ic := NewSessionClient(api)

	measure := func() time.Duration {
		body, status := ic.Connect(t, map[string]interface{}{
			"template": "python-default",
			"language": "typescript",
		})
		require.Equal(t, http.StatusOK, status)
		wsURL, _ := body["wsUrl"].(string)
		ctxID, _ := body["sessionId"].(string)
		t.Cleanup(func() { _ = ic.DeleteSession(t, ctxID) })

		ws, closer := dialSessionWebSocket(t, wsURL)
		defer closer()

		start := time.Now()
		sendExec(t, ws, `import _ from 'lodash-es'; console.log(_.chunk([1,2,3,4],2).length);`, nil)
		_ = collectFrames(t, ws, 60*time.Second)
		return time.Since(start)
	}

	first := measure()
	second := measure()
	t.Logf("first import: %s, second (cached): %s", first, second)
	assert.Less(t, second, first, "second import in the same sandbox must be faster (bundle cache hit)")
}
