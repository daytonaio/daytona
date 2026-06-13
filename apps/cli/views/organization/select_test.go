// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package organization_test

import (
	"errors"
	"testing"

	"github.com/daytonaio/daytona/cli/internal"
	"github.com/daytonaio/daytona/cli/internal/clierr"
	"github.com/daytonaio/daytona/cli/views/organization"
	apiclient "github.com/daytonaio/daytona/libs/api-client-go"
)

func TestGetOrganizationIdFromPromptNonInteractive(t *testing.T) {
	prev := internal.NoInput
	internal.NoInput = true
	t.Cleanup(func() { internal.NoInput = prev })

	org, err := organization.GetOrganizationIdFromPrompt([]apiclient.Organization{
		{Id: "org-1", Name: "Org One"},
		{Id: "org-2", Name: "Org Two"},
	})
	if org != nil {
		t.Errorf("GetOrganizationIdFromPrompt() in non-interactive mode returned %v, want nil", org)
	}

	var cliErr *clierr.Error
	if !errors.As(err, &cliErr) {
		t.Fatalf("expected *clierr.Error, got %T: %v", err, err)
	}
	if cliErr.Category != clierr.CategoryUsage {
		t.Errorf("category = %q, want %q", cliErr.Category, clierr.CategoryUsage)
	}
	if cliErr.Hint != "pass the organization ID or name as an argument" {
		t.Errorf("hint = %q, want %q", cliErr.Hint, "pass the organization ID or name as an argument")
	}
}
