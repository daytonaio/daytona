// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

// Helper functions for creating pointers
func strPtr(s string) *string {
	return &s
}

func uint32Ptr(i uint32) *uint32 {
	return &i
}

func TestGiteeCanHandle(t *testing.T) {
	provider := NewGiteeGitProvider("", "")

	tests := []struct {
		name    string
		url     string
		want    bool
		wantErr bool
	}{
		{
			name:    "valid Gitee HTTPS URL",
			url:     "https://gitee.com/user/repo",
			want:    true,
			wantErr: false,
		},
		{
			name:    "valid Gitee SSH URL",
			url:     "git@gitee.com:user/repo.git",
			want:    true,
			wantErr: false,
		},
		{
			name:    "invalid GitHub URL",
			url:     "https://github.com/user/repo",
			want:    false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := provider.CanHandle(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("GiteeGitProvider.CanHandle() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GiteeGitProvider.CanHandle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGiteeGetUser(t *testing.T) {
	// Create a test server that mimics Gitee's API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.URL.Path != "/user" {
			t.Errorf("Expected path /user, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "token test-token" {
			t.Errorf("Expected Authorization header 'token test-token', got %s", r.Header.Get("Authorization"))
		}

		// Return mock user data
		mockUser := struct {
			ID    int    `json:"id"`
			Login string `json:"login"`
			Name  string `json:"name"`
			Email string `json:"email"`
		}{
			ID:    12345,
			Login: "testuser",
			Name:  "Test User",
			Email: "test@example.com",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockUser)
	}))
	defer server.Close()

	// Create provider with test server URL
	provider := NewGiteeGitProvider("test-token", server.URL)

	// Test GetUser
	user, err := provider.GetUser()
	if err != nil {
		t.Fatalf("GetUser() error = %v", err)
	}

	// Verify user data
	if user.Id != "12345" {
		t.Errorf("Expected user ID 12345, got %s", user.Id)
	}
	if user.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got %s", user.Username)
	}
	if user.Name != "Test User" {
		t.Errorf("Expected name 'Test User', got %s", user.Name)
	}
	if user.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got %s", user.Email)
	}
}

func TestGiteeGetNamespaces(t *testing.T) {
	// Create a test server that mimics Gitee's API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/user":
			// Mock user response for personal namespace
			mockUser := struct {
				ID    int    `json:"id"`
				Login string `json:"login"`
				Name  string `json:"name"`
				Email string `json:"email"`
			}{
				ID:    12345,
				Login: "testuser",
				Name:  "Test User",
				Email: "test@example.com",
			}
			json.NewEncoder(w).Encode(mockUser)

		case "/user/orgs":
			// Verify pagination parameters
			if r.URL.Query().Get("page") != "1" {
				t.Errorf("Expected page=1, got %s", r.URL.Query().Get("page"))
			}
			if r.URL.Query().Get("per_page") != "10" {
				t.Errorf("Expected per_page=10, got %s", r.URL.Query().Get("per_page"))
			}

			// Mock organizations response
			mockOrgs := []struct {
				ID   int    `json:"id"`
				Path string `json:"path"`
				Name string `json:"name"`
			}{
				{ID: 1, Path: "org1", Name: "Organization 1"},
				{ID: 2, Path: "org2", Name: "Organization 2"},
			}
			json.NewEncoder(w).Encode(mockOrgs)
		}
	}))
	defer server.Close()

	// Create provider with test server URL
	provider := NewGiteeGitProvider("test-token", server.URL)

	// Test GetNamespaces
	namespaces, err := provider.GetNamespaces(ListOptions{Page: 1, PerPage: 10})
	if err != nil {
		t.Fatalf("GetNamespaces() error = %v", err)
	}

	// Verify namespaces
	if len(namespaces) != 3 { // 1 personal + 2 organizations
		t.Errorf("Expected 3 namespaces, got %d", len(namespaces))
	}

	// Verify personal namespace
	if namespaces[0].Id != personalNamespaceId {
		t.Errorf("Expected personal namespace ID %s, got %s", personalNamespaceId, namespaces[0].Id)
	}
	if namespaces[0].Name != "testuser" {
		t.Errorf("Expected personal namespace name 'testuser', got %s", namespaces[0].Name)
	}

	// Verify organization namespaces
	if namespaces[1].Id != "1" {
		t.Errorf("Expected org ID '1', got %s", namespaces[1].Id)
	}
	if namespaces[1].Name != "Organization 1" {
		t.Errorf("Expected org name 'Organization 1', got %s", namespaces[1].Name)
	}
}

func TestGiteeGetRepositories(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Verify authorization header
		if r.Header.Get("Authorization") != "token token" {
			t.Errorf("Expected Authorization header 'token token', got %s", r.Header.Get("Authorization"))
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Add API version prefix to match implementation
		path := r.URL.Path
		if !strings.HasPrefix(path, "/api/v5/") {
			path = "/api/v5" + path
		}

		// Log the path for debugging
		t.Logf("Request path: %s", path)

		switch path {
		case "/api/v5/user/repos", "/api/v5/user":
			if path == "/api/v5/user" {
				// Return user info for personal namespace
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{
					"id": 1,
					"login": "user",
					"name": "Test User",
					"email": "user@example.com"
				}`))
				return
			}
			// Return repositories for personal namespace
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[
				{
					"id": 1,
					"name": "repo1",
					"html_url": "https://gitee.com/user/repo1",
					"ssh_url": "git@gitee.com:user/repo1.git",
					"default_branch": "main",
					"owner": {
						"login": "user"
					}
				},
				{
					"id": 2,
					"name": "repo2",
					"html_url": "https://gitee.com/user/repo2",
					"ssh_url": "git@gitee.com:user/repo2.git",
					"default_branch": "master",
					"owner": {
						"login": "user"
					}
				}
			]`))
			return
		case "/api/v5/orgs/testorg/repos":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[
				{
					"id": 3,
					"name": "repo3",
					"html_url": "https://gitee.com/testorg/repo3",
					"ssh_url": "git@gitee.com:testorg/repo3.git",
					"default_branch": "main",
					"owner": {
						"login": "testorg"
					}
				}
			]`))
			return
		default:
			t.Logf("Unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	provider := NewGiteeGitProvider("token", server.URL)

	// Test personal repositories
	t.Run("personal repositories", func(t *testing.T) {
		repos, err := provider.GetRepositories(personalNamespaceId, ListOptions{})
		if err != nil {
			t.Errorf("GetRepositories() error = %v", err)
			return
		}

		if len(repos) != 2 {
			t.Errorf("Expected 2 repositories, got %d", len(repos))
		}

		if repos[0].Name != "repo1" {
			t.Errorf("Expected repo name 'repo1', got %s", repos[0].Name)
		}
		if repos[0].Branch != "main" {
			t.Errorf("Expected branch 'main', got %s", repos[0].Branch)
		}
		if repos[0].Owner != "user" {
			t.Errorf("Expected owner 'user', got %s", repos[0].Owner)
		}
		if repos[0].Source != "gitee.com" {
			t.Errorf("Expected source 'gitee.com', got %s", repos[0].Source)
		}
		if repos[0].Url != "https://gitee.com/user/repo1" {
			t.Errorf("Expected url 'https://gitee.com/user/repo1', got %s", repos[0].Url)
		}
	})

	// Test organization repositories
	t.Run("organization repositories", func(t *testing.T) {
		repos, err := provider.GetRepositories("testorg", ListOptions{})
		if err != nil {
			t.Errorf("GetRepositories() error = %v", err)
			return
		}

		if len(repos) != 1 {
			t.Errorf("Expected 1 repository, got %d", len(repos))
		}

		if repos[0].Name != "repo3" {
			t.Errorf("Expected repo name 'repo3', got %s", repos[0].Name)
		}
		if repos[0].Owner != "testorg" {
			t.Errorf("Expected owner 'testorg', got %s", repos[0].Owner)
		}
	})
}

func TestGiteeParseStaticGitContext(t *testing.T) {
	provider := NewGiteeGitProvider("", "")

	tests := []struct {
		name    string
		url     string
		want    *StaticGitContext
		wantErr bool
	}{
		{
			name: "Basic repository URL",
			url:  "https://gitee.com/owner/repo1",
			want: &StaticGitContext{
				Source: "gitee.com",
				Owner:  "owner",
				Name:   "repo1",
				Url:    "https://gitee.com/owner/repo1",
			},
			wantErr: false,
		},
		{
			name: "SSH repository URL",
			url:  "git@gitee.com:owner/repo1.git",
			want: &StaticGitContext{
				Source: "gitee.com",
				Owner:  "owner",
				Name:   "repo1",
				Url:    "git@gitee.com:owner/repo1.git",
			},
			wantErr: false,
		},
		{
			name: "Repository URL with branch",
			url:  "https://gitee.com/owner/repo1/tree/dev",
			want: &StaticGitContext{
				Source: "gitee.com",
				Owner:  "owner",
				Name:   "repo1",
				Branch: strPtr("dev"),
				Url:    "https://gitee.com/owner/repo1/tree/dev",
			},
			wantErr: false,
		},
		{
			name: "Repository URL with commit",
			url:  "https://gitee.com/owner/repo1/commit/abc123",
			want: &StaticGitContext{
				Source: "gitee.com",
				Owner:  "owner",
				Name:   "repo1",
				Sha:    strPtr("abc123"),
				Url:    "https://gitee.com/owner/repo1/commit/abc123",
			},
			wantErr: false,
		},
		{
			name: "Repository URL with pull request",
			url:  "https://gitee.com/owner/repo1/pulls/42",
			want: &StaticGitContext{
				Source:   "gitee.com",
				Owner:    "owner",
				Name:     "repo1",
				PrNumber: uint32Ptr(42),
				Url:      "https://gitee.com/owner/repo1/pulls/42",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := provider.ParseStaticGitContext(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("GiteeGitProvider.ParseStaticGitContext() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GiteeGitProvider.ParseStaticGitContext() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGiteeGetUrlFromRepo(t *testing.T) {
	provider := NewGiteeGitProvider("", "")

	tests := []struct {
		name   string
		repo   *GitRepository
		branch string
		want   string
	}{
		{
			name: "URL without branch",
			repo: &GitRepository{
				Owner: "owner",
				Name:  "repo",
			},
			branch: "",
			want:   "https://gitee.com/owner/repo",
		},
		{
			name: "URL with branch",
			repo: &GitRepository{
				Owner: "owner",
				Name:  "repo",
			},
			branch: "main",
			want:   "https://gitee.com/owner/repo/tree/main",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := provider.GetUrlFromRepo(tt.repo, tt.branch)
			if got != tt.want {
				t.Errorf("GetUrlFromRepo() = %v, want %v", got, tt.want)
			}
		})
	}
}
