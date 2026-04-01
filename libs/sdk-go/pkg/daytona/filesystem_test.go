// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package daytona

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/daytonaio/daytona/libs/sdk-go/pkg/options"
	toolbox "github.com/daytonaio/daytona/libs/toolbox-api-client-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestToolboxClient(server *httptest.Server) *toolbox.APIClient {
	cfg := toolbox.NewConfiguration()
	cfg.Servers = toolbox.ServerConfigurations{{URL: server.URL}}
	return toolbox.NewAPIClient(cfg)
}

func TestFileSystemServiceCreation(t *testing.T) {
	fs := NewFileSystemService(nil, nil)
	require.NotNil(t, fs)
}

func TestCreateFolder(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "folder")
		assert.Equal(t, "/home/user/mydir", r.URL.Query().Get("path"))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := createTestToolboxClient(server)
	fs := NewFileSystemService(client, nil)

	ctx := context.Background()
	err := fs.CreateFolder(ctx, "/home/user/mydir")
	assert.NoError(t, err)
}

func TestCreateFolderWithMode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "0700", r.URL.Query().Get("mode"))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := createTestToolboxClient(server)
	fs := NewFileSystemService(client, nil)

	ctx := context.Background()
	err := fs.CreateFolder(ctx, "/home/user/private", options.WithMode("0700"))
	assert.NoError(t, err)
}

func TestListFilesError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"message": "path not found"})
	}))
	defer server.Close()

	client := createTestToolboxClient(server)
	fs := NewFileSystemService(client, nil)

	ctx := context.Background()
	_, err := fs.ListFiles(ctx, "/nonexistent")
	require.Error(t, err)
}

func TestGetFileInfoError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"message": "file not found"})
	}))
	defer server.Close()

	client := createTestToolboxClient(server)
	fs := NewFileSystemService(client, nil)

	ctx := context.Background()
	_, err := fs.GetFileInfo(ctx, "/nonexistent/file.txt")
	require.Error(t, err)
}

func TestDeleteFile(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		recursive bool
	}{
		{name: "delete file", path: "/home/user/file.txt", recursive: false},
		{name: "delete dir recursively", path: "/home/user/mydir", recursive: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			client := createTestToolboxClient(server)
			fs := NewFileSystemService(client, nil)

			ctx := context.Background()
			err := fs.DeleteFile(ctx, tt.path, tt.recursive)
			assert.NoError(t, err)
		})
	}
}

func TestDownloadFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write([]byte("file content here"))
	}))
	defer server.Close()

	client := createTestToolboxClient(server)
	fs := NewFileSystemService(client, nil)

	ctx := context.Background()
	data, err := fs.DownloadFile(ctx, "/home/user/file.txt", nil)
	require.NoError(t, err)
	assert.Equal(t, []byte("file content here"), data)
}

func TestMoveFiles(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := createTestToolboxClient(server)
	fs := NewFileSystemService(client, nil)

	ctx := context.Background()
	err := fs.MoveFiles(ctx, "/home/user/old.txt", "/home/user/new.txt")
	assert.NoError(t, err)
}

func TestSearchFilesError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "search failed"})
	}))
	defer server.Close()

	client := createTestToolboxClient(server)
	fs := NewFileSystemService(client, nil)

	ctx := context.Background()
	_, err := fs.SearchFiles(ctx, "/home/user", "*.go")
	require.Error(t, err)
}

func TestFindFilesError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "find failed"})
	}))
	defer server.Close()

	client := createTestToolboxClient(server)
	fs := NewFileSystemService(client, nil)

	ctx := context.Background()
	_, err := fs.FindFiles(ctx, "/home/user/project", "TODO:")
	require.Error(t, err)
}

func TestReplaceInFilesError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "replace failed"})
	}))
	defer server.Close()

	client := createTestToolboxClient(server)
	fs := NewFileSystemService(client, nil)

	ctx := context.Background()
	_, err := fs.ReplaceInFiles(ctx, []string{"/home/user/file1.txt"}, "old", "new")
	require.Error(t, err)
}

func TestSetFilePermissions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := createTestToolboxClient(server)
	fs := NewFileSystemService(client, nil)

	ctx := context.Background()
	err := fs.SetFilePermissions(ctx, "/home/user/script.sh",
		options.WithPermissionMode("0755"),
		options.WithOwner("root"),
		options.WithGroup("users"),
	)
	assert.NoError(t, err)
}

func TestUploadFileFromBytes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{})
	}))
	defer server.Close()

	client := createTestToolboxClient(server)
	fs := NewFileSystemService(client, nil)

	ctx := context.Background()
	err := fs.UploadFile(ctx, []byte("hello world"), "/home/user/hello.txt")
	assert.NoError(t, err)
}

func TestUploadFileInvalidSource(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := createTestToolboxClient(server)
	fs := NewFileSystemService(client, nil)

	ctx := context.Background()
	err := fs.UploadFile(ctx, 12345, "/home/user/file.txt")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid source type")
}

func TestFileSystemErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"message": "file not found"})
	}))
	defer server.Close()

	client := createTestToolboxClient(server)
	fs := NewFileSystemService(client, nil)

	ctx := context.Background()
	_, err := fs.ListFiles(ctx, "/nonexistent")
	require.Error(t, err)
}
