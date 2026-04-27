// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package daytona

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

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
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "path not found"})
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
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "file not found"})
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
		_, _ = w.Write([]byte("file content here"))
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
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "search failed"})
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
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "find failed"})
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
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "replace failed"})
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
		_ = json.NewEncoder(w).Encode(map[string]interface{}{})
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
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "file not found"})
	}))
	defer server.Close()

	client := createTestToolboxClient(server)
	fs := NewFileSystemService(client, nil)

	ctx := context.Background()
	_, err := fs.ListFiles(ctx, "/nonexistent")
	require.Error(t, err)
}

func TestFileSystemListAndInfoConversions(t *testing.T) {
	t.Run("list files parses timestamps and directory flags", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			writeJSONResponse(t, w, http.StatusOK, []map[string]any{{
				"group":       "staff",
				"name":        "main.go",
				"owner":       "daytona",
				"permissions": "rw-r--r--",
				"size":        12,
				"mode":        "0644",
				"modTime":     time.Now().UTC().Format(time.RFC3339),
				"isDir":       false,
			}, {
				"group":       "staff",
				"name":        "nested",
				"owner":       "daytona",
				"permissions": "rwxr-xr-x",
				"size":        0,
				"mode":        "0755",
				"modTime":     "not-a-timestamp",
				"isDir":       true,
			}})
		}))
		defer server.Close()

		fs := NewFileSystemService(createTestToolboxClient(server), nil)
		files, err := fs.ListFiles(context.Background(), "/workspace")
		require.NoError(t, err)
		require.Len(t, files, 2)
		assert.Equal(t, "main.go", files[0].Name)
		assert.True(t, files[1].IsDirectory)
		assert.True(t, files[1].ModifiedTime.IsZero())
	})

	t.Run("get file info maps payload", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			writeJSONResponse(t, w, http.StatusOK, map[string]any{
				"group":       "staff",
				"name":        "README.md",
				"owner":       "daytona",
				"permissions": "rw-r--r--",
				"size":        44,
				"mode":        "0644",
				"modTime":     time.Now().UTC().Format(time.RFC3339),
				"isDir":       false,
			})
		}))
		defer server.Close()

		fs := NewFileSystemService(createTestToolboxClient(server), nil)
		info, err := fs.GetFileInfo(context.Background(), "/workspace/README.md")
		require.NoError(t, err)
		assert.Equal(t, int64(44), info.Size)
		assert.False(t, info.IsDirectory)
	})
}

func TestFileSystemDownloadAndUploadEdgeCases(t *testing.T) {
	t.Run("download writes to local path", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("persist me"))
		}))
		defer server.Close()

		fs := NewFileSystemService(createTestToolboxClient(server), nil)
		localPath := filepath.Join(t.TempDir(), "download.txt")
		data, err := fs.DownloadFile(context.Background(), "/remote.txt", &localPath)
		require.NoError(t, err)
		assert.Equal(t, []byte("persist me"), data)
		written, err := os.ReadFile(localPath)
		require.NoError(t, err)
		assert.Equal(t, "persist me", string(written))
	})

	t.Run("download returns write error for invalid local path", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("content"))
		}))
		defer server.Close()

		fs := NewFileSystemService(createTestToolboxClient(server), nil)
		invalidPath := filepath.Join(t.TempDir(), "missing-parent", "file.txt")
		_, err := fs.DownloadFile(context.Background(), "/remote.txt", &invalidPath)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Failed to write file")
	})

	t.Run("upload from file path", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		source := filepath.Join(t.TempDir(), "source.txt")
		require.NoError(t, os.WriteFile(source, []byte("hello from disk"), 0o644))
		fs := NewFileSystemService(createTestToolboxClient(server), nil)
		require.NoError(t, fs.UploadFile(context.Background(), source, "/workspace/source.txt"))
	})

	t.Run("upload from missing file path errors", func(t *testing.T) {
		server := httptest.NewServer(http.NotFoundHandler())
		defer server.Close()
		fs := NewFileSystemService(createTestToolboxClient(server), nil)
		err := fs.UploadFile(context.Background(), "/does/not/exist.txt", "/workspace/missing.txt")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Failed to read file")
	})
}

func TestFileSystemSearchAndReplaceSuccessMappings(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "search"):
			writeJSONResponse(t, w, http.StatusOK, map[string]any{"files": []string{"/workspace/main.go"}})
		case strings.Contains(r.URL.Path, "find"):
			writeJSONResponse(t, w, http.StatusOK, []map[string]any{{"file": "/workspace/main.go", "line": 10, "content": "TODO"}})
		case strings.Contains(r.URL.Path, "replace"):
			writeJSONResponse(t, w, http.StatusOK, []map[string]any{{"file": "/workspace/main.go", "success": true}, {"file": "/workspace/other.go", "success": false, "error": "permission denied"}})
		default:
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	fs := NewFileSystemService(createTestToolboxClient(server), nil)
	search, err := fs.SearchFiles(context.Background(), "/workspace", "*.go")
	require.NoError(t, err)
	assert.Equal(t, []string{"/workspace/main.go"}, toStringSlice(search.(map[string]any)["files"]))
	find, err := fs.FindFiles(context.Background(), "/workspace", "TODO")
	require.NoError(t, err)
	assert.Len(t, find.([]map[string]any), 1)
	replaced, err := fs.ReplaceInFiles(context.Background(), []string{"/workspace/main.go"}, "TODO", "DONE")
	require.NoError(t, err)
	results := replaced.([]map[string]any)
	assert.Equal(t, "permission denied", results[1]["error"])
}
