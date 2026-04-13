// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package session

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"
)

func newTestSessionService(t *testing.T) *SessionService {
	t.Helper()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewSessionService(logger, t.TempDir(), 250*time.Millisecond, 25*time.Millisecond)
}

func waitForCommandExit(t *testing.T, svc *SessionService, sessionID, commandID string, timeout time.Duration) *Command {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		command, err := svc.GetSessionCommand(sessionID, commandID)
		if err != nil {
			t.Fatalf("get session command: %v", err)
		}
		if command.ExitCode != nil {
			return command
		}
		time.Sleep(25 * time.Millisecond)
	}

	t.Fatalf("command %s did not exit within %s", commandID, timeout)
	return nil
}

func waitForInputPipe(t *testing.T, svc *SessionService, sessionID, commandID string, timeout time.Duration) {
	t.Helper()

	session, ok := svc.sessions.Get(sessionID)
	if !ok {
		t.Fatalf("session %s not found", sessionID)
	}

	command, ok := session.commands.Get(commandID)
	if !ok {
		t.Fatalf("command %s not found", commandID)
	}

	inputPath := command.InputFilePath(session.Dir(svc.configDir))
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(inputPath); err == nil {
			return
		}
		time.Sleep(25 * time.Millisecond)
	}

	t.Fatalf("input pipe %s was not created within %s", inputPath, timeout)
}

func TestExecuteSyncCommandGetsEOFOnStdin(t *testing.T) {
	svc := newTestSessionService(t)
	const sessionID = "sync-stdin"

	if err := svc.Create(sessionID, false); err != nil {
		t.Fatalf("create session: %v", err)
	}
	t.Cleanup(func() {
		_ = svc.Delete(context.Background(), sessionID)
	})

	resultCh := make(chan *SessionExecute, 1)
	errCh := make(chan error, 1)

	go func() {
		result, err := svc.Execute(
			context.Background(),
			sessionID,
			"",
			`if IFS= read -r line; then printf 'line:%s\n' "$line"; else printf 'eof\n'; fi`,
			false,
			true,
			true,
		)
		if err != nil {
			errCh <- err
			return
		}
		resultCh <- result
	}()

	select {
	case err := <-errCh:
		t.Fatalf("execute sync command: %v", err)
	case result := <-resultCh:
		if result.ExitCode == nil || *result.ExitCode != 0 {
			t.Fatalf("expected exit code 0, got %#v", result.ExitCode)
		}
		if result.Output == nil || !strings.Contains(*result.Output, "eof") {
			t.Fatalf("expected output to contain eof, got %#v", result.Output)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("sync session command hung waiting on stdin")
	}
}

func TestExecuteSyncHeredoc(t *testing.T) {
	svc := newTestSessionService(t)
	const sessionID = "heredoc-test"

	if err := svc.Create(sessionID, false); err != nil {
		t.Fatalf("create session: %v", err)
	}
	t.Cleanup(func() {
		_ = svc.Delete(context.Background(), sessionID)
	})

	resultCh := make(chan *SessionExecute, 1)
	errCh := make(chan error, 1)

	go func() {
		result, err := svc.Execute(
			context.Background(),
			sessionID,
			"",
			"cat <<'__EOF__'\nhello from heredoc\n__EOF__",
			false,
			true,
			true,
		)
		if err != nil {
			errCh <- err
			return
		}
		resultCh <- result
	}()

	select {
	case err := <-errCh:
		t.Fatalf("execute heredoc command: %v", err)
	case result := <-resultCh:
		if result.ExitCode == nil || *result.ExitCode != 0 {
			t.Fatalf("expected exit code 0, got %#v", result.ExitCode)
		}
		if result.Output == nil || !strings.Contains(*result.Output, "hello from heredoc") {
			t.Fatalf("expected heredoc output, got %#v", result.Output)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("heredoc command hung")
	}
}

func TestExecuteAsyncCommandStillAcceptsInput(t *testing.T) {
	svc := newTestSessionService(t)
	const sessionID = "async-stdin"

	if err := svc.Create(sessionID, false); err != nil {
		t.Fatalf("create session: %v", err)
	}
	t.Cleanup(func() {
		_ = svc.Delete(context.Background(), sessionID)
	})

	result, err := svc.Execute(
		context.Background(),
		sessionID,
		"",
		`if IFS= read -r line; then printf 'line:%s\n' "$line"; else printf 'eof\n'; fi`,
		true,
		true,
		true,
	)
	if err != nil {
		t.Fatalf("execute async command: %v", err)
	}

	waitForInputPipe(t, svc, sessionID, result.CommandId, 3*time.Second)

	if err := svc.SendInput(sessionID, result.CommandId, "hello from test"); err != nil {
		t.Fatalf("send input: %v", err)
	}

	command := waitForCommandExit(t, svc, sessionID, result.CommandId, 2*time.Second)
	if command.ExitCode == nil || *command.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %#v", command.ExitCode)
	}

	logs, err := svc.GetSessionCommandLogs(sessionID, result.CommandId, nil, nil, FetchLogsOptions{IsCombinedOutput: true})
	if err != nil {
		t.Fatalf("get session command logs: %v", err)
	}

	if !strings.Contains(string(logs), "line:hello from test") {
		t.Fatalf("expected async command logs to contain sent input, got %q", string(logs))
	}
}
