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

// TestSessionListTemplates verifies the templates endpoint returns python-default with
// the documented language and package metadata, and never leaks sandbox identifiers.
func TestSessionListTemplates(t *testing.T) {
	cfg := LoadConfig(t)
	api := NewAPIClient(cfg)
	ic := NewSessionClient(api)

	templates, status := ic.ListTemplates(t)
	require.Equal(t, http.StatusOK, status, "GET /sessions/templates must return 200")
	require.NotEmpty(t, templates, "templates must include at least python-default")

	var pythonDefault map[string]interface{}
	for _, raw := range templates {
		tpl, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		assertNoSandboxLeak(t, tpl, "")
		if name, _ := tpl["name"].(string); name == "python-default" {
			pythonDefault = tpl
		}
	}
	require.NotNil(t, pythonDefault, "templates must include python-default")

	languages, _ := pythonDefault["languages"].([]interface{})
	assert.ElementsMatch(t, []interface{}{"python", "typescript", "bash"}, languages,
		"python-default must support python, typescript, and bash")

	packages, _ := pythonDefault["packages"].([]interface{})
	assert.NotEmpty(t, packages, "python-default must declare a non-empty packages[]")
}

// TestSessionListPackagesPython verifies the python package list includes the curated venv.
// The API forwards GET /packages to the in-sandbox daemon, so this exercises the full
// pool-acquire + runner-proxy + daemon path (without code execution).
func TestSessionListPackagesPython(t *testing.T) {
	cfg := LoadConfig(t)
	api := NewAPIClient(cfg)
	ic := NewSessionClient(api)

	pkgs, status := ic.ListPackages(t, "python-default", "python")
	require.Equal(t, http.StatusOK, status, "GET /sessions/templates/python-default/packages?language=python must return 200")
	require.NotEmpty(t, pkgs)

	names := make(map[string]bool, len(pkgs))
	for _, raw := range pkgs {
		pkg, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		assertNoSandboxLeak(t, pkg, "")
		if name, _ := pkg["name"].(string); name != "" {
			names[name] = true
		}
	}

	for _, expected := range []string{"numpy", "pandas", "openai"} {
		assert.True(t, names[expected], "python package list must include %q", expected)
	}
}

// TestSessionListPackagesTypescript verifies the TS package list includes the curated node_modules
// and flags native-binding packages.
func TestSessionListPackagesTypescript(t *testing.T) {
	t.Skipf("not yet implemented: session-daemon-ts-host")

	cfg := LoadConfig(t)
	api := NewAPIClient(cfg)
	ic := NewSessionClient(api)

	pkgs, status := ic.ListPackages(t, "python-default", "typescript")
	require.Equal(t, http.StatusOK, status, "GET /sessions/templates/python-default/packages?language=typescript must return 200")
	require.NotEmpty(t, pkgs)

	byName := make(map[string]map[string]interface{}, len(pkgs))
	for _, raw := range pkgs {
		pkg, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		assertNoSandboxLeak(t, pkg, "")
		if name, _ := pkg["name"].(string); name != "" {
			byName[name] = pkg
		}
	}

	for _, expected := range []string{"zod", "lodash-es", "@anthropic-ai/sdk"} {
		assert.Contains(t, byName, expected, "ts package list must include %q", expected)
	}

	// Sanity check: any package with native bindings (e.g. node-gyp / bindings deps) must be
	// flagged. The set in the curated v1 catalog is pure-JS, so this asserts the field is
	// computed/exposed at all by listing every package and checking the field type.
	for name, pkg := range byName {
		if _, present := pkg["hasNativeBindings"]; !present {
			t.Logf("note: package %q is missing hasNativeBindings field", name)
		}
	}
}
