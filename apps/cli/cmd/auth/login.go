// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package auth

import (
	"context"
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/daytonaio/daytona/cli/auth"
	"github.com/daytonaio/daytona/cli/cmd/common"
	"github.com/daytonaio/daytona/cli/config"
	"github.com/daytonaio/daytona/cli/internal"
	view_common "github.com/daytonaio/daytona/cli/views/common"
	"github.com/iancoleman/strcase"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

var (
	apiKeyFlag string
)

var LoginCmd = &cobra.Command{
	Use:     "login",
	Short:   "Log in to Daytona",
	Args:    cobra.NoArgs,
	GroupID: internal.USER_GROUP,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		if apiKeyFlag != "" {
			return upsertProfileWithLogin(nil, &apiKeyFlag)
		}

		items := []view_common.SelectItem{
			{Title: "Login with Browser", Desc: "Authenticate using OAuth in your browser"},
			{Title: "Set Daytona API Key", Desc: "Authenticate using Daytona API key"},
		}

		choice, err := view_common.Select("Select Authentication Method", items)
		if err != nil {
			return fmt.Errorf("error running selection prompt: %w", err)
		}

		if choice == "" {
			return nil
		}

		var tokenConfig *config.Token
		setApiKey := choice == "Set Daytona API Key"

		if setApiKey {
			// Prompt for API key
			apiKey, err := view_common.PromptForInput(
				"",
				"Enter your Daytona API key",
				"You can find it in the Daytona dashboard - https://app.daytona.io/dashboard",
				"",
			)
			if err != nil {
				return err
			}
			return upsertProfileWithLogin(nil, &apiKey)
		}

		token, err := login(ctx)
		if err != nil {
			return err
		}

		tokenConfig = &config.Token{
			AccessToken:  token.AccessToken,
			RefreshToken: token.RefreshToken,
			ExpiresAt:    token.Expiry,
		}

		return upsertProfileWithLogin(tokenConfig, nil)
	},
}

func init() {
	LoginCmd.Flags().StringVar(&apiKeyFlag, "api-key", "", "API key to use for authentication")
}

func upsertProfileWithLogin(tokenConfig *config.Token, apiKey *string) error {
	c, err := config.GetConfig()
	if err != nil {
		return err
	}

	var shouldOverwriteProfile bool
	var newProfile bool
	var profileName string
	var profileId string

	// Loop until user decides on the name, i.e. if user is prompted with the
	// decision to overwrite the existing profile (in case user chose name that
	// already exists) they can select "NO" which would loop back to the prompt
	// to define the profile name.
	for !shouldOverwriteProfile {
		profileName, err = view_common.PromptForInput(
			"",
			"Enter profile name",
			"The name this profile will be known as",
			"Default",
		)
		if err != nil {
			return err
		}
		if profileName == "" {
			profileName = "Default"
		}

		// Check if profile already exists and inform user that the profile will be
		// reauthenticated if it does.
		profileId = strcase.ToSnake(profileName)
		if c.ProfileExists(profileId) {
			err = huh.NewConfirm().
				Title(fmt.Sprintf(
					"Profile %s already exists, do you want to overwrite it?",
					profileId,
				)).
				Affirmative("Yes").
				Negative("No").
				Value(&shouldOverwriteProfile).
				Run()
			if err != nil {
				return err
			}
		} else {
			shouldOverwriteProfile = true
			newProfile = true
		}
	}

	var activeProfile config.Profile

	if newProfile {
		activeProfile, err = config.CreateProfile(profileName, c)
		if err != nil {
			return err
		}
	} else {
		activeProfile, err = c.GetProfile(profileId)
		if err != nil {
			// Should not occur
			return err
		}
	}

	if apiKey != nil {
		activeProfile.Api.Token = nil
		activeProfile.Api.Key = apiKey

		view_common.RenderInfoMessageBold("Successfully set Daytona API key!")
	}

	if tokenConfig != nil {
		activeProfile.Api.Key = nil
		activeProfile.Api.Token = tokenConfig

		err = c.EditProfile(activeProfile)
		if err != nil {
			return err
		}

		if activeProfile.Api.Key == nil {
			personalOrganizationId, err := common.GetPersonalOrganizationId(activeProfile)
			if err != nil {
				return err
			}

			activeProfile.ActiveOrganizationId = &personalOrganizationId
		}
	}

	return c.EditProfile(activeProfile)
}

func login(ctx context.Context) (*oauth2.Token, error) {
	provider, err := oidc.NewProvider(ctx, config.GetAuth0Domain())
	if err != nil {
		return nil, fmt.Errorf("failed to initialize OIDC provider: %w", err)
	}

	verifier := provider.Verifier(&oidc.Config{ClientID: config.GetAuth0ClientId()})

	oauth2Config := oauth2.Config{
		ClientID:     config.GetAuth0ClientId(),
		ClientSecret: config.GetAuth0ClientSecret(),
		RedirectURL:  fmt.Sprintf("http://localhost:%s/callback", config.GetAuth0CallbackPort()),
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, oidc.ScopeOfflineAccess, "profile"},
	}

	state, err := auth.GenerateRandomState()
	if err != nil {
		return nil, fmt.Errorf("failed to generate random state: %w", err)
	}

	authURL := oauth2Config.AuthCodeURL(
		state,
		oauth2.SetAuthURLParam("audience", config.GetAuth0Audience()),
	)

	view_common.RenderInfoMessageBold("Opening the browser for authentication ...")

	view_common.RenderInfoMessage("If opening fails, visit:\n")

	fmt.Println(authURL)

	_ = browser.OpenURL(authURL)

	code, err := auth.StartCallbackServer(state)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	token, err := oauth2Config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange token: %w", err)
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("no id_token in token response")
	}

	_, err = verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("failed to verify ID token: %w", err)
	}

	view_common.RenderInfoMessageBold("Successfully logged in!")

	return token, nil
}
