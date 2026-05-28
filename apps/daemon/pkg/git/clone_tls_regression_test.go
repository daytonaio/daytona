// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package git

import (
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"sync/atomic"
	"testing"

	"github.com/daytonaio/daemon/pkg/gitprovider"
	"github.com/stretchr/testify/require"
)

// newChallengingTLSServer returns an httptest TLS server that always responds
// with 401 + WWW-Authenticate: Basic. The challenge is required to provoke
// native git (CLI path) into sending credentials via GIT_ASKPASS — without it,
// native git never sends Authorization and a credential-leak regression would
// silently pass. The go-git path sends BasicAuth preemptively on the first
// request and is captured on the way in either way. The captured Authorization
// header is exposed via the returned atomic value.
func newChallengingTLSServer(t *testing.T) (*httptest.Server, *atomic.Value) {
	t.Helper()
	var receivedAuth atomic.Value
	receivedAuth.Store("")
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h := r.Header.Get("Authorization"); h != "" {
			receivedAuth.Store(h)
		}
		w.Header().Set("WWW-Authenticate", `Basic realm="git"`)
		w.WriteHeader(http.StatusUnauthorized)
	}))
	t.Cleanup(server.Close)
	return server, &receivedAuth
}

// Regression for GHSA-375h-72g4-hc9c on the go-git clone path: a credentialed
// clone against an HTTPS endpoint with an untrusted certificate must fail TLS
// verification before any Authorization header reaches the server.
func TestCloneRejectsUntrustedTLSBeforeSendingCredentials_GoGit(t *testing.T) {
	t.Setenv(experimentalUseGitCLIEnv, "false")
	t.Setenv(experimentalUseGitCloneCLIEnv, "false")

	server, receivedAuth := newChallengingTLSServer(t)

	svc := &Service{WorkDir: t.TempDir()}
	err := svc.CloneRepository(&gitprovider.GitRepository{Url: server.URL}, testCreds)
	require.Error(t, err, "expected clone to fail TLS verification")
	require.Empty(t, receivedAuth.Load().(string),
		"GHSA-375h-72g4-hc9c regression: credentials leaked to untrusted TLS endpoint")
}

// unsetEnvForTest fully removes an env var for the test's lifetime and
// restores any prior value at cleanup. t.Setenv(key, "") sets the var to an
// empty string but leaves it defined — and native git treats GIT_SSL_NO_VERIFY
// as "skip verify" whenever it's defined at all (presence-based, not
// value-based), so empty or "false" would silently disable TLS verification.
func unsetEnvForTest(t *testing.T, key string) {
	t.Helper()
	prev, had := os.LookupEnv(key)
	require.NoError(t, os.Unsetenv(key))
	t.Cleanup(func() {
		if had {
			_ = os.Setenv(key, prev)
		} else {
			_ = os.Unsetenv(key)
		}
	})
}

// Regression for GHSA-375h-72g4-hc9c on the native-git CLI clone path. Forces
// the CLI mode AND isolates from any inherited native-git escape valves
// (GIT_SSL_NO_VERIFY env, ~/.gitconfig / /etc/gitconfig http.sslVerify=false)
// so the test exercises the safe-default code path regardless of the
// developer/CI environment.
func TestCloneRejectsUntrustedTLSBeforeSendingCredentials_CLI(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git binary not found in PATH")
	}

	t.Setenv(experimentalUseGitCLIEnv, "true")
	unsetEnvForTest(t, "GIT_SSL_NO_VERIFY")
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_SYSTEM", os.DevNull)
	t.Setenv("HOME", t.TempDir())

	server, receivedAuth := newChallengingTLSServer(t)

	svc := &Service{WorkDir: t.TempDir()}
	err := svc.CloneRepository(&gitprovider.GitRepository{Url: server.URL}, testCreds)
	require.Error(t, err, "expected clone to fail TLS verification")
	require.Empty(t, receivedAuth.Load().(string),
		"GHSA-375h-72g4-hc9c regression: credentials leaked to untrusted TLS endpoint")
}
