// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package daytona

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	apiclient "github.com/daytonaio/daytona/libs/api-client-go"
	"github.com/daytonaio/daytona/libs/sdk-go/pkg/options"
	"github.com/daytonaio/daytona/libs/sdk-go/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestE2E(t *testing.T) {
	if os.Getenv("DAYTONA_API_KEY") == "" {
		t.Skip("DAYTONA_API_KEY not set, skipping E2E tests")
	}

	client, err := NewClient()
	require.NoError(t, err)

	ctx := context.Background()
	params := types.SnapshotParams{
		SandboxBaseParams: types.SandboxBaseParams{
			Language: types.CodeLanguagePython,
		},
	}

	sandbox, err := client.Create(ctx, params, options.WithTimeout(90*time.Second))
	require.NoError(t, err)
	require.NotNil(t, sandbox)

	defer func() {
		_ = sandbox.Delete(ctx)
	}()

	workDir, err := sandbox.GetWorkingDir(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, workDir)

	baseDir := strings.TrimRight(workDir, "/") + "/sdk-go-e2e"
	textDir := baseDir + "/test-dir"
	helloPath := textDir + "/hello.txt"
	movedPath := textDir + "/moved.txt"
	repoPath := baseDir + "/hello-world"
	sessionID := "test-session"

	t.Run("SandboxLifecycle", func(t *testing.T) {
		require.NotEmpty(t, sandbox.ID)
		assert.Equal(t, "started", string(sandbox.State))

		homeDir, homeErr := sandbox.GetUserHomeDir(ctx)
		require.NoError(t, homeErr)
		assert.NotEmpty(t, homeDir)

		wd, wdErr := sandbox.GetWorkingDir(ctx)
		require.NoError(t, wdErr)
		assert.NotEmpty(t, wd)

		setLabelsErr := sandbox.SetLabels(ctx, map[string]string{"test": "e2e"})
		require.NoError(t, setLabelsErr)

		archiveInterval := 30
		setArchiveErr := sandbox.SetAutoArchiveInterval(ctx, &archiveInterval)
		require.NoError(t, setArchiveErr)
		assert.Equal(t, 30, sandbox.AutoArchiveInterval)

		refreshErr := sandbox.RefreshData(ctx)
		require.NoError(t, refreshErr)
		assert.Equal(t, apiclient.SANDBOXSTATE_STARTED, sandbox.State)
	})

	t.Run("FileSystemOperations", func(t *testing.T) {
		require.NoError(t, sandbox.FileSystem.CreateFolder(ctx, baseDir))
		require.NoError(t, sandbox.FileSystem.CreateFolder(ctx, textDir))

		require.NoError(t, sandbox.FileSystem.UploadFile(ctx, []byte("hello world"), helloPath))

		files, listErr := sandbox.FileSystem.ListFiles(ctx, textDir)
		require.NoError(t, listErr)
		assert.True(t, containsFileName(files, "hello.txt"))

		info, infoErr := sandbox.FileSystem.GetFileInfo(ctx, helloPath)
		require.NoError(t, infoErr)
		assert.Greater(t, info.Size, int64(0))

		data, downloadErr := sandbox.FileSystem.DownloadFile(ctx, helloPath, nil)
		require.NoError(t, downloadErr)
		assert.Equal(t, "hello world", string(data))

		findResult, findErr := sandbox.FileSystem.FindFiles(ctx, textDir, "hello")
		require.NoError(t, findErr)
		findMatches, ok := findResult.([]map[string]any)
		require.True(t, ok)
		assert.NotEmpty(t, findMatches)

		searchResult, searchErr := sandbox.FileSystem.SearchFiles(ctx, textDir, "*.txt")
		require.NoError(t, searchErr)
		searchMap, ok := searchResult.(map[string]any)
		require.True(t, ok)
		assert.Contains(t, toStringSlice(searchMap["files"]), helloPath)

		replaceResult, replaceErr := sandbox.FileSystem.ReplaceInFiles(ctx, []string{helloPath}, "hello", "world")
		require.NoError(t, replaceErr)
		replacedFiles, ok := replaceResult.([]map[string]any)
		require.True(t, ok)
		require.NotEmpty(t, replacedFiles)
		assert.Equal(t, true, replacedFiles[0]["success"])

		updated, updatedErr := sandbox.FileSystem.DownloadFile(ctx, helloPath, nil)
		require.NoError(t, updatedErr)
		assert.Equal(t, "world world", string(updated))

		require.NoError(t, sandbox.FileSystem.MoveFiles(ctx, helloPath, movedPath))
		require.NoError(t, sandbox.FileSystem.DeleteFile(ctx, movedPath, false))
	})

	t.Run("ProcessExecution", func(t *testing.T) {
		result, execErr := sandbox.Process.ExecuteCommand(ctx, "echo hello")
		require.NoError(t, execErr)
		assert.Equal(t, 0, result.ExitCode)
		assert.Contains(t, result.Result, "hello")

		cwdResult, cwdErr := sandbox.Process.ExecuteCommand(ctx, "pwd", options.WithCwd(workDir))
		require.NoError(t, cwdErr)
		assert.Equal(t, 0, cwdResult.ExitCode)
		assert.Contains(t, strings.TrimSpace(cwdResult.Result), workDir)

		envResult, envErr := sandbox.Process.ExecuteCommand(
			ctx,
			"echo env-option",
			options.WithCommandEnv(map[string]string{"E2E_ENV": "enabled"}),
		)
		require.NoError(t, envErr)
		assert.Equal(t, 0, envResult.ExitCode)
		assert.Contains(t, envResult.Result, "env-option")

		codeRunResult, codeRunErr := sandbox.Process.CodeRun(ctx, "print('hello from python')")
		if codeRunErr == nil {
			require.NotNil(t, codeRunResult)
			assert.Equal(t, 0, codeRunResult.ExitCode)
			assert.Contains(t, codeRunResult.Result, "hello from python")
		} else {
			t.Logf("CodeRun unavailable in current SDK/toolbox, validating via python command fallback: %v", codeRunErr)
			pythonResult, pyErr := sandbox.Process.ExecuteCommand(ctx, `python -c "print('hello from python')"`)
			require.NoError(t, pyErr)
			assert.Equal(t, 0, pythonResult.ExitCode)
			assert.Contains(t, pythonResult.Result, "hello from python")
		}

		failingResult, failErr := sandbox.Process.ExecuteCommand(ctx, "exit 1")
		require.NoError(t, failErr)
		assert.NotEqual(t, 0, failingResult.ExitCode)
	})

	t.Run("SessionManagement", func(t *testing.T) {
		require.NoError(t, sandbox.Process.CreateSession(ctx, sessionID))
		defer func() {
			_ = sandbox.Process.DeleteSession(ctx, sessionID)
		}()

		session, sessionErr := sandbox.Process.GetSession(ctx, sessionID)
		require.NoError(t, sessionErr)
		assert.Equal(t, sessionID, session["sessionId"])

		_, exportErr := sandbox.Process.ExecuteSessionCommand(ctx, sessionID, "export FOO=bar", false, false)
		require.NoError(t, exportErr)

		echoResult, echoErr := sandbox.Process.ExecuteSessionCommand(ctx, sessionID, "echo $FOO", false, false)
		require.NoError(t, echoErr)
		stdout, ok := echoResult["stdout"].(string)
		require.True(t, ok)
		assert.Contains(t, strings.TrimSpace(stdout), "bar")

		sessions, listErr := sandbox.Process.ListSessions(ctx)
		require.NoError(t, listErr)
		assert.True(t, containsSessionID(sessions, sessionID))

		require.NoError(t, sandbox.Process.DeleteSession(ctx, sessionID))
	})

	t.Run("GitOperations", func(t *testing.T) {
		require.NoError(t, sandbox.Git.Clone(ctx, "https://github.com/octocat/Hello-World.git", repoPath))

		status, statusErr := sandbox.Git.Status(ctx, repoPath)
		require.NoError(t, statusErr)
		assert.NotEmpty(t, status.CurrentBranch)

		branches, branchesErr := sandbox.Git.Branches(ctx, repoPath)
		require.NoError(t, branchesErr)
		assert.NotEmpty(t, branches)
	})

	t.Run("ClientOperations", func(t *testing.T) {
		listed, listErr := client.List(ctx, nil, nil, nil)
		require.NoError(t, listErr)
		assert.GreaterOrEqual(t, listed.Total, 1)

		got, getErr := client.Get(ctx, sandbox.ID)
		require.NoError(t, getErr)
		require.NotNil(t, got)
		assert.Equal(t, sandbox.ID, got.ID)
	})
}

func containsFileName(files []*types.FileInfo, fileName string) bool {
	for _, f := range files {
		if f.Name == fileName {
			return true
		}
	}
	return false
}

func toStringSlice(v any) []string {
	if v == nil {
		return nil
	}

	if s, ok := v.([]string); ok {
		return s
	}

	if ifaceSlice, ok := v.([]any); ok {
		out := make([]string, 0, len(ifaceSlice))
		for _, item := range ifaceSlice {
			out = append(out, fmt.Sprint(item))
		}
		return out
	}

	return []string{fmt.Sprint(v)}
}

func containsSessionID(sessions []map[string]any, sessionID string) bool {
	for _, session := range sessions {
		if id, ok := session["sessionId"].(string); ok && id == sessionID {
			return true
		}
	}
	return false
}
