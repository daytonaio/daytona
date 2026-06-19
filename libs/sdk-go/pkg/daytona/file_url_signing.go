// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package daytona

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/url"
	"time"

	"github.com/daytonaio/daytona/libs/sdk-go/pkg/errors"
)

const (
	fileURLSignatureV1Prefix = "v1_"
	defaultFileURLTTLSeconds = 3600
)

func computeFileUrlSignature(signingKey, method, path string, expires int64) string {
	canonical := fmt.Sprintf("v1:files:%s:%s:%d", method, path, expires)
	mac := hmac.New(sha256.New, []byte(signingKey))
	_, _ = mac.Write([]byte(canonical))

	return fileURLSignatureV1Prefix + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func resolveExpires(ttlSeconds *int) (int64, error) {
	if ttlSeconds == nil {
		return time.Now().Unix() + defaultFileURLTTLSeconds, nil
	}

	if *ttlSeconds <= 0 {
		return 0, nil
	}

	return time.Now().Unix() + int64(*ttlSeconds), nil
}

func buildSignedFileUrl(toolboxProxyUrl, sandboxId, operationPath, method, filePath, signingKey string, ttlSeconds *int) (string, error) {
	if signingKey == "" {
		return "", errors.NewDaytonaError("Sandbox signing key is not available. Call RefreshData or fetch the sandbox by ID to load it.", 0, nil)
	}

	expires, err := resolveExpires(ttlSeconds)
	if err != nil {
		return "", err
	}

	signature := computeFileUrlSignature(signingKey, method, filePath, expires)
	query := url.Values{}
	query.Set("path", filePath)
	query.Set("expires", fmt.Sprintf("%d", expires))
	query.Set("signature", signature)

	return fmt.Sprintf("%s/%s%s?%s", toolboxProxyUrl, sandboxId, operationPath, query.Encode()), nil
}
