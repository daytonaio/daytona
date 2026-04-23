// Copyright Daytona Platforms Inc.
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

func TestGitServiceCreation(t *testing.T) {
	gs := NewGitService(nil, nil)
	require.NotNil(t, gs)
}

func TestGitClone(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := createTestToolboxClient(server)
	git := NewGitService(client, nil)

	ctx := context.Background()
	err := git.Clone(ctx, "https://github.com/user/repo.git", "/home/user/repo")
	assert.NoError(t, err)
}

func TestGitCloneWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := createTestToolboxClient(server)
	git := NewGitService(client, nil)

	ctx := context.Background()
	err := git.Clone(ctx, "https://github.com/user/repo.git", "/home/user/repo",
		options.WithBranch("develop"),
		options.WithUsername("user"),
		options.WithPassword("token"),
	)
	assert.NoError(t, err)
}

func TestGitCloneWithCommitId(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := createTestToolboxClient(server)
	git := NewGitService(client, nil)

	ctx := context.Background()
	err := git.Clone(ctx, "https://github.com/user/repo.git", "/home/user/repo",
		options.WithCommitId("abc123"),
	)
	assert.NoError(t, err)
}

func TestGitStatusError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "not a git repo"})
	}))
	defer server.Close()

	client := createTestToolboxClient(server)
	git := NewGitService(client, nil)

	ctx := context.Background()
	_, err := git.Status(ctx, "/not/a/repo")
	require.Error(t, err)
}

func TestGitAdd(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := createTestToolboxClient(server)
	git := NewGitService(client, nil)

	ctx := context.Background()
	err := git.Add(ctx, "/home/user/repo", []string{"."})
	assert.NoError(t, err)
}

func TestGitCommitError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "commit failed"})
	}))
	defer server.Close()

	client := createTestToolboxClient(server)
	git := NewGitService(client, nil)

	ctx := context.Background()
	_, err := git.Commit(ctx, "/home/user/repo", "commit", "John", "john@example.com")
	require.Error(t, err)
}

func TestGitBranchesError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "list failed"})
	}))
	defer server.Close()

	client := createTestToolboxClient(server)
	git := NewGitService(client, nil)

	ctx := context.Background()
	_, err := git.Branches(ctx, "/home/user/repo")
	require.Error(t, err)
}

func TestGitCheckout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := createTestToolboxClient(server)
	git := NewGitService(client, nil)

	ctx := context.Background()
	err := git.Checkout(ctx, "/home/user/repo", "develop")
	assert.NoError(t, err)
}

func TestGitCreateBranch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := createTestToolboxClient(server)
	git := NewGitService(client, nil)

	ctx := context.Background()
	err := git.CreateBranch(ctx, "/home/user/repo", "feature/new-feature")
	assert.NoError(t, err)
}

func TestGitDeleteBranch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := createTestToolboxClient(server)
	git := NewGitService(client, nil)

	ctx := context.Background()
	err := git.DeleteBranch(ctx, "/home/user/repo", "feature/old-feature")
	assert.NoError(t, err)
}

func TestGitDeleteBranchWithForce(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := createTestToolboxClient(server)
	git := NewGitService(client, nil)

	ctx := context.Background()
	err := git.DeleteBranch(ctx, "/home/user/repo", "feature/abandoned",
		options.WithForce(true),
	)
	assert.NoError(t, err)
}

func TestGitPush(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := createTestToolboxClient(server)
	git := NewGitService(client, nil)

	ctx := context.Background()
	err := git.Push(ctx, "/home/user/repo")
	assert.NoError(t, err)
}

func TestGitPushWithAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := createTestToolboxClient(server)
	git := NewGitService(client, nil)

	ctx := context.Background()
	err := git.Push(ctx, "/home/user/repo",
		options.WithPushUsername("username"),
		options.WithPushPassword("github_token"),
	)
	assert.NoError(t, err)
}

func TestGitPull(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := createTestToolboxClient(server)
	git := NewGitService(client, nil)

	ctx := context.Background()
	err := git.Pull(ctx, "/home/user/repo")
	assert.NoError(t, err)
}

func TestGitPullWithAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := createTestToolboxClient(server)
	git := NewGitService(client, nil)

	ctx := context.Background()
	err := git.Pull(ctx, "/home/user/repo",
		options.WithPullUsername("username"),
		options.WithPullPassword("github_token"),
	)
	assert.NoError(t, err)
}

func TestStatusToCode(t *testing.T) {
	tests := []struct {
		status   toolbox.Status
		expected rune
	}{
		{toolbox.STATUS_Unmodified, ' '},
		{toolbox.STATUS_Modified, 'M'},
		{toolbox.STATUS_Added, 'A'},
		{toolbox.STATUS_Deleted, 'D'},
		{toolbox.STATUS_Renamed, 'R'},
		{toolbox.STATUS_Copied, 'C'},
		{toolbox.STATUS_Untracked, '?'},
		{toolbox.STATUS_UpdatedButUnmerged, 'U'},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			assert.Equal(t, tt.expected, statusToCode(tt.status))
		})
	}
}

func TestConvertFileStatus(t *testing.T) {
	input := []toolbox.FileStatus{
		{
			Name:     "main.go",
			Staging:  toolbox.STATUS_Modified,
			Worktree: toolbox.STATUS_Unmodified,
		},
	}

	result := convertFileStatus(input)
	require.Len(t, result, 1)
	assert.Equal(t, "main.go", result[0].Path)
	assert.Equal(t, "M ", result[0].Status)
}

func TestGitErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "internal error"})
	}))
	defer server.Close()

	client := createTestToolboxClient(server)
	git := NewGitService(client, nil)

	ctx := context.Background()
	_, err := git.Status(ctx, "/home/user/repo")
	require.Error(t, err)
}

func TestGitStatusAndCommitMappings(t *testing.T) {
	t.Run("status maps ahead behind and file states", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			writeJSONResponse(t, w, http.StatusOK, map[string]any{
				"currentBranch":   "main",
				"ahead":           2,
				"behind":          1,
				"branchPublished": true,
				"fileStatus": []map[string]any{{
					"extra":    "",
					"name":     "main.go",
					"staging":  toolbox.STATUS_Modified,
					"worktree": toolbox.STATUS_Untracked,
				}},
			})
		}))
		defer server.Close()

		git := NewGitService(createTestToolboxClient(server), nil)
		status, err := git.Status(context.Background(), "/repo")
		require.NoError(t, err)
		assert.Equal(t, "main", status.CurrentBranch)
		assert.Equal(t, 2, status.Ahead)
		assert.Equal(t, "M?", status.FileStatus[0].Status)
	})

	t.Run("commit maps hash response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
			assert.Equal(t, "message", body["message"])
			writeJSONResponse(t, w, http.StatusOK, map[string]any{"hash": "abc123"})
		}))
		defer server.Close()

		git := NewGitService(createTestToolboxClient(server), nil)
		resp, err := git.Commit(context.Background(), "/repo", "message", "Author", "author@example.com", options.WithAllowEmpty(true))
		require.NoError(t, err)
		assert.Equal(t, "abc123", resp.SHA)
	})
}

func TestGitBranchAndRemoteOperationsRequests(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			writeJSONResponse(t, w, http.StatusOK, map[string]any{"branches": []string{"main", "dev"}})
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	git := NewGitService(createTestToolboxClient(server), nil)
	ctx := context.Background()
	branches, err := git.Branches(ctx, "/repo")
	require.NoError(t, err)
	assert.Equal(t, []string{"main", "dev"}, branches)
	require.NoError(t, git.Checkout(ctx, "/repo", "dev"))
	require.NoError(t, git.CreateBranch(ctx, "/repo", "feature/x"))
	require.NoError(t, git.DeleteBranch(ctx, "/repo", "feature/x", options.WithForce(true)))
	require.NoError(t, git.Push(ctx, "/repo", options.WithPushUsername("user"), options.WithPushPassword("pass")))
	require.NoError(t, git.Pull(ctx, "/repo", options.WithPullUsername("user"), options.WithPullPassword("pass")))
}

func TestStatusToCodeDefaultFallback(t *testing.T) {
	assert.Equal(t, '?', statusToCode(toolbox.Status("unexpected")))
}
