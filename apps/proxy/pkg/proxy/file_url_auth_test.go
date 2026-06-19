// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package proxy

import (
	"testing"
)

// Cross-language test vectors — the Python SDK test suite asserts the exact
// same signatures (libs/sdk-python/tests/test_file_url_signing.py). If one of
// these changes, both sides must change together.
const crossLangTestKey = "testsigningkey1234567890abcdefgh"

var crossLangVectors = []struct {
	name     string
	method   string
	path     string
	expires  int64
	expected string
}{
	{
		name:     "download with expiry",
		method:   "GET",
		path:     "/home/user/report.pdf",
		expires:  1781234567,
		expected: "v1_QXoy36mypac2FAv33L7jnN44GEUx8KrdwT0vuKgKiQg",
	},
	{
		name:     "download no expiry",
		method:   "GET",
		path:     "/home/user/report.pdf",
		expires:  0,
		expected: "v1_lpj67Q-1iHxBviass5MZhGs36X80uk3DgCaRjjmyPrk",
	},
	{
		name:     "upload with expiry",
		method:   "POST",
		path:     "/tmp/incoming/data.bin",
		expires:  1781234567,
		expected: "v1_CziiRdFkC9asB7q1mi0-fDvvwkpTxcI7yR8N35ht9Vw",
	},
	{
		name:     "path with spaces",
		method:   "GET",
		path:     "/path with spaces/f.txt",
		expires:  1900000000,
		expected: "v1_GynMaKcifGfdmrBJHusa_ucAXowjZ_g4KP6lcMJ4WXE",
	},
}

func fileUrlSignature(signingKey, method, path string, expires int64) string {
	return computeSignature(signingKey, fileUrlCanonical(method, path, expires))
}

func TestFileUrlSignatureCrossLanguageVectors(t *testing.T) {
	for _, v := range crossLangVectors {
		t.Run(v.name, func(t *testing.T) {
			got := fileUrlSignature(crossLangTestKey, v.method, v.path, v.expires)
			if got != v.expected {
				t.Errorf("fileUrlSignature(%q, %q, %d) = %q, want %q", v.method, v.path, v.expires, got, v.expected)
			}
		})
	}
}

func TestFileUrlSignatureKeySensitivity(t *testing.T) {
	a := fileUrlSignature("key-a", "GET", "/f.txt", 0)
	b := fileUrlSignature("key-b", "GET", "/f.txt", 0)
	if a == b {
		t.Error("signatures with different keys must differ")
	}

	withExpiry := fileUrlSignature("key-a", "GET", "/f.txt", 12345)
	if a == withExpiry {
		t.Error("signatures with different expiry must differ")
	}

	post := fileUrlSignature("key-a", "POST", "/f.txt", 0)
	if a == post {
		t.Error("signatures with different methods must differ")
	}
}
