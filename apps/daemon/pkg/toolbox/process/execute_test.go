// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package process

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
	"github.com/gin-gonic/gin"
)

func execute(t *testing.T, body string) (int, ExecuteResponse) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(common_errors.NewErrorMiddleware(func(_ *gin.Context, err error) common_errors.ErrorResponse {
		return common_errors.ErrorResponse{StatusCode: http.StatusInternalServerError, Message: err.Error()}
	}))
	router.POST("/process/execute", ExecuteCommand(slog.New(slog.NewTextHandler(io.Discard, nil))))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/process/execute", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)

	var response ExecuteResponse
	if recorder.Code == http.StatusOK {
		if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
			t.Fatalf("decode response: %v", err)
		}
	}
	return recorder.Code, response
}

func executeRequest(t *testing.T, request ExecuteRequest) (int, ExecuteResponse) {
	t.Helper()
	body, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("encode request: %v", err)
	}
	return execute(t, string(body))
}

func TestExecuteCommandSeparatesOutput(t *testing.T) {
	status, response := execute(t, `{"command":"printf out; printf err >&2; exit 7"}`)

	if status != http.StatusOK {
		t.Fatalf("status = %d, want %d", status, http.StatusOK)
	}
	if response.ExitCode != 7 {
		t.Fatalf("exit code = %d, want 7", response.ExitCode)
	}
	if response.Stdout != "out" || response.Stderr != "err" {
		t.Fatalf("streams = (%q, %q), want (%q, %q)", response.Stdout, response.Stderr, "out", "err")
	}
	if !strings.Contains(response.Result, "out") || !strings.Contains(response.Result, "err") {
		t.Fatalf("combined result %q does not contain both streams", response.Result)
	}
	if response.TimedOut || response.StdoutTruncated || response.StderrTruncated {
		t.Fatalf("unexpected response flags: %+v", response)
	}
}

func TestExecuteCommandPreservesInterleavedStreamIdentity(t *testing.T) {
	status, response := execute(t, `{"command":"printf out1; printf err1 >&2; printf out2; printf err2 >&2"}`)

	if status != http.StatusOK {
		t.Fatalf("status = %d, want %d", status, http.StatusOK)
	}
	if response.Stdout != "out1out2" || response.Stderr != "err1err2" {
		t.Fatalf("streams = (%q, %q), want (%q, %q)", response.Stdout, response.Stderr, "out1out2", "err1err2")
	}
}

func TestExecuteCommandBoundsStreamsIndependently(t *testing.T) {
	status, response := execute(t, `{"command":"printf 1234567890; printf abcdefghij >&2","maxOutputBytes":4}`)

	if status != http.StatusOK {
		t.Fatalf("status = %d, want %d", status, http.StatusOK)
	}
	if response.Stdout != "7890" || response.Stderr != "ghij" {
		t.Fatalf("streams = (%q, %q), want (%q, %q)", response.Stdout, response.Stderr, "7890", "ghij")
	}
	if !response.StdoutTruncated || !response.StderrTruncated {
		t.Fatalf("truncation flags = (%t, %t), want both true", response.StdoutTruncated, response.StderrTruncated)
	}
	if len(response.Result) > 4 {
		t.Fatalf("combined result has %d bytes, want at most 4", len(response.Result))
	}
}

func TestExecuteCommandDrainsLargeStreams(t *testing.T) {
	status, response := execute(t, `{"command":"head -c 1048576 /dev/zero | tr '\\0' a; head -c 1048576 /dev/zero | tr '\\0' b >&2","maxOutputBytes":1024}`)

	if status != http.StatusOK {
		t.Fatalf("status = %d, want %d", status, http.StatusOK)
	}
	if len(response.Stdout) != 1024 || !bytes.Equal([]byte(response.Stdout), bytes.Repeat([]byte{'a'}, 1024)) {
		t.Fatalf("stdout is not the expected 1024-byte tail")
	}
	if len(response.Stderr) != 1024 || !bytes.Equal([]byte(response.Stderr), bytes.Repeat([]byte{'b'}, 1024)) {
		t.Fatalf("stderr is not the expected 1024-byte tail")
	}
}

func TestExecuteCommandReturnsValidUTF8WithinLimit(t *testing.T) {
	status, response := execute(t, `{"command":"printf '\\303\\251x'","maxOutputBytes":2}`)

	if status != http.StatusOK {
		t.Fatalf("status = %d, want %d", status, http.StatusOK)
	}
	if response.Stdout != "x" || !response.StdoutTruncated {
		t.Fatalf("stdout = %q, truncated = %t, want valid tail %q", response.Stdout, response.StdoutTruncated, "x")
	}
}

func TestExecuteCommandPreservesUnboundedInvalidUTF8(t *testing.T) {
	status, response := execute(t, `{"command":"printf '\\377\\376'"}`)

	if status != http.StatusOK {
		t.Fatalf("status = %d, want %d", status, http.StatusOK)
	}
	if response.Result != "\uFFFD\uFFFD" {
		t.Fatalf("result = %q, want two replacement characters", response.Result)
	}
}

func TestExecuteCommandReturnsPartialOutputOnTimeout(t *testing.T) {
	status, response := execute(t, `{"command":"printf partial; sleep 5","timeout":1,"maxOutputBytes":100}`)

	if status != http.StatusOK {
		t.Fatalf("status = %d, want %d", status, http.StatusOK)
	}
	if !response.TimedOut || response.ExitCode != -1 {
		t.Fatalf("timeout result = %+v, want timed out with exit code -1", response)
	}
	if response.Stdout != "partial" || response.Stderr != "" || response.Result != "partial" {
		t.Fatalf("timeout streams = (%q, %q, %q)", response.Result, response.Stdout, response.Stderr)
	}
}

func TestExecuteCommandKillsProcessGroupOnTimeout(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "survived")
	timeout, limit := uint32(1), uint32(100)
	status, response := executeRequest(t, ExecuteRequest{
		Command:        fmt.Sprintf("(sleep 2; touch %q) & wait", marker),
		Timeout:        &timeout,
		MaxOutputBytes: &limit,
	})
	if status != http.StatusOK || !response.TimedOut {
		t.Fatalf("timeout result = (%d, %+v)", status, response)
	}

	time.Sleep(1500 * time.Millisecond)
	if _, err := os.Stat(marker); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("background process survived process-group timeout: %v", err)
	}
}

func TestExecuteCommandPreservesTimeoutErrorWithoutOutputLimit(t *testing.T) {
	status, _ := execute(t, `{"command":"sleep 5","timeout":1}`)
	if status != http.StatusRequestTimeout {
		t.Fatalf("status = %d, want %d", status, http.StatusRequestTimeout)
	}
}

func TestExecuteCommandRejectsZeroOutputLimit(t *testing.T) {
	status, _ := execute(t, `{"command":"echo no","maxOutputBytes":0}`)
	if status != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", status, http.StatusBadRequest)
	}
}

func TestExecuteCommandRejectsOutputLimitAboveClientRange(t *testing.T) {
	status, _ := execute(t, `{"command":"echo no","maxOutputBytes":2147483648}`)
	if status != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", status, http.StatusBadRequest)
	}
}
