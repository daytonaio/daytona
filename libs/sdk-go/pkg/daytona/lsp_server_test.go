// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package daytona

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/daytonaio/daytona/libs/sdk-go/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLspServerServiceCreation(t *testing.T) {
	lsp := NewLspServerService(nil, types.LspLanguagePython, "/project", nil)
	require.NotNil(t, lsp)
	assert.Equal(t, types.LspLanguageID("python"), lsp.languageID)
	assert.Equal(t, "/project", lsp.projectPath)
}

func TestLspServerStart(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := createTestToolboxClient(server)
	lsp := NewLspServerService(client, types.LspLanguagePython, "/project", nil)

	ctx := context.Background()
	err := lsp.Start(ctx)
	assert.NoError(t, err)
}

func TestLspServerStop(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := createTestToolboxClient(server)
	lsp := NewLspServerService(client, types.LspLanguagePython, "/project", nil)

	ctx := context.Background()
	err := lsp.Stop(ctx)
	assert.NoError(t, err)
}

func TestLspServerDidOpen(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := createTestToolboxClient(server)
	lsp := NewLspServerService(client, types.LspLanguagePython, "/project", nil)

	ctx := context.Background()
	err := lsp.DidOpen(ctx, "/project/main.py")
	assert.NoError(t, err)
}

func TestLspServerDidOpenWithFilePrefix(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := createTestToolboxClient(server)
	lsp := NewLspServerService(client, types.LspLanguagePython, "/project", nil)

	ctx := context.Background()
	err := lsp.DidOpen(ctx, "file:///project/main.py")
	assert.NoError(t, err)
}

func TestLspServerDidClose(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := createTestToolboxClient(server)
	lsp := NewLspServerService(client, types.LspLanguagePython, "/project", nil)

	ctx := context.Background()
	err := lsp.DidClose(ctx, "/project/main.py")
	assert.NoError(t, err)
}

func TestLspServerDocumentSymbolsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "symbols failed"})
	}))
	defer server.Close()

	client := createTestToolboxClient(server)
	lsp := NewLspServerService(client, types.LspLanguagePython, "/project", nil)

	ctx := context.Background()
	_, err := lsp.DocumentSymbols(ctx, "/project/main.py")
	require.Error(t, err)
}

func TestLspServerSandboxSymbolsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "workspace symbols failed"})
	}))
	defer server.Close()

	client := createTestToolboxClient(server)
	lsp := NewLspServerService(client, types.LspLanguagePython, "/project", nil)

	ctx := context.Background()
	_, err := lsp.SandboxSymbols(ctx, "MyClass")
	require.Error(t, err)
}

func TestLspServerCompletionsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "completions failed"})
	}))
	defer server.Close()

	client := createTestToolboxClient(server)
	lsp := NewLspServerService(client, types.LspLanguagePython, "/project", nil)

	ctx := context.Background()
	pos := types.Position{Line: 10, Character: 5}
	_, err := lsp.Completions(ctx, "/project/main.py", pos)
	require.Error(t, err)
}

func TestLspServerErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "server error"})
	}))
	defer server.Close()

	client := createTestToolboxClient(server)
	lsp := NewLspServerService(client, types.LspLanguagePython, "/project", nil)

	ctx := context.Background()
	err := lsp.Start(ctx)
	require.Error(t, err)
}

func TestLspServerWithTypeScript(t *testing.T) {
	lsp := NewLspServerService(nil, types.LspLanguageTypeScript, "/ts-project", nil)
	assert.Equal(t, types.LspLanguageID("typescript"), lsp.languageID)
	assert.Equal(t, "/ts-project", lsp.projectPath)
}

func TestLspServerWithJavaScript(t *testing.T) {
	lsp := NewLspServerService(nil, types.LspLanguageJavaScript, "/js-project", nil)
	assert.Equal(t, types.LspLanguageID("javascript"), lsp.languageID)
}
