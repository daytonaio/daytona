// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/daytonaio/daytona/cli/config"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

//go:embed auth_success.html
var successHTML []byte

func StartCallbackServer(expectedState string, callbackPort string) (string, error) {
	var code string
	var err error
	var wg sync.WaitGroup
	wg.Add(1)

	server := &http.Server{Addr: fmt.Sprintf(":%s", callbackPort)}

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

// GeneratePKCEVerifier generates a PKCE code verifier (random string)
func GeneratePKCEVerifier() (string, error) {
	// Generate 32 random bytes (256 bits) for the verifier
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("failed to generate PKCE verifier: %w", err)
	}

	// Base64URL encode without padding
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// GeneratePKCEChallenge generates a PKCE code challenge from a verifier using S256 method
func GeneratePKCEChallenge(verifier string) string {
	// SHA256 hash the verifier
	hash := sha256.Sum256([]byte(verifier))
	// Base64URL encode without padding
	return base64.RawURLEncoding.EncodeToString(hash[:])
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

	// Fetch CLI auth config from API (no authentication required)
	cliAuthConfig, err := config.GetCliAuthConfigFromAPI(activeProfile.Api.Url)
	if err != nil {
		return fmt.Errorf("failed to fetch CLI authentication configuration from API: %w", err)
	}

	// Fetch discovery document to get the actual issuer (handles trailing slash issues)
	discoveryURL := strings.TrimSuffix(cliAuthConfig.Issuer, "/") + "/.well-known/openid-configuration"

	type DiscoveryDoc struct {
		Issuer string `json:"issuer"`
	}

	var discoveryDoc DiscoveryDoc
	resp, err := http.Get(discoveryURL)
	if err == nil && resp.StatusCode == 200 {
		if err := json.NewDecoder(resp.Body).Decode(&discoveryDoc); err == nil && discoveryDoc.Issuer != "" {
			// Use the issuer from the discovery document
			cliAuthConfig.Issuer = discoveryDoc.Issuer
		}
		resp.Body.Close()
	}

	provider, err := oidc.NewProvider(ctx, cliAuthConfig.Issuer)
	if err != nil {
		return fmt.Errorf("failed to initialize OIDC provider: %w", err)
	}

	// Get callback port from CLI config, default to 3009 if not provided
	callbackPort := cliAuthConfig.CallbackPort
	if callbackPort == "" {
		callbackPort = "3009"
	}

	// Configure OAuth2 without client secret (public client)
	oauth2Config := oauth2.Config{
		ClientID:    cliAuthConfig.ClientId,
		RedirectURL: fmt.Sprintf("http://localhost:%s/callback", callbackPort),
		Endpoint:    provider.Endpoint(),
		Scopes:      []string{oidc.ScopeOpenID, oidc.ScopeOfflineAccess, "profile"},
	}

	token := &oauth2.Token{
		RefreshToken: activeProfile.Api.Token.RefreshToken,
	}

	// Refresh token - for public clients, the refresh token itself is sufficient
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
