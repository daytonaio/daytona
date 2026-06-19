// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package controllers

import (
	"errors"
	"io"
	"net/http"
	"testing"
)

// flakyBody yields its payload, then a stored error on every subsequent read.
type flakyBody struct {
	payload []byte
	err     error
	off     int
}

func (f *flakyBody) Read(p []byte) (int, error) {
	if f.off < len(f.payload) {
		n := copy(p, f.payload[f.off:])
		f.off += n
		return n, nil
	}
	return 0, f.err
}

func (f *flakyBody) Close() error { return nil }

func TestSniffFdExhaustionReplaysPartialReadError(t *testing.T) {
	readErr := errors.New("unexpected EOF mid-body")
	payload := []byte(`{"statusCode":500,"message":"truncat`)
	resp := &http.Response{
		StatusCode:    http.StatusInternalServerError,
		ContentLength: 128,
		Header:        http.Header{"Content-Type": []string{"application/json"}},
		Body:          &flakyBody{payload: payload, err: readErr},
	}

	sniffFdExhaustion(resp, "sandbox-id", "/process/execute")

	got, err := io.ReadAll(resp.Body)
	if string(got) != string(payload) {
		t.Fatalf("replayed body = %q, want consumed prefix %q", got, payload)
	}
	if !errors.Is(err, readErr) {
		t.Fatalf("replayed error = %v, want original read error %v", err, readErr)
	}
	if closeErr := resp.Body.Close(); closeErr != nil {
		t.Fatalf("close after replay: %v", closeErr)
	}
}
