// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package git

import (
	"fmt"

	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

// Authenticate persists Git credentials globally (via the "store" helper) so
// future operations against the host authenticate automatically. The credential
// is approved through the git CLI — go-git has no credential subsystem — and
// passed on stdin, never argv.
func (s *Service) Authenticate(username, password, host, protocol string) error {
	if host == "" {
		host = "github.com"
	}
	if protocol == "" {
		protocol = "https"
	}

	if err := s.SetConfigValue("credential.helper", "store", "global"); err != nil {
		return err
	}

	input := fmt.Sprintf("protocol=%s\nhost=%s\nusername=%s\npassword=%s\n\n", protocol, host, username, password)
	return s.runGitCLI(gitCLIOptions{
		op:       "git credential approve",
		args:     []string{"credential", "approve"},
		stdin:    input,
		redact:   &http.BasicAuth{Username: username, Password: password},
		tailSize: 16 * 1024,
	})
}
