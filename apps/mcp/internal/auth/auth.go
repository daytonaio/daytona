// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
)

// CustomClaims contains custom claims for our JWT
type CustomClaims struct {
	Scope string `json:"scope"`
}

// Validate does nothing for this example, but we need it to satisfy validator.CustomClaims interface
func (c CustomClaims) Validate(ctx context.Context) error {
	return nil
}

// HandleOAuthProtectedResource returns the OAuth protected resource metadata
// This endpoint is required by the MCP specification (RFC 9728)
func HandleOAuthProtectedResource(auth0Domain string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Determine the base path for this specific MCP server
		// Extract the path before /.well-known/oauth-protected-resource
		requestPath := r.URL.Path
		wellKnownPath := "/.well-known/oauth-protected-resource"
		basePath := requestPath[:len(requestPath)-len(wellKnownPath)]

		// Derive base URL from request
		scheme := "https"
		if r.TLS == nil {
			scheme = "http"
		}
		baseURL := fmt.Sprintf("%s://%s", scheme, r.Host)

		// Construct URLs
		authorizationServerURL := fmt.Sprintf("%s%s/.well-known/oauth-authorization-server", baseURL, basePath)
		resourceIdentifier := fmt.Sprintf("%s%s", baseURL, basePath)

		// RFC 9728 format for protected resource metadata
		metadata := map[string]interface{}{
			"resource": resourceIdentifier,
			"authorization_servers": []map[string]string{
				{
					"issuer":                     fmt.Sprintf("https://%s", auth0Domain),
					"authorization_endpoint_uri": authorizationServerURL,
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(metadata)
	}
}

// CreateAuthMiddleware creates the Auth0 middleware with proper WWW-Authenticate header
func CreateAuthMiddleware(auth0Domain, auth0ClientId, auth0Audience string) func(http.Handler) http.Handler {
	issuerURL, err := url.Parse(auth0Domain)
	if err != nil {
		log.Fatalf("Failed to parse issuer URL: %v", err)
	}

	provider := jwks.NewCachingProvider(issuerURL, 5*time.Minute)

	jwtValidator, err := validator.New(
		provider.KeyFunc,
		validator.RS256,
		auth0Domain,
		[]string{auth0Audience},
		validator.WithCustomClaims(
			func() validator.CustomClaims {
				return &CustomClaims{}
			},
		),
		validator.WithAllowedClockSkew(time.Minute),
	)
	if err != nil {
		log.Fatalf("Failed to create JWT validator: %v", err)
	}

	middleware := jwtmiddleware.New(
		jwtValidator.ValidateToken,
		jwtmiddleware.WithErrorHandler(func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("JWT validation error: %v", err)

			// Determine the base path for this specific MCP server
			requestPath := r.URL.Path
			var basePath string
			if strings.HasPrefix(requestPath, "/daytona/sandbox/") {
				basePath = "/daytona/sandbox"
			} else if strings.HasPrefix(requestPath, "/daytona/fs/") {
				basePath = "/daytona/fs"
			} else if strings.HasPrefix(requestPath, "/daytona/git/") {
				basePath = "/daytona/git"
			} else {
				basePath = "/daytona"
			}

			// Derive base URL from request
			scheme := "https"
			if r.TLS == nil {
				scheme = "http"
			}
			baseURL := fmt.Sprintf("%s://%s", scheme, r.Host)

			// Set WWW-Authenticate header as required by MCP spec
			protectedResourceURL := fmt.Sprintf("%s%s/.well-known/oauth-protected-resource", baseURL, basePath)
			w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Bearer realm="%s"`, protectedResourceURL))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"error":   "Unauthorized",
				"message": "Invalid or missing token",
			})
		}),
	)

	return middleware.CheckJWT
}
