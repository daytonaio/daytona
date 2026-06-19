// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package proxy

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
	"github.com/daytonaio/common-go/pkg/utils"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const (
	SignatureQueryParam = "signature"
	ExpiresQueryParam   = "expires"

	signatureV1Prefix  = "v1_"
	signingKeyCacheTTL = 30 * time.Second
)

// parseExpires reads and validates the `expires` query parameter (a unix timestamp).
// An expires of 0 means the signature never expires.
func parseExpires(ctx *gin.Context) (int64, error) {
	expiresStr := ctx.Query(ExpiresQueryParam)
	if expiresStr == "" {
		return 0, common_errors.NewUnauthorizedError(errors.New("expires query parameter is required"))
	}
	expires, err := strconv.ParseInt(expiresStr, 10, 64)
	if err != nil || expires < 0 {
		return 0, common_errors.NewUnauthorizedError(errors.New("invalid expires query parameter"))
	}
	if expires != 0 && time.Now().Unix() > expires {
		return 0, common_errors.NewUnauthorizedError(errors.New("signature is expired"))
	}
	return expires, nil
}

// stripSignatureParams removes the signature and expires parameters from the forwarded
// query so they never reach the upstream daemon.
func stripSignatureParams(ctx *gin.Context) {
	newQuery := ctx.Request.URL.Query()
	newQuery.Del(SignatureQueryParam)
	newQuery.Del(ExpiresQueryParam)
	ctx.Request.URL.RawQuery = newQuery.Encode()
}

// computeSignature derives the v1 signature over a canonical string using HMAC-SHA256
// with the sandbox signing key. Each signed feature owns its own canonical string format
// (e.g. "v1:files:..."), using a distinct domain label so signatures can never be replayed
// across features.
func computeSignature(signingKey string, canonical string) string {
	mac := hmac.New(sha256.New, []byte(signingKey))
	mac.Write([]byte(canonical))
	return signatureV1Prefix + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func verifySignature(signingKey, canonical, signature string) bool {
	expected := computeSignature(signingKey, canonical)
	return hmac.Equal([]byte(expected), []byte(signature))
}

// verifySignedRequest authenticates a request signed with the sandbox's signing key over
// the given canonical string. On a signature mismatch it refetches the key once (to recover
// from a rotation) before failing.
func (p *Proxy) verifySignedRequest(ctx context.Context, sandboxId, canonical, signature string) error {
	if signature == "" {
		return common_errors.NewUnauthorizedError(errors.New("signature is required"))
	}
	if !strings.HasPrefix(signature, signatureV1Prefix) {
		return common_errors.NewUnauthorizedError(errors.New("unsupported signature scheme"))
	}

	signingKey, err := p.getSandboxSigningKey(ctx, sandboxId)
	if err != nil {
		return common_errors.NewUnauthorizedError(fmt.Errorf("failed to resolve signing key: %w", err))
	}
	if verifySignature(signingKey, canonical, signature) {
		return nil
	}

	// A mismatch may mean the cached key is stale because the sandbox's signing key was
	// rotated. Refetch from the API (bypassing the cache) and retry once.
	refreshed, ok := p.refetchSigningKeyOnMismatch(ctx, sandboxId, signingKey)
	if ok && verifySignature(refreshed, canonical, signature) {
		return nil
	}
	return common_errors.NewUnauthorizedError(errors.New("invalid signature"))
}

func (p *Proxy) getSandboxSigningKey(ctx context.Context, sandboxId string) (string, error) {
	cached, err := p.sandboxSigningKeyCache.Get(ctx, sandboxId)
	if err == nil && cached != nil && *cached != "" {
		return *cached, nil
	}
	return p.fetchSandboxSigningKey(ctx, sandboxId)
}

// fetchSandboxSigningKey fetches the signing key from the API, bypassing the cache, and
// refreshes the cache with the result.
func (p *Proxy) fetchSandboxSigningKey(ctx context.Context, sandboxId string) (string, error) {
	var signingKey string
	err := utils.RetryWithExponentialBackoff(ctx, "fetchSandboxSigningKey", proxyMaxRetries, proxyBaseDelay, proxyMaxDelay, func() error {
		key, res, err := p.apiclient.PreviewAPI.GetSigningKey(ctx, sandboxId).Execute()
		if res != nil && res.StatusCode == http.StatusOK {
			signingKey = key
			return nil
		}
		openapiErr := common_errors.ConvertOpenAPIError(err)
		if openapiErr != nil {
			if res != nil && res.StatusCode >= 400 && res.StatusCode < 500 &&
				res.StatusCode != http.StatusRequestTimeout && res.StatusCode != http.StatusTooManyRequests {
				return &utils.NonRetryableError{Err: openapiErr}
			}
			if !common_errors.IsRetryableOpenAPIError(openapiErr) {
				return &utils.NonRetryableError{Err: openapiErr}
			}
			return openapiErr
		}
		return errors.New("failed to fetch sandbox signing key")
	})
	if err != nil {
		return "", err
	}
	if signingKey == "" {
		return "", errors.New("empty sandbox signing key")
	}

	if cacheErr := p.sandboxSigningKeyCache.Set(ctx, sandboxId, signingKey, signingKeyCacheTTL); cacheErr != nil {
		log.Errorf("Failed to set sandbox signing key in cache: %v", cacheErr)
	}

	return signingKey, nil
}

// refetchSigningKeyOnMismatch recovers from a rotated key by fetching a fresh one from the
// API. Each refetch hits the preview controller's Redis cache (sub-ms, no DB load), giving
// the same DoS surface as the existing auth-key validation flow.
func (p *Proxy) refetchSigningKeyOnMismatch(ctx context.Context, sandboxId, staleKey string) (string, bool) {
	fresh, err := p.fetchSandboxSigningKey(ctx, sandboxId)
	if err != nil {
		log.Errorf("Failed to refetch sandbox signing key for %s: %v", sandboxId, err)
		return "", false
	}
	if fresh == staleKey {
		return "", false
	}
	return fresh, true
}
