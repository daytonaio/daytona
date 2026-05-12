// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

//go:build e2e

package daytona

import (
	"context"
	"encoding/json"
	"fmt"
	apiclient "github.com/daytonaio/daytona/libs/api-client-go"
	"github.com/daytonaio/daytona/libs/sdk-go/pkg/options"
	"github.com/daytonaio/daytona/libs/sdk-go/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestE2E(t *testing.T) {
	if os.Getenv("DAYTONA_API_KEY") == "" {
		t.Fatal("DAYTONA_API_KEY environment variable is required for E2E tests")
	}

	client, err := NewClient()
	require.NoError(t, err)
	defer func() {
		_ = client.Close(context.Background())
	}()

	ctx := context.Background()
	sandbox, err := client.Create(ctx, types.SnapshotParams{
		SandboxBaseParams: types.SandboxBaseParams{
			Language: types.CodeLanguagePython,
			Labels: map[string]string{
				"suite": "sdk-go-e2e",
			},
		},
	}, options.WithTimeout(90*time.Second))
	require.NoError(t, err)
	require.NotNil(t, sandbox)

	defer func() {
		_ = sandbox.Delete(ctx)
	}()

	workDir, err := sandbox.GetWorkingDir(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, workDir)

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	baseDir := strings.TrimRight(workDir, "/") + "/sdk-go-e2e-" + suffix
	textDir := baseDir + "/test-dir"
	nestedDir := textDir + "/nested/deeper"
	helloPath := textDir + "/hello.txt"
	binaryPath := textDir + "/binary.bin"
	movedPath := textDir + "/moved.txt"
	repoPath := baseDir + "/repo-default"
	cloneBranchPath := baseDir + "/repo-branch"
	sessionID := "test-session-" + suffix
	ptySessionID := "test-pty-" + suffix
	volumeName := "sdk-go-e2e-vol-" + suffix
	lspProjectDir := baseDir + "/lsp-project"
	lspFilePath := lspProjectDir + "/sample.py"

	var (
		createdVolume   *types.Volume
		createdContexts []string
		sessionCmdID    string
		snapshotName    string
		lspServer       *LspServerService
	)

	t.Run("SandboxLifecycle/HasValidID", func(t *testing.T) {
		require.NotEmpty(t, sandbox.ID)
	})

	t.Run("SandboxLifecycle/StateIsStarted", func(t *testing.T) {
		assert.Equal(t, apiclient.SANDBOXSTATE_STARTED, sandbox.State)
	})

	t.Run("SandboxLifecycle/HasResources", func(t *testing.T) {
		res, execErr := sandbox.Process.ExecuteCommand(ctx, `nproc && grep MemTotal /proc/meminfo && df -k /`)
		require.NoError(t, execErr)
		assert.Equal(t, 0, res.ExitCode)
		assert.Contains(t, res.Result, "MemTotal")
	})

	t.Run("SandboxLifecycle/GetUserHomeDir", func(t *testing.T) {
		homeDir, homeErr := sandbox.GetUserHomeDir(ctx)
		require.NoError(t, homeErr)
		assert.NotEmpty(t, homeDir)
	})

	t.Run("SandboxLifecycle/GetWorkingDir", func(t *testing.T) {
		wd, wdErr := sandbox.GetWorkingDir(ctx)
		require.NoError(t, wdErr)
		assert.NotEmpty(t, wd)
	})

	t.Run("SandboxLifecycle/SetLabels", func(t *testing.T) {
		err = sandbox.SetLabels(ctx, map[string]string{"test": "e2e", "suite": "sdk-go"})
		require.NoError(t, err)
	})

	t.Run("SandboxLifecycle/SetAutoArchiveInterval", func(t *testing.T) {
		archiveInterval := 30
		err = sandbox.SetAutoArchiveInterval(ctx, &archiveInterval)
		require.NoError(t, err)
		assert.Equal(t, 30, sandbox.AutoArchiveInterval)
	})

	t.Run("SandboxLifecycle/SetAutoDeleteInterval", func(t *testing.T) {
		autoDelete := -1
		err = sandbox.SetAutoDeleteInterval(ctx, &autoDelete)
		require.NoError(t, err)
		assert.Equal(t, -1, sandbox.AutoDeleteInterval)
	})

	t.Run("SandboxLifecycle/RefreshData", func(t *testing.T) {
		refreshErr := sandbox.RefreshData(ctx)
		require.NoError(t, refreshErr)
		assert.Equal(t, apiclient.SANDBOXSTATE_STARTED, sandbox.State)
	})

	t.Run("DeclarativeImageBuild/CreateSandboxFromCustomImageWithBuildLogs", func(t *testing.T) {
		cacheKey := fmt.Sprintf("e2e-build-%d", time.Now().UnixNano())
		version := "3.12"
		image := DebianSlim(&version).
			PipInstall([]string{"numpy"}).
			Env("CACHE_BUSTER", cacheKey)

		imageSandbox, createErr := client.Create(ctx, types.ImageParams{
			SandboxBaseParams: types.SandboxBaseParams{
				Language: types.CodeLanguagePython,
				Name:     fmt.Sprintf("sdk-go-e2e-build-%d", time.Now().UnixNano()),
			},
			Image: image,
		}, options.WithTimeout(300*time.Second))
		require.NoError(t, createErr)
		require.NotNil(t, imageSandbox)
		defer func() {
			_ = imageSandbox.Delete(ctx)
		}()

		assert.Equal(t, apiclient.SANDBOXSTATE_STARTED, imageSandbox.State)

		result, execErr := imageSandbox.Process.ExecuteCommand(ctx, `python3 -c "import numpy; print(numpy.__version__)"`)
		require.NoError(t, execErr)
		assert.Equal(t, 0, result.ExitCode)
		assert.Regexp(t, `\d+\.\d+`, strings.TrimSpace(result.Result))
	})

	t.Run("SandboxLifecycle/StopAndStartCycle", func(t *testing.T) {
		require.NoError(t, sandbox.StopWithTimeout(ctx, 90*time.Second, false))
		require.NoError(t, sandbox.RefreshData(ctx))
		assert.Equal(t, apiclient.SANDBOXSTATE_STOPPED, sandbox.State)

		require.NoError(t, sandbox.StartWithTimeout(ctx, 90*time.Second))
		require.NoError(t, sandbox.RefreshData(ctx))
		assert.Equal(t, apiclient.SANDBOXSTATE_STARTED, sandbox.State)
	})

	t.Run("FileSystem/CreateFolderBase", func(t *testing.T) {
		require.NoError(t, sandbox.FileSystem.CreateFolder(ctx, baseDir))
	})

	t.Run("FileSystem/CreateFolder", func(t *testing.T) {
		require.NoError(t, sandbox.FileSystem.CreateFolder(ctx, textDir))
	})

	t.Run("FileSystem/NestedDirectoryOps", func(t *testing.T) {
		require.NoError(t, sandbox.FileSystem.CreateFolder(ctx, nestedDir, options.WithMode("0755")))
		files, listErr := sandbox.FileSystem.ListFiles(ctx, textDir)
		require.NoError(t, listErr)
		assert.True(t, containsFileName(files, "nested"))
	})

	t.Run("FileSystem/UploadFile", func(t *testing.T) {
		require.NoError(t, sandbox.FileSystem.UploadFile(ctx, []byte("hello world"), helloPath))
	})

	t.Run("FileSystem/UploadBinaryContent", func(t *testing.T) {
		binary := []byte{0x00, 0x01, 0x02, 0x7F, 0xFF}
		require.NoError(t, sandbox.FileSystem.UploadFile(ctx, binary, binaryPath))

		downloaded, downloadErr := sandbox.FileSystem.DownloadFile(ctx, binaryPath, nil)
		require.NoError(t, downloadErr)
		assert.Equal(t, binary, downloaded)
	})

	t.Run("FileSystem/ListFiles", func(t *testing.T) {
		files, listErr := sandbox.FileSystem.ListFiles(ctx, textDir)
		require.NoError(t, listErr)
		assert.True(t, containsFileName(files, "hello.txt"))
	})

	t.Run("FileSystem/GetFileInfo", func(t *testing.T) {
		info, infoErr := sandbox.FileSystem.GetFileInfo(ctx, helloPath)
		require.NoError(t, infoErr)
		assert.Greater(t, info.Size, int64(0))
		assert.False(t, info.IsDirectory)
	})

	t.Run("FileSystem/DownloadFile", func(t *testing.T) {
		localDir := t.TempDir()
		localPath := filepath.Join(localDir, "downloaded.txt")

		data, downloadErr := sandbox.FileSystem.DownloadFile(ctx, helloPath, &localPath)
		require.NoError(t, downloadErr)
		assert.Equal(t, "hello world", string(data))

		localData, localReadErr := os.ReadFile(localPath)
		require.NoError(t, localReadErr)
		assert.Equal(t, "hello world", string(localData))
	})

	t.Run("FileSystem/DownloadFileStream", func(t *testing.T) {
		stream, err := sandbox.FileSystem.DownloadFileStream(ctx, helloPath)
		require.NoError(t, err)
		defer stream.Close()

		data, err := io.ReadAll(stream)
		require.NoError(t, err)
		assert.Equal(t, "hello world", string(data))
	})

	t.Run("FileSystem/DownloadFileStreamProgress", func(t *testing.T) {
		var lastProgress DownloadProgress
		stream, err := sandbox.FileSystem.DownloadFileStream(ctx, helloPath, WithDownloadProgress(func(progress DownloadProgress) {
			lastProgress = progress
		}))
		require.NoError(t, err)
		defer stream.Close()

		data, err := io.ReadAll(stream)
		require.NoError(t, err)
		assert.Equal(t, "hello world", string(data))
		assert.Equal(t, int64(len(data)), lastProgress.BytesReceived)
		assert.Equal(t, int64(len(data)), lastProgress.TotalBytes)
	})

	t.Run("FileSystem/UploadFileStream", func(t *testing.T) {
		streamPath := textDir + "/upload-stream.bin"
		payload := strings.Repeat("upload-stream-content-", 1024)
		var lastProgress UploadProgress

		err := sandbox.FileSystem.UploadFileStream(
			ctx,
			strings.NewReader(payload),
			streamPath,
			WithUploadProgress(func(p UploadProgress) { lastProgress = p }),
		)
		require.NoError(t, err)

		data, err := sandbox.FileSystem.DownloadFile(ctx, streamPath, nil)
		require.NoError(t, err)
		assert.Equal(t, payload, string(data))
		assert.Equal(t, int64(len(payload)), lastProgress.BytesSent)
	})

	t.Run("FileSystem/UploadFileStreamCancellation", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(ctx)
		time.AfterFunc(100*time.Millisecond, cancel)

		err := sandbox.FileSystem.UploadFileStream(cancelCtx, endlessReader{}, textDir+"/upload-stream-cancel.bin")
		require.Error(t, err)
		assert.True(t, containsAny(strings.ToLower(err.Error()), "context canceled", "context deadline exceeded"))
	})

	t.Run("FileSystem/FindFiles", func(t *testing.T) {
		findResult, findErr := sandbox.FileSystem.FindFiles(ctx, textDir, "hello")
		require.NoError(t, findErr)
		findMatches, ok := findResult.([]map[string]any)
		require.True(t, ok)
		assert.NotEmpty(t, findMatches)
	})

	t.Run("FileSystem/SearchFiles", func(t *testing.T) {
		searchResult, searchErr := sandbox.FileSystem.SearchFiles(ctx, textDir, "*.txt")
		require.NoError(t, searchErr)
		searchMap, ok := searchResult.(map[string]any)
		require.True(t, ok)
		assert.Contains(t, toStringSlice(searchMap["files"]), helloPath)
	})

	t.Run("FileSystem/ReplaceInFiles", func(t *testing.T) {
		replaceResult, replaceErr := sandbox.FileSystem.ReplaceInFiles(ctx, []string{helloPath}, "hello", "world")
		require.NoError(t, replaceErr)
		replacedFiles, ok := replaceResult.([]map[string]any)
		require.True(t, ok)
		require.NotEmpty(t, replacedFiles)
		assert.Equal(t, true, replacedFiles[0]["success"])

		updated, updatedErr := sandbox.FileSystem.DownloadFile(ctx, helloPath, nil)
		require.NoError(t, updatedErr)
		assert.Equal(t, "world world", string(updated))
	})

	t.Run("FileSystem/MoveFiles", func(t *testing.T) {
		require.NoError(t, sandbox.FileSystem.MoveFiles(ctx, helloPath, movedPath))
		files, listErr := sandbox.FileSystem.ListFiles(ctx, textDir)
		require.NoError(t, listErr)
		assert.True(t, containsFileName(files, "moved.txt"))
	})

	t.Run("FileSystem/DeleteFile", func(t *testing.T) {
		require.NoError(t, sandbox.FileSystem.DeleteFile(ctx, movedPath, false))
		files, listErr := sandbox.FileSystem.ListFiles(ctx, textDir)
		require.NoError(t, listErr)
		assert.False(t, containsFileName(files, "moved.txt"))
	})

	t.Run("FileSystem/DeleteNestedDirectory", func(t *testing.T) {
		require.NoError(t, sandbox.FileSystem.DeleteFile(ctx, textDir+"/nested", true))
		files, listErr := sandbox.FileSystem.ListFiles(ctx, textDir)
		require.NoError(t, listErr)
		assert.False(t, containsFileName(files, "nested"))
	})

	t.Run("Process/ExecuteBasicEcho", func(t *testing.T) {
		result, execErr := sandbox.Process.ExecuteCommand(ctx, "echo hello")
		require.NoError(t, execErr)
		assert.Equal(t, 0, result.ExitCode)
		assert.Contains(t, result.Result, "hello")
	})

	t.Run("Process/ExecuteWithCwd", func(t *testing.T) {
		cwdResult, cwdErr := sandbox.Process.ExecuteCommand(ctx, "pwd", options.WithCwd(workDir))
		require.NoError(t, cwdErr)
		assert.Equal(t, 0, cwdResult.ExitCode)
		assert.Contains(t, strings.TrimSpace(cwdResult.Result), workDir)
	})

	t.Run("Process/ExecuteWithEnv", func(t *testing.T) {
		envResult, envErr := sandbox.Process.ExecuteCommand(
			ctx,
			"echo env-option",
			options.WithCommandEnv(map[string]string{"E2E_ENV": "enabled"}),
		)
		require.NoError(t, envErr)
		assert.Equal(t, 0, envResult.ExitCode)
		assert.Contains(t, envResult.Result, "env-option")
	})

	t.Run("Process/ExecuteNonzeroExitCode", func(t *testing.T) {
		failingResult, failErr := sandbox.Process.ExecuteCommand(ctx, "exit 1")
		require.NoError(t, failErr)
		assert.NotEqual(t, 0, failingResult.ExitCode)
	})

	t.Run("Process/CodeRunPython", func(t *testing.T) {
		pythonResult, pyErr := sandbox.Process.ExecuteCommand(ctx, `python -c "print('hello from python')"`)
		require.NoError(t, pyErr)
		assert.Equal(t, 0, pythonResult.ExitCode)
		assert.Contains(t, pythonResult.Result, "hello from python")
	})

	t.Run("Process/CodeRunMultiline", func(t *testing.T) {
		pythonResult, pyErr := sandbox.Process.ExecuteCommand(ctx, `python - <<'PY'
total = 0
for i in range(5):
    total += i
print(total)
PY`)
		require.NoError(t, pyErr)
		assert.Equal(t, 0, pythonResult.ExitCode)
		assert.Contains(t, pythonResult.Result, "10")
	})

	t.Run("Process/CodeRunSyntaxError", func(t *testing.T) {
		pythonResult, pyErr := sandbox.Process.ExecuteCommand(ctx, `python -c "if True print('x')"`)
		require.NoError(t, pyErr)
		assert.NotEqual(t, 0, pythonResult.ExitCode)
	})

	t.Run("Process/CodeRunUnsupportedMethod", func(t *testing.T) {
		codeRunResult, codeRunErr := sandbox.Process.CodeRun(ctx, "print('unsupported')")
		require.NoError(t, codeRunErr)
		require.NotNil(t, codeRunResult)
		assert.Equal(t, 0, codeRunResult.ExitCode)
		assert.Contains(t, codeRunResult.Result, "unsupported")
	})

	t.Run("Process/GetEntrypointSession", func(t *testing.T) {
		entrypoint, entryErr := sandbox.Process.GetEntrypointSession(ctx)
		require.NoError(t, entryErr)
		require.NotNil(t, entrypoint)
		assert.NotEmpty(t, entrypoint.GetSessionId())
	})

	t.Run("Process/GetEntrypointLogs", func(t *testing.T) {
		logs, logsErr := sandbox.Process.GetEntrypointLogs(ctx)
		require.NoError(t, logsErr)
		assert.NotNil(t, logs)
	})

	t.Run("Sessions/CreateSession", func(t *testing.T) {
		require.NoError(t, sandbox.Process.CreateSession(ctx, sessionID))
	})

	t.Run("Sessions/GetSession", func(t *testing.T) {
		session, sessionErr := sandbox.Process.GetSession(ctx, sessionID)
		require.NoError(t, sessionErr)
		assert.Equal(t, sessionID, session["sessionId"])
	})

	t.Run("Sessions/ExecuteSessionCommand", func(t *testing.T) {
		result, execErr := sandbox.Process.ExecuteSessionCommand(ctx, sessionID, "echo session-ready", false, false)
		require.NoError(t, execErr)
		stdout, ok := result["stdout"].(string)
		require.True(t, ok)
		assert.Contains(t, stdout, "session-ready")
	})

	t.Run("Sessions/SessionStatePersistence", func(t *testing.T) {
		_, exportErr := sandbox.Process.ExecuteSessionCommand(ctx, sessionID, "export E2E_VAR=persisted", false, false)
		require.NoError(t, exportErr)

		echoResult, echoErr := sandbox.Process.ExecuteSessionCommand(ctx, sessionID, "echo $E2E_VAR", false, false)
		require.NoError(t, echoErr)
		stdout, ok := echoResult["stdout"].(string)
		require.True(t, ok)
		assert.Contains(t, strings.TrimSpace(stdout), "persisted")
	})

	t.Run("Sessions/ExecuteAsyncAndGetCommand", func(t *testing.T) {
		asyncResult, asyncErr := sandbox.Process.ExecuteSessionCommand(ctx, sessionID, "echo async-log-line", true, false)
		require.NoError(t, asyncErr)

		idVal, ok := asyncResult["id"].(string)
		require.True(t, ok)
		require.NotEmpty(t, idVal)
		sessionCmdID = idVal

		var status map[string]any
		for i := 0; i < 10; i++ {
			status, asyncErr = sandbox.Process.GetSessionCommand(ctx, sessionID, sessionCmdID)
			require.NoError(t, asyncErr)
			if _, done := status["exitCode"]; done {
				break
			}
			time.Sleep(300 * time.Millisecond)
		}

		assert.Equal(t, sessionCmdID, status["id"])
		assert.Contains(t, fmt.Sprint(status["command"]), "echo async-log-line")
	})

	t.Run("Sessions/GetSessionCommandLogs", func(t *testing.T) {
		require.NotEmpty(t, sessionCmdID)
		logs, logsErr := sandbox.Process.GetSessionCommandLogs(ctx, sessionID, sessionCmdID)
		require.NoError(t, logsErr)
		assert.Contains(t, logs.GetOutput(), "async-log-line")
	})

	t.Run("Sessions/ListSessions", func(t *testing.T) {
		sessions, listErr := sandbox.Process.ListSessions(ctx)
		require.NoError(t, listErr)
		assert.True(t, containsSessionID(sessions, sessionID))
	})

	t.Run("Sessions/DeleteSession", func(t *testing.T) {
		require.NoError(t, sandbox.Process.DeleteSession(ctx, sessionID))
		sessions, listErr := sandbox.Process.ListSessions(ctx)
		require.NoError(t, listErr)
		assert.False(t, containsSessionID(sessions, sessionID))
	})

	t.Run("Git/ClonePublicRepo", func(t *testing.T) {
		require.NoError(t, sandbox.Git.Clone(ctx, "https://github.com/octocat/Hello-World.git", repoPath))
	})

	t.Run("Git/Status", func(t *testing.T) {
		status, statusErr := sandbox.Git.Status(ctx, repoPath)
		require.NoError(t, statusErr)
		assert.NotEmpty(t, status.CurrentBranch)
	})

	t.Run("Git/Branches", func(t *testing.T) {
		branches, branchesErr := sandbox.Git.Branches(ctx, repoPath)
		require.NoError(t, branchesErr)
		assert.NotEmpty(t, branches)
	})

	t.Run("Git/CreateBranch", func(t *testing.T) {
		require.NoError(t, sandbox.Git.CreateBranch(ctx, repoPath, "sdk-go-e2e-branch"))
		branches, branchesErr := sandbox.Git.Branches(ctx, repoPath)
		require.NoError(t, branchesErr)
		assert.Contains(t, branches, "sdk-go-e2e-branch")
	})

	t.Run("Git/CheckoutBranch", func(t *testing.T) {
		require.NoError(t, sandbox.Git.Checkout(ctx, repoPath, "sdk-go-e2e-branch"))
		status, statusErr := sandbox.Git.Status(ctx, repoPath)
		require.NoError(t, statusErr)
		assert.Equal(t, "sdk-go-e2e-branch", status.CurrentBranch)
	})

	t.Run("Git/AddFiles", func(t *testing.T) {
		require.NoError(t, sandbox.FileSystem.UploadFile(ctx, []byte("new content\n"), repoPath+"/e2e-added.txt"))
		require.NoError(t, sandbox.Git.Add(ctx, repoPath, []string{"e2e-added.txt"}))
		status, statusErr := sandbox.Git.Status(ctx, repoPath)
		require.NoError(t, statusErr)
		assert.NotEmpty(t, status.FileStatus)
	})

	t.Run("Git/Commit", func(t *testing.T) {
		commitResp, commitErr := sandbox.Git.Commit(ctx, repoPath, "Add e2e-added file", "SDK E2E", "sdk-e2e@example.com")
		require.NoError(t, commitErr)
		require.NotNil(t, commitResp)
		assert.NotEmpty(t, commitResp.SHA)
	})

	t.Run("Git/DeleteBranch", func(t *testing.T) {
		_ = sandbox.Git.Checkout(ctx, repoPath, "master")
		if status, statusErr := sandbox.Git.Status(ctx, repoPath); statusErr == nil && status.CurrentBranch != "master" {
			_ = sandbox.Git.Checkout(ctx, repoPath, "main")
		}

		require.NoError(t, sandbox.Git.DeleteBranch(ctx, repoPath, "sdk-go-e2e-branch"))
		branches, branchesErr := sandbox.Git.Branches(ctx, repoPath)
		require.NoError(t, branchesErr)
		assert.NotContains(t, branches, "sdk-go-e2e-branch")
	})

	t.Run("Git/CloneSpecificBranch", func(t *testing.T) {
		err := sandbox.Git.Clone(ctx, "https://github.com/octocat/Hello-World.git", cloneBranchPath, options.WithBranch("master"))
		if err != nil {
			require.NoError(t, sandbox.Git.Clone(ctx, "https://github.com/octocat/Hello-World.git", cloneBranchPath, options.WithBranch("main")))
		} else {
			require.NoError(t, err)
		}

		status, statusErr := sandbox.Git.Status(ctx, cloneBranchPath)
		require.NoError(t, statusErr)
		assert.NotEmpty(t, status.CurrentBranch)
	})

	t.Run("CodeInterpreter/RunCodeSimple", func(t *testing.T) {
		channels, runErr := sandbox.CodeInterpreter.RunCode(ctx, "print('simple-run')")
		require.NoError(t, runErr)

		result := <-channels.Done
		require.NotNil(t, result)
		assert.Nil(t, result.Error)
		assert.Contains(t, result.Stdout, "simple-run")
	})

	t.Run("CodeInterpreter/CreateContext", func(t *testing.T) {
		ctxInfo, ctxErr := sandbox.CodeInterpreter.CreateContext(ctx, nil)
		require.NoError(t, ctxErr)

		id, ok := ctxInfo["id"].(string)
		require.True(t, ok)
		require.NotEmpty(t, id)
		createdContexts = append(createdContexts, id)
	})

	t.Run("CodeInterpreter/RunCodeStatePersistence", func(t *testing.T) {
		require.NotEmpty(t, createdContexts)
		contextID := createdContexts[0]

		channels, runErr := sandbox.CodeInterpreter.RunCode(ctx, "x = 41", options.WithCustomContext(contextID))
		require.NoError(t, runErr)
		first := <-channels.Done
		require.NotNil(t, first)
		assert.Nil(t, first.Error)

		channels, runErr = sandbox.CodeInterpreter.RunCode(ctx, "print(x + 1)", options.WithCustomContext(contextID))
		require.NoError(t, runErr)
		second := <-channels.Done
		require.NotNil(t, second)
		assert.Nil(t, second.Error)
		assert.Contains(t, second.Stdout, "42")
	})

	t.Run("CodeInterpreter/RunCodeInIsolatedContext", func(t *testing.T) {
		ctxInfo, ctxErr := sandbox.CodeInterpreter.CreateContext(ctx, nil)
		require.NoError(t, ctxErr)
		contextID, ok := ctxInfo["id"].(string)
		require.True(t, ok)
		createdContexts = append(createdContexts, contextID)

		channels, runErr := sandbox.CodeInterpreter.RunCode(ctx, "print('isolated')", options.WithCustomContext(contextID))
		require.NoError(t, runErr)
		result := <-channels.Done
		require.NotNil(t, result)
		assert.Nil(t, result.Error)
		assert.Contains(t, result.Stdout, "isolated")
	})

	t.Run("CodeInterpreter/ListContexts", func(t *testing.T) {
		contexts, listErr := sandbox.CodeInterpreter.ListContexts(ctx)
		require.NoError(t, listErr)
		contextIDs := extractContextIDs(contexts)
		for _, id := range createdContexts {
			assert.Contains(t, contextIDs, id)
		}
	})

	t.Run("CodeInterpreter/DeleteContext", func(t *testing.T) {
		require.NotEmpty(t, createdContexts)
		last := createdContexts[len(createdContexts)-1]
		require.NoError(t, sandbox.CodeInterpreter.DeleteContext(ctx, last))
		createdContexts = createdContexts[:len(createdContexts)-1]
	})

	t.Run("LSP/StartPythonServer", func(t *testing.T) {
		require.NoError(t, sandbox.FileSystem.CreateFolder(ctx, lspProjectDir))
		require.NoError(t, sandbox.FileSystem.UploadFile(ctx, []byte("class Greeter:\n    def greet(self) -> str:\n        return 'hello'\n\ngreeter = Greeter()\ngreeter.\n"), lspFilePath))

		lspServer = NewLspServerService(sandbox.ToolboxClient, types.LspLanguagePython, lspProjectDir, sandbox.otel)
		require.NoError(t, lspServer.Start(ctx))
	})

	t.Run("LSP/DidOpen", func(t *testing.T) {
		require.NotNil(t, lspServer)
		require.NoError(t, lspServer.DidOpen(ctx, lspFilePath))
		time.Sleep(5 * time.Second)
	})

	t.Run("LSP/DocumentSymbols", func(t *testing.T) {
		require.NotNil(t, lspServer)
		symbols, symbolsErr := lspServer.DocumentSymbols(ctx, lspFilePath)
		require.NoError(t, symbolsErr)
		assert.NotEmpty(t, symbols)
		assert.Contains(t, fmt.Sprint(symbols), "Greeter")
	})

	t.Run("LSP/SandboxSymbols", func(t *testing.T) {
		require.NotNil(t, lspServer)
		symbols, symbolsErr := lspServer.SandboxSymbols(ctx, "Greeter")
		require.NoError(t, symbolsErr)
		assert.NotEmpty(t, symbols)
	})

	t.Run("LSP/DidClose", func(t *testing.T) {
		require.NotNil(t, lspServer)
		require.NoError(t, lspServer.DidClose(ctx, lspFilePath))
	})

	t.Run("LSP/Stop", func(t *testing.T) {
		require.NotNil(t, lspServer)
		require.NoError(t, lspServer.Stop(ctx))
		lspServer = nil
	})

	t.Run("PTY/CreateSessionAndList", func(t *testing.T) {
		handle, createErr := sandbox.Process.CreatePty(ctx, ptySessionID, options.WithCreatePtySize(types.PtySize{Rows: 24, Cols: 80}))
		require.NoError(t, createErr)

		require.NoError(t, handle.Disconnect())
		sessions, listErr := sandbox.Process.ListPtySessions(ctx)
		require.NoError(t, listErr)
		assert.True(t, containsPtySessionID(sessions, ptySessionID))
	})

	t.Run("PTY/GetSessionInfo", func(t *testing.T) {
		info, infoErr := sandbox.Process.GetPtySessionInfo(ctx, ptySessionID)
		require.NoError(t, infoErr)
		assert.Equal(t, ptySessionID, info.ID)
	})

	t.Run("PTY/ResizeSession", func(t *testing.T) {
		info, infoErr := sandbox.Process.ResizePtySession(ctx, ptySessionID, types.PtySize{Rows: 30, Cols: 100})
		require.NoError(t, infoErr)
		assert.Equal(t, 30, info.Rows)
		assert.Equal(t, 100, info.Cols)
	})

	t.Run("PTY/ConnectWriteReadAndClose", func(t *testing.T) {
		handle, connectErr := sandbox.Process.ConnectPty(ctx, ptySessionID)
		require.NoError(t, connectErr)
		defer func() { _ = handle.Disconnect() }()

		require.NoError(t, handle.WaitForConnection(ctx))

		var output strings.Builder
		done := make(chan struct{})
		go func() {
			defer close(done)
			for data := range handle.DataChan() {
				output.Write(data)
			}
		}()

		require.NoError(t, handle.SendInput([]byte("printf 'pty-output\\n'\n")))
		time.Sleep(2 * time.Second)
		require.NoError(t, handle.SendInput([]byte("exit\n")))

		waitCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()

		result, waitErr := handle.Wait(waitCtx)
		require.NoError(t, waitErr)
		if result.ExitCode != nil {
			assert.Equal(t, 0, *result.ExitCode)
		}

		<-done
		assert.Contains(t, output.String(), "pty-output")
	})

	t.Run("Process/ExecuteNonexistentPath", func(t *testing.T) {
		result, execErr := sandbox.Process.ExecuteCommand(ctx, "ls /definitely-missing-e2e-path")
		require.NoError(t, execErr)
		assert.NotEqual(t, 0, result.ExitCode)
	})

	t.Run("FileSystem/DownloadNonexistentFile", func(t *testing.T) {
		_, downloadErr := sandbox.FileSystem.DownloadFile(ctx, textDir+"/does-not-exist.txt", nil)
		require.Error(t, downloadErr)
	})

	t.Run("Sessions/DuplicateSessionGraceful", func(t *testing.T) {
		duplicateSessionID := "duplicate-session-" + suffix
		require.NoError(t, sandbox.Process.CreateSession(ctx, duplicateSessionID))

		err := sandbox.Process.CreateSession(ctx, duplicateSessionID)
		if err != nil {
			assert.NotEmpty(t, err.Error())
		} else {
			session, sessionErr := sandbox.Process.GetSession(ctx, duplicateSessionID)
			require.NoError(t, sessionErr)
			assert.Equal(t, duplicateSessionID, session["sessionId"])
		}

		_ = sandbox.Process.DeleteSession(ctx, duplicateSessionID)
	})

	t.Run("Process/ExecuteWithTimeout", func(t *testing.T) {
		result, execErr := sandbox.Process.ExecuteCommand(ctx, "sleep 2", options.WithExecuteTimeout(1*time.Second))
		if execErr != nil {
			assert.Contains(t, strings.ToLower(execErr.Error()), "timeout")
			return
		}
		assert.NotEqual(t, 0, result.ExitCode)
	})

	t.Run("Process/CodeRunStructuredData", func(t *testing.T) {
		result, runErr := sandbox.Process.CodeRun(ctx, `import json
print(json.dumps({"items": [1, 2, 3], "meta": {"ok": True}}))`)
		require.NoError(t, runErr)
		require.Equal(t, 0, result.ExitCode)

		var payload map[string]any
		require.NoError(t, json.Unmarshal([]byte(strings.TrimSpace(result.Result)), &payload))
		assert.Equal(t, []any{float64(1), float64(2), float64(3)}, payload["items"])
		assert.Equal(t, map[string]any{"ok": true}, payload["meta"])
	})

	t.Run("Process/LongRunningCommand", func(t *testing.T) {
		result, execErr := sandbox.Process.ExecuteCommand(ctx, `python - <<'PY'
import time
time.sleep(1)
print("long-run-complete")
PY`, options.WithExecuteTimeout(10*time.Second))
		require.NoError(t, execErr)
		assert.Equal(t, 0, result.ExitCode)
		assert.Contains(t, result.Result, "long-run-complete")
	})

	t.Run("SandboxLifecycle/ArchiveAndUnarchive", func(t *testing.T) {
		require.NoError(t, sandbox.StopWithTimeout(ctx, 90*time.Second, false))

		require.NoError(t, sandbox.Archive(ctx))

		require.NoError(t, sandbox.RefreshData(ctx))
		assert.Contains(t, []apiclient.SandboxState{apiclient.SANDBOXSTATE_ARCHIVED, apiclient.SANDBOXSTATE_ARCHIVING, apiclient.SANDBOXSTATE_STOPPED}, sandbox.State)
		require.NoError(t, sandbox.StartWithTimeout(ctx, 2*time.Minute))
		require.NoError(t, sandbox.RefreshData(ctx))
		assert.Equal(t, apiclient.SANDBOXSTATE_STARTED, sandbox.State)
	})

	t.Run("SandboxLifecycle/PostArchiveUsable", func(t *testing.T) {
		result, execErr := sandbox.Process.ExecuteCommand(ctx, "echo post-archive-check")
		require.NoError(t, execErr)
		assert.Equal(t, 0, result.ExitCode)
		assert.Contains(t, result.Result, "post-archive-check")
	})

	t.Run("Volume/Create", func(t *testing.T) {
		vol, volErr := client.Volume.Create(ctx, volumeName)
		require.NoError(t, volErr)
		require.NotNil(t, vol)
		createdVolume = vol
		assert.Equal(t, volumeName, vol.Name)
	})

	t.Run("Volume/List", func(t *testing.T) {
		vols, listErr := client.Volume.List(ctx)
		require.NoError(t, listErr)

		names := make([]string, 0, len(vols))
		for _, v := range vols {
			names = append(names, v.Name)
		}
		assert.Contains(t, names, volumeName)
	})

	t.Run("Volume/Get", func(t *testing.T) {
		vol, getErr := client.Volume.Get(ctx, volumeName)
		require.NoError(t, getErr)
		require.NotNil(t, vol)
		assert.Equal(t, volumeName, vol.Name)
	})

	t.Run("Volume/Delete", func(t *testing.T) {
		require.NotNil(t, createdVolume)
		for i := 0; i < 15; i++ {
			vol, err := client.Volume.Get(ctx, volumeName)
			if err == nil && (vol.State == "ready" || vol.State == "error") {
				break
			}
			time.Sleep(1 * time.Second)
		}
		require.NoError(t, client.Volume.Delete(ctx, createdVolume))
		createdVolume = nil
	})

	t.Run("Snapshot/List", func(t *testing.T) {
		snapshots, listErr := client.Snapshot.List(ctx, nil, nil)
		require.NoError(t, listErr)
		require.NotNil(t, snapshots)

		if len(snapshots.Items) > 0 {
			snapshotName = snapshots.Items[0].Name
		}
	})

	t.Run("Snapshot/ListWithPagination", func(t *testing.T) {
		page, limit := 1, 5
		snapshots, listErr := client.Snapshot.List(ctx, &page, &limit)
		require.NoError(t, listErr)
		require.NotNil(t, snapshots)
		assert.Equal(t, 1, snapshots.Page)
	})

	t.Run("Snapshot/GetByName", func(t *testing.T) {
		require.NotEmpty(t, snapshotName, "snapshot list should have returned at least the default snapshot")

		snapshot, getErr := client.Snapshot.Get(ctx, snapshotName)
		require.NoError(t, getErr)
		require.NotNil(t, snapshot)
		assert.Equal(t, snapshotName, snapshot.Name)
	})

	t.Run("ClientOps/ListSandboxes", func(t *testing.T) {
		listed, listErr := client.List(ctx, nil, nil, nil)
		require.NoError(t, listErr)
		assert.GreaterOrEqual(t, listed.Total, 1)
	})

	t.Run("ClientOps/ListWithPagination", func(t *testing.T) {
		page, limit := 1, 10
		listed, listErr := client.List(ctx, nil, &page, &limit)
		require.NoError(t, listErr)
		assert.Equal(t, 1, listed.Page)
		assert.GreaterOrEqual(t, listed.Total, 1)
	})

	t.Run("ClientOps/GetByID", func(t *testing.T) {
		got, getErr := client.Get(ctx, sandbox.ID)
		require.NoError(t, getErr)
		require.NotNil(t, got)
		assert.Equal(t, sandbox.ID, got.ID)
	})

	t.Run("ClientOps/GetByName", func(t *testing.T) {
		got, getErr := client.Get(ctx, sandbox.Name)
		require.NoError(t, getErr)
		require.NotNil(t, got)
		assert.Equal(t, sandbox.Name, got.Name)
	})

	for _, contextID := range createdContexts {
		_ = sandbox.CodeInterpreter.DeleteContext(ctx, contextID)
	}
	if lspServer != nil {
		_ = lspServer.Stop(ctx)
	}
	if createdVolume != nil {
		_ = client.Volume.Delete(ctx, createdVolume)
	}
	_ = sandbox.FileSystem.DeleteFile(ctx, baseDir, true)
	_ = sandbox.Process.DeleteSession(ctx, sessionID)
	_ = sandbox.StartWithTimeout(ctx, 60*time.Second)
}

func extractContextIDs(contexts []map[string]any) []string {
	ids := make([]string, 0, len(contexts))
	for _, c := range contexts {
		if id, ok := c["id"].(string); ok {
			ids = append(ids, id)
		}
	}
	return ids
}

func containsFileName(files []*types.FileInfo, fileName string) bool {
	for _, f := range files {
		if f.Name == fileName {
			return true
		}
	}
	return false
}

func containsSessionID(sessions []map[string]any, sessionID string) bool {
	for _, session := range sessions {
		if id, ok := session["sessionId"].(string); ok && id == sessionID {
			return true
		}
	}
	return false
}

func containsPtySessionID(sessions []*types.PtySessionInfo, sessionID string) bool {
	for _, session := range sessions {
		if session != nil && session.ID == sessionID {
			return true
		}
	}
	return false
}

func containsAny(message string, parts ...string) bool {
	for _, part := range parts {
		if strings.Contains(message, part) {
			return true
		}
	}
	return false
}
