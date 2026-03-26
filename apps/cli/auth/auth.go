// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package auth

import (
	"context"
	"crypto/rand"
	_ "embed"
	"encoding/base64"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/daytonaio/daytona/cli/config"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

//go:embed auth_success.html
var successHTML []byte

func StartCallbackServer(expectedState string) (string, error) {
	var code string
	var err error
	var wg sync.WaitGroup
	wg.Add(1)

	server := &http.Server{Addr: fmt.Sprintf(":%s", config.GetAuth0CallbackPort())}

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("state") != expectedState {
			err = fmt.Errorf("invalid state parameter")
			http.Error(w, "State invalid", http.StatusBadRequest)
			wg.Done()
			return
		}

		code = r.URL.Query().Get("code")
		if code == "" {
			err = fmt.Errorf("no code in callback")
			http.Error(w, "No code", http.StatusBadRequest)
			wg.Done()
			return
		}

		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write(successHTML)

		// Delay server close to ensure browser receives the success page
		go func() {
			time.Sleep(500 * time.Millisecond)
			wg.Done()
			server.Close()
		}()
	})

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Errorf("HTTP server error: %v", err)
		}
	}()
	wg.Wait()

	if err != nil {
		return "", err
	}
	return code, nil
}

func GenerateRandomState() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(b), nil
}

func RefreshTokenIfNeeded(ctx context.Context) error {
	c, err := config.GetConfig()
	if err != nil {
		return err
	}

	activeProfile, err := c.GetActiveProfile()
	if err != nil {
		return err
	}

	if activeProfile.Api.Key != nil {
		return nil
	}

	if activeProfile.Api.Token == nil {
		return fmt.Errorf("no valid token found, use 'daytona login' to reauthenticate")
	}

	// Check if token is about to expire (within 5 minutes)
	if time.Until(activeProfile.Api.Token.ExpiresAt) > 5*time.Minute {
		return nil
	}

	provider, err := oidc.NewProvider(ctx, config.GetAuth0Domain())
	if err != nil {
		return fmt.Errorf("failed to initialize OIDC provider: %w", err)
	}

	oauth2Config := oauth2.Config{
		ClientID:     config.GetAuth0ClientId(),
		ClientSecret: config.GetAuth0ClientSecret(),
		RedirectURL:  fmt.Sprintf("http://localhost:%s/callback", config.GetAuth0CallbackPort()),
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, oidc.ScopeOfflineAccess, "profile"},
	}

	token := &oauth2.Token{
		RefreshToken: activeProfile.Api.Token.RefreshToken,
	}

	newToken, err := oauth2Config.TokenSource(ctx, token).Token()
	if err != nil {
		return fmt.Errorf("use 'daytona login' to reauthenticate: %w", err)
	}

	activeProfile.Api.Token = &config.Token{
		AccessToken:  newToken.AccessToken,
		RefreshToken: newToken.RefreshToken,
		ExpiresAt:    newToken.Expiry,
	}

	return c.EditProfile(activeProfile)
}
