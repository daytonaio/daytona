// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package fs

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestDirectory creates a temporary directory structure for testing glob patterns
func setupTestDirectory(t *testing.T) string {
	tempDir, err := os.MkdirTemp("", "search_files_test")
	require.NoError(t, err)

	// Create directory structure:
	// tempDir/
	// ├── file1.txt
	// ├── file2.go
	// ├── src/
	// │   ├── main.go
	// │   ├── utils.go
	// │   └── components/
	// │       ├── button.tsx
	// │       └── accordion.tsx
	// └── docs/
	//     └── readme.md

	// Create directories
	require.NoError(t, os.MkdirAll(filepath.Join(tempDir, "src", "components"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(tempDir, "docs"), 0755))

	// Create files
	files := []string{
		"file1.txt",
		"file2.go",
		"src/main.go",
		"src/utils.go",
		"src/components/button.tsx",
		"src/components/accordion.tsx",
		"docs/readme.md",
	}

	for _, file := range files {
		filePath := filepath.Join(tempDir, file)
		require.NoError(t, os.WriteFile(filePath, []byte("test content"), 0644))
	}

	return tempDir
}

// cleanupTestDirectory removes the temporary directory
func cleanupTestDirectory(t *testing.T, dir string) {
	require.NoError(t, os.RemoveAll(dir))
}

// makeRequest creates a test request and returns the response
func makeRequest(t *testing.T, path, pattern string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/files/search", SearchFiles)

	req, err := http.NewRequest("GET", "/files/search?path="+path+"&pattern="+pattern, nil)
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	return recorder
}

func TestSearchFiles_DoubleStarPattern(t *testing.T) {
	tempDir := setupTestDirectory(t)
	defer cleanupTestDirectory(t, tempDir)

	// Test **/*.tsx - should find all .tsx files recursively
	recorder := makeRequest(t, tempDir, "**/*.tsx")
	assert.Equal(t, http.StatusOK, recorder.Code)

	var response SearchFilesResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	require.NoError(t, err)

	// Should find button.tsx and accordion.tsx
	assert.Len(t, response.Files, 2)

	// Verify the files are the expected ones
	fileNames := make([]string, len(response.Files))
	for i, f := range response.Files {
		fileNames[i] = filepath.Base(f)
	}
	assert.Contains(t, fileNames, "button.tsx")
	assert.Contains(t, fileNames, "accordion.tsx")
}

func TestSearchFiles_DoubleStarWithSpecificFile(t *testing.T) {
	tempDir := setupTestDirectory(t)
	defer cleanupTestDirectory(t, tempDir)

	// Test **/accordion.tsx - should find accordion.tsx in any subdirectory
	recorder := makeRequest(t, tempDir, "**/accordion.tsx")
	assert.Equal(t, http.StatusOK, recorder.Code)

	var response SearchFilesResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	require.NoError(t, err)

	// Should find exactly one file
	assert.Len(t, response.Files, 1)
	assert.Equal(t, "accordion.tsx", filepath.Base(response.Files[0]))
}

func TestSearchFiles_SingleStarPattern(t *testing.T) {
	tempDir := setupTestDirectory(t)
	defer cleanupTestDirectory(t, tempDir)

	// Test *.txt - should find only .txt files in root
	recorder := makeRequest(t, tempDir, "*.txt")
	assert.Equal(t, http.StatusOK, recorder.Code)

	var response SearchFilesResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	require.NoError(t, err)

	// Should find file1.txt
	assert.Len(t, response.Files, 1)
	assert.Equal(t, "file1.txt", filepath.Base(response.Files[0]))
}

func TestSearchFiles_DoubleStarAllGoFiles(t *testing.T) {
	tempDir := setupTestDirectory(t)
	defer cleanupTestDirectory(t, tempDir)

	// Test **/*.go - should find all .go files recursively
	recorder := makeRequest(t, tempDir, "**/*.go")
	assert.Equal(t, http.StatusOK, recorder.Code)

	var response SearchFilesResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	require.NoError(t, err)

	// Should find file2.go, main.go, utils.go
	assert.Len(t, response.Files, 3)
}

func TestSearchFiles_PathPrefixPattern(t *testing.T) {
	tempDir := setupTestDirectory(t)
	defer cleanupTestDirectory(t, tempDir)

	// Test src/**/*.go - should find .go files only in src directory
	recorder := makeRequest(t, tempDir, "src/**/*.go")
	assert.Equal(t, http.StatusOK, recorder.Code)

	var response SearchFilesResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	require.NoError(t, err)

	// Should find main.go and utils.go (not file2.go which is in root)
	assert.Len(t, response.Files, 2)

	fileNames := make([]string, len(response.Files))
	for i, f := range response.Files {
		fileNames[i] = filepath.Base(f)
	}
	assert.Contains(t, fileNames, "main.go")
	assert.Contains(t, fileNames, "utils.go")
	assert.NotContains(t, fileNames, "file2.go")
}

func TestSearchFiles_QuestionMarkWildcard(t *testing.T) {
	tempDir := setupTestDirectory(t)
	defer cleanupTestDirectory(t, tempDir)

	// Test file?.txt - should match file1.txt (single character wildcard)
	recorder := makeRequest(t, tempDir, "file?.txt")
	assert.Equal(t, http.StatusOK, recorder.Code)

	var response SearchFilesResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Len(t, response.Files, 1)
	assert.Equal(t, "file1.txt", filepath.Base(response.Files[0]))
}

func TestSearchFiles_NoMatches(t *testing.T) {
	tempDir := setupTestDirectory(t)
	defer cleanupTestDirectory(t, tempDir)

	// Test pattern that matches nothing
	recorder := makeRequest(t, tempDir, "**/*.xyz")
	assert.Equal(t, http.StatusOK, recorder.Code)

	var response SearchFilesResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	require.NoError(t, err)

	// Should return empty array, not null
	assert.NotNil(t, response.Files)
	assert.Len(t, response.Files, 0)
}

func TestSearchFiles_MissingParameters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/files/search", SearchFiles)

	// Test missing path
	req, _ := http.NewRequest("GET", "/files/search?pattern=*.txt", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	// Test missing pattern
	req, _ = http.NewRequest("GET", "/files/search?path=/tmp", nil)
	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	// Test both missing
	req, _ = http.NewRequest("GET", "/files/search", nil)
	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestSearchFiles_BraceExpansion(t *testing.T) {
	tempDir := setupTestDirectory(t)
	defer cleanupTestDirectory(t, tempDir)

	// Test {a,b} group patterns - should match .go and .tsx files
	recorder := makeRequest(t, tempDir, "**/*.{go,tsx}")
	assert.Equal(t, http.StatusOK, recorder.Code)

	var response SearchFilesResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	require.NoError(t, err)

	// Should find all .go files (3) and all .tsx files (2) = 5 total
	assert.Len(t, response.Files, 5)
}
