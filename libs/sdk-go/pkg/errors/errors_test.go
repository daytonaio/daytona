// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package errors_test

import (
	stderrors "errors"
	"fmt"
	"net/http"
	"testing"

	sdkerrors "github.com/daytonaio/daytona/libs/sdk-go/pkg/errors"
)

// ----- Base error -----

func TestDaytonaError_Error(t *testing.T) {
	err := sdkerrors.NewDaytonaError("boom", http.StatusBadGateway, nil)

	want := "Daytona error (status 502): boom"
	if err.Error() != want {
		t.Fatalf("Error() = %q, want %q", err.Error(), want)
	}

	plain := sdkerrors.NewDaytonaError("offline", 0, nil)
	if plain.Error() != "Daytona error: offline" {
		t.Fatalf("plain Error() = %q", plain.Error())
	}
}

func TestNewDaytonaError_HoldsAllFields(t *testing.T) {
	headers := http.Header{"X-Test": []string{"v"}}
	err := sdkerrors.NewDaytonaError("msg", 418, headers)
	err.Code = "TEAPOT"
	err.Source = sdkerrors.SourceAPI

	var de *sdkerrors.DaytonaError
	if !stderrors.As(err, &de) {
		t.Fatalf("errors.As(*DaytonaError) failed")
	}
	if de.Message != "msg" || de.StatusCode != 418 || de.Code != "TEAPOT" || de.Source != sdkerrors.SourceAPI {
		t.Fatalf("fields: %+v", de)
	}
	if de.Headers.Get("X-Test") != "v" {
		t.Fatalf("headers not set: %v", de.Headers)
	}
}

func TestNewDaytonaTimeoutError_HasRequestTimeoutStatus(t *testing.T) {
	err := sdkerrors.NewDaytonaTimeoutError("slow")
	if err.StatusCode != http.StatusRequestTimeout {
		t.Fatalf("StatusCode = %d, want %d", err.StatusCode, http.StatusRequestTimeout)
	}
	if !stderrors.Is(err, sdkerrors.ErrTimeout) {
		t.Fatalf("errors.Is(err, ErrTimeout) = false")
	}
}

func TestNewDaytonaConnectionError_NoStatus(t *testing.T) {
	err := sdkerrors.NewDaytonaConnectionError("no route")
	if err.StatusCode != 0 {
		t.Fatalf("connection error should have StatusCode 0, got %d", err.StatusCode)
	}
	// Doesn't match any HTTP status sentinel.
	if stderrors.Is(err, sdkerrors.ErrNotFound) {
		t.Fatalf("connection error should not match ErrNotFound")
	}
}

// ----- errors.Is sentinel matching -----

func TestErrorsIs_StatusClassSentinels(t *testing.T) {
	err := sdkerrors.NewDaytonaError("missing", http.StatusNotFound, nil)
	if !stderrors.Is(err, sdkerrors.ErrNotFound) {
		t.Fatalf("errors.Is(err, ErrNotFound) = false")
	}
	if stderrors.Is(err, sdkerrors.ErrValidation) {
		t.Fatalf("errors.Is(err, ErrValidation) should be false for 404")
	}
}

func TestErrorsIs_DomainCodeAlsoMatchesParentStatus(t *testing.T) {
	body := []byte(`{"statusCode":401,"message":"creds rejected","code":"GIT_AUTH_FAILED","source":"DAYTONA_DAEMON"}`)
	err := sdkerrors.NewDaytonaErrorFromBody(body, http.StatusUnauthorized, nil)

	if !stderrors.Is(err, sdkerrors.ErrGitAuthFailed) {
		t.Fatalf("errors.Is(err, ErrGitAuthFailed) = false")
	}
	if !stderrors.Is(err, sdkerrors.ErrAuthentication) {
		t.Fatalf("errors.Is(err, ErrAuthentication) = false; want true (domain inherits from status)")
	}
	if stderrors.Is(err, sdkerrors.ErrNotFound) {
		t.Fatalf("errors.Is(err, ErrNotFound) = true; want false")
	}
	if stderrors.Is(err, sdkerrors.ErrGitRepoNotFound) {
		t.Fatalf("errors.Is(err, ErrGitRepoNotFound) = true; want false (different code)")
	}
}

func TestErrorsIs_DomainCodesRequireBothSourceAndCode(t *testing.T) {
	body := []byte(`{"statusCode":404,"code":"SANDBOX_NOT_FOUND","source":"DAYTONA_API"}`)
	err := sdkerrors.NewDaytonaErrorFromBody(body, http.StatusNotFound, nil)

	if stderrors.Is(err, sdkerrors.ErrSandboxNotFound) {
		t.Fatalf("errors.Is should require both source AND code to match")
	}
	if !stderrors.Is(err, sdkerrors.ErrNotFound) {
		t.Fatalf("errors.Is(err, ErrNotFound) = false")
	}
}

func TestErrorsIs_ApiDomainCodes(t *testing.T) {
	body := []byte(`{"statusCode":409,"code":"SANDBOX_STATE_CHANGE_IN_PROGRESS","source":"DAYTONA_API","message":"in flight"}`)
	err := sdkerrors.NewDaytonaErrorFromBody(body, http.StatusConflict, nil)

	if !stderrors.Is(err, sdkerrors.ErrSandboxStateChangeInProgress) {
		t.Fatalf("precise sentinel did not match")
	}
	if !stderrors.Is(err, sdkerrors.ErrConflict) {
		t.Fatalf("parent ErrConflict did not match")
	}
}

func TestErrorsIs_ProxySandboxNotFound(t *testing.T) {
	body := []byte(`{"statusCode":404,"code":"SANDBOX_NOT_FOUND","source":"DAYTONA_PROXY","message":"no sandbox"}`)
	err := sdkerrors.NewDaytonaErrorFromBody(body, http.StatusNotFound, nil)

	if !stderrors.Is(err, sdkerrors.ErrSandboxNotFound) {
		t.Fatalf("precise sentinel did not match")
	}
	if !stderrors.Is(err, sdkerrors.ErrNotFound) {
		t.Fatalf("parent ErrNotFound did not match")
	}
}

func TestIs_RejectsNonDaytonaErrorTargets(t *testing.T) {
	err := sdkerrors.NewDaytonaError("x", http.StatusNotFound, nil)
	if stderrors.Is(err, stderrors.New("plain")) {
		t.Fatalf("Is should not match arbitrary errors")
	}
}

// ----- NewDaytonaErrorFromBody -----

func TestNewDaytonaErrorFromBody_FillsAllFields(t *testing.T) {
	body := []byte(`{"statusCode":404,"message":"missing","code":"FILE_NOT_FOUND","source":"DAYTONA_DAEMON"}`)
	err := sdkerrors.NewDaytonaErrorFromBody(body, http.StatusNotFound, nil)

	if err.StatusCode != http.StatusNotFound {
		t.Fatalf("StatusCode = %d, want 404", err.StatusCode)
	}
	if err.Code != "FILE_NOT_FOUND" || err.Source != "DAYTONA_DAEMON" {
		t.Fatalf("code/source: %q / %q", err.Code, err.Source)
	}
	if err.Message != "missing" {
		t.Fatalf("Message = %q", err.Message)
	}

	if !stderrors.Is(err, sdkerrors.ErrFileNotFound) {
		t.Fatalf("errors.Is(err, ErrFileNotFound) = false")
	}
	if !stderrors.Is(err, sdkerrors.ErrNotFound) {
		t.Fatalf("errors.Is(err, ErrNotFound) = false")
	}
}

func TestNewDaytonaErrorFromBody_PrefersBodyStatusCode(t *testing.T) {
	body := []byte(`{"statusCode":410,"message":"gone"}`)
	err := sdkerrors.NewDaytonaErrorFromBody(body, http.StatusInternalServerError, nil)

	if err.StatusCode != http.StatusGone {
		t.Fatalf("StatusCode = %d, want 410", err.StatusCode)
	}
	if !stderrors.Is(err, sdkerrors.ErrGone) {
		t.Fatalf("errors.Is(err, ErrGone) = false")
	}
}

func TestNewDaytonaErrorFromBody_FallbackMessageForEmptyBody(t *testing.T) {
	err := sdkerrors.NewDaytonaErrorFromBody(nil, http.StatusNotFound, nil)
	if err.Message != "Request failed" {
		t.Fatalf("Message = %q, want fallback", err.Message)
	}
	if err.StatusCode != http.StatusNotFound {
		t.Fatalf("StatusCode = %d", err.StatusCode)
	}
}

func TestNewDaytonaErrorFromBody_PrefersErrorFieldWhenMessageAbsent(t *testing.T) {
	body := []byte(`{"error":"bad data"}`)
	err := sdkerrors.NewDaytonaErrorFromBody(body, http.StatusBadRequest, nil)
	if err.Message != "bad data" {
		t.Fatalf("Message = %q", err.Message)
	}
}

func TestNewDaytonaErrorFromBody_FallsBackToRawBodyForNonJson(t *testing.T) {
	body := []byte("not-json")
	err := sdkerrors.NewDaytonaErrorFromBody(body, http.StatusTeapot, nil)
	if err.Message != "not-json" {
		t.Fatalf("Message = %q", err.Message)
	}
}

// ----- Sentinels are zero-value DaytonaErrors with non-empty Error() output -----

func TestSentinels_HaveDescriptiveErrorString(t *testing.T) {
	// Sentinels should still implement error and produce a non-empty string,
	// so log statements like `log.Println(ErrNotFound)` don't surprise users.
	got := fmt.Sprintf("%v", sdkerrors.ErrNotFound)
	if got == "" {
		t.Fatalf("ErrNotFound has empty Error()")
	}
}
