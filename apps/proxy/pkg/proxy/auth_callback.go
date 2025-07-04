// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package proxy

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"

	"github.com/daytonaio/apiclient"
	common_errors "github.com/daytonaio/common-go/pkg/errors"
)

func (p *Proxy) AuthCallback(ctx *gin.Context) {
	if ctx.Query("error") != "" {
		err := ctx.Query("error")
		if ctx.Query("error_description") != "" {
			err = ctx.Query("error_description")
		}
		ctx.Error(common_errors.NewUnauthorizedError(errors.New(err)))
		return
	}

	code := ctx.Query("code")
	if code == "" {
		ctx.Error(common_errors.NewBadRequestError(errors.New("no code in callback")))
		return
	}

	state := ctx.Query("state")
	if state == "" {
		ctx.Error(common_errors.NewBadRequestError(errors.New("no state in callback")))
		return
	}

	// Decode state
	stateJson, err := base64.URLEncoding.DecodeString(state)
	if err != nil {
		ctx.Error(common_errors.NewBadRequestError(fmt.Errorf("failed to decode state: %w", err)))
		return
	}

	var stateData map[string]string
	err = json.Unmarshal(stateJson, &stateData)
	if err != nil {
		ctx.Error(common_errors.NewBadRequestError(fmt.Errorf("failed to unmarshal state: %w", err)))
		return
	}

	returnTo := stateData["returnTo"]
	if returnTo == "" {
		ctx.Error(common_errors.NewBadRequestError(errors.New("no returnTo in state")))
		return
	}

	sandboxId := stateData["sandboxId"]
	if sandboxId == "" {
		ctx.Error(common_errors.NewBadRequestError(errors.New("no sandboxId in state")))
		return
	}

	// Exchange code for token
	provider, err := oidc.NewProvider(ctx, p.config.Oidc.Domain)
	if err != nil {
		ctx.Error(common_errors.NewBadRequestError(fmt.Errorf("failed to initialize OIDC provider: %w", err)))
		return
	}

	oauth2Config := oauth2.Config{
		ClientID:     p.config.Oidc.ClientId,
		ClientSecret: p.config.Oidc.ClientSecret,
		RedirectURL:  fmt.Sprintf("%s://%s/callback", p.config.ProxyProtocol, ctx.Request.Host),
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile"},
	}

	token, err := oauth2Config.Exchange(ctx, code)
	if err != nil {
		ctx.Error(common_errors.NewBadRequestError(fmt.Errorf("failed to exchange token: %w", err)))
		return
	}

	hasAccess := p.hasSandboxAccess(ctx, sandboxId, token.AccessToken)
	if !hasAccess {
		ctx.Error(common_errors.NewNotFoundError(errors.New("sandbox not found")))
		return
	}

	// Set the sandboxId in a cookie for cached access
	cookieDomain := p.config.ProxyDomain
	cookieDomain = strings.Split(cookieDomain, ":")[0]
	cookieDomain = fmt.Sprintf(".%s", cookieDomain)

	encoded, err := p.secureCookie.Encode(DAYTONA_SANDBOX_AUTH_COOKIE_NAME+sandboxId, sandboxId)
	if err != nil {
		ctx.Error(common_errors.NewBadRequestError(fmt.Errorf("failed to encode cookie: %w", err)))
		return
	}

	ctx.SetCookie(DAYTONA_SANDBOX_AUTH_COOKIE_NAME+sandboxId, encoded, 3600, "/", cookieDomain, p.config.EnableTLS, true)

	// Redirect back to the original URL
	ctx.Redirect(http.StatusFound, returnTo)
}

func (p *Proxy) getAuthUrl(ctx *gin.Context, sandboxId string) (string, error) {
	provider, err := oidc.NewProvider(ctx, p.config.Oidc.Domain)
	if err != nil {
		return "", fmt.Errorf("failed to initialize OIDC provider: %w", err)
	}

	oauth2Config := oauth2.Config{
		ClientID:     p.config.Oidc.ClientId,
		ClientSecret: p.config.Oidc.ClientSecret,
		RedirectURL:  fmt.Sprintf("%s://%s/callback", p.config.ProxyProtocol, p.config.ProxyDomain),
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile"},
	}

	state, err := GenerateRandomState()
	if err != nil {
		return "", fmt.Errorf("failed to generate random state: %w", err)
	}

	// Store the original request URL in the state
	stateData := map[string]string{
		"state":     state,
		"returnTo":  fmt.Sprintf("%s://%s%s", p.config.ProxyProtocol, ctx.Request.Host, ctx.Request.URL.String()),
		"sandboxId": sandboxId,
	}
	stateJson, err := json.Marshal(stateData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal state: %w", err)
	}
	encodedState := base64.URLEncoding.EncodeToString(stateJson)

	authURL := oauth2Config.AuthCodeURL(
		encodedState,
		oauth2.SetAuthURLParam("audience", p.config.Oidc.Audience),
	)

	return authURL, nil
}

func (p *Proxy) hasSandboxAccess(ctx context.Context, sandboxId string, authToken string) bool {
	clientConfig := apiclient.NewConfiguration()
	clientConfig.Servers = apiclient.ServerConfigurations{
		{
			URL: p.config.DaytonaApiUrl,
		},
	}
	clientConfig.AddDefaultHeader("Authorization", "Bearer "+authToken)

	apiClient := apiclient.NewAPIClient(clientConfig)

	res, _ := apiClient.PreviewAPI.HasSandboxAccess(ctx, sandboxId).Execute()

	return res != nil && res.StatusCode == http.StatusOK
}

func GenerateRandomState() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(b), nil
}
