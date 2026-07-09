// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package dind

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func decode(t *testing.T, b []byte) map[string]any {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.Unmarshal(b, &m))
	return m
}

func TestMergeRegistryMirror_EmptyConfigAddsMirror(t *testing.T) {
	out, changed, err := mergeRegistryMirror(nil, "https://mirror.example.com")
	require.NoError(t, err)
	require.True(t, changed)

	cfg := decode(t, out)
	require.Equal(t, []any{"https://mirror.example.com"}, cfg["registry-mirrors"])
	// HTTPS mirrors must not be added to insecure-registries.
	_, hasInsecure := cfg["insecure-registries"]
	require.False(t, hasInsecure)
}

func TestMergeRegistryMirror_HTTPMirrorTrustedAsInsecure(t *testing.T) {
	out, changed, err := mergeRegistryMirror(nil, "http://mirror.example.com:5001")
	require.NoError(t, err)
	require.True(t, changed)

	cfg := decode(t, out)
	require.Equal(t, []any{"http://mirror.example.com:5001"}, cfg["registry-mirrors"])
	require.Equal(t, []any{"mirror.example.com:5001"}, cfg["insecure-registries"])
}

func TestMergeRegistryMirror_PreservesExistingKeys(t *testing.T) {
	existing := []byte(`{"insecure-registries":["registry:6000"],"log-level":"warn"}`)

	out, changed, err := mergeRegistryMirror(existing, "http://mirror.example.com:5001")
	require.NoError(t, err)
	require.True(t, changed)

	cfg := decode(t, out)
	require.Equal(t, "warn", cfg["log-level"])
	require.ElementsMatch(t, []any{"registry:6000", "mirror.example.com:5001"}, cfg["insecure-registries"])
	require.Equal(t, []any{"http://mirror.example.com:5001"}, cfg["registry-mirrors"])
}

func TestMergeRegistryMirror_DoesNotOverwriteExistingMirror(t *testing.T) {
	existing := []byte(`{"registry-mirrors":["https://user-mirror.example.com"]}`)

	out, changed, err := mergeRegistryMirror(existing, "https://mirror.example.com")
	require.NoError(t, err)
	require.False(t, changed)
	require.Equal(t, existing, out)
}

func TestMergeRegistryMirror_InvalidExistingJSON(t *testing.T) {
	_, _, err := mergeRegistryMirror([]byte("not json"), "https://mirror.example.com")
	require.Error(t, err)
}

func TestMergeRegistryMirror_InvalidMirrorURL(t *testing.T) {
	_, _, err := mergeRegistryMirror(nil, "://bad")
	require.Error(t, err)
}

func TestAddInsecureRegistry_DeduplicatesAndSorts(t *testing.T) {
	got := addInsecureRegistry([]any{"b:1", "a:1", "b:1"}, "a:1")
	require.Equal(t, []string{"a:1", "b:1"}, got)
}

func TestConfigureRegistryMirror_EmptyURLIsNoop(t *testing.T) {
	require.NoError(t, ConfigureRegistryMirror(nil, ""))
}

func TestConfigureRegistryMirror_InvalidURL(t *testing.T) {
	require.Error(t, ConfigureRegistryMirror(nil, "://nope"))
}

// TestConfigureRegistryMirror_WritesFile exercises the full write path against
// a temporary daemon.json.
func TestConfigureRegistryMirror_WritesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "daemon.json")

	require.NoError(t, configureRegistryMirrorAt(nil, path, "https://mirror.example.com"))

	b, err := os.ReadFile(path)
	require.NoError(t, err)
	cfg := decode(t, b)
	require.Equal(t, []any{"https://mirror.example.com"}, cfg["registry-mirrors"])
}

// TestConfigureRegistryMirror_DoesNotOverwriteExistingFile ensures a user's
// pre-existing mirror in daemon.json is left untouched.
func TestConfigureRegistryMirror_DoesNotOverwriteExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "daemon.json")
	original := []byte(`{"registry-mirrors":["https://user-mirror.example.com"]}`)
	require.NoError(t, os.WriteFile(path, original, 0644))

	require.NoError(t, configureRegistryMirrorAt(nil, path, "https://mirror.example.com"))

	b, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, original, b)
}
