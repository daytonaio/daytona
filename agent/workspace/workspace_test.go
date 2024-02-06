// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"os"

	"github.com/daytonaio/daytona/credentials"

	"testing"

	"github.com/stretchr/testify/assert"
)

type MockCredentials struct {
}

func (c MockCredentials) GetAccessToken() (*credentials.Credential, error) {
	return nil, nil
}

func TestInitWorkspace(t *testing.T) {
	cwd := "/tmp/test_workspace"
	err := os.RemoveAll(cwd)
	if err != nil {
		t.Fatal(err)
	}

	err = os.MkdirAll(cwd, 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Create test data
	creds := &MockCredentials{}

	workspace, err := New(WorkspaceParams{
		Cwd:         cwd,
		Credentials: creds,
		Extensions:  []Extension{},
		Repositories: []Repository{
			{
				Url: "https://github.com/microsoft/vscode-remote-try-go",
			},
		},
	})
	if err != nil {
		panic(err)
	}

	err = workspace.Create()

	assert.NoError(t, err)

	// Add assertions here to verify the behavior of InitWorkspace
}
