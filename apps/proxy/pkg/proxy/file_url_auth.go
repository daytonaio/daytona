// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package proxy

import (
	"errors"
	"fmt"
	"net/http"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
	"github.com/gin-gonic/gin"
)

const (
	fileUrlDownloadPath = "/files/download"
	fileUrlUploadPath   = "/files/upload-v2"
)

// isSignedFileUrlRequest reports whether the request targets a file route with a
// signature query parameter, i.e. should be authenticated by signature instead
// of the regular auth chain.
func isSignedFileUrlRequest(ctx *gin.Context, targetPath string) bool {
	if ctx.Query(SignatureQueryParam) == "" {
		return false
	}
	switch ctx.Request.Method {
	case http.MethodGet, http.MethodHead:
		return targetPath == fileUrlDownloadPath
	case http.MethodPost:
		return targetPath == fileUrlUploadPath
	}
	return false
}

// fileUrlCanonical builds the "files" domain canonical string. The distinct domain label
// keeps file-URL signatures from being replayable against other signed features. The SDKs
// compute the identical string client-side; the cross-language test vectors lock the format.
func fileUrlCanonical(method, filePath string, expires int64) string {
	return fmt.Sprintf("v1:files:%s:%s:%d", method, filePath, expires)
}

// verifySignedFileUrl authenticates a pre-signed file URL request. On success the signature
// parameters are stripped from the forwarded query.
func (p *Proxy) verifySignedFileUrl(ctx *gin.Context, sandboxId string, targetPath string) error {
	filePath := ctx.Query("path")
	if filePath == "" {
		return common_errors.NewBadRequestError(errors.New("path query parameter is required"))
	}

	expires, err := parseExpires(ctx)
	if err != nil {
		return err
	}

	operation := http.MethodGet
	if targetPath == fileUrlUploadPath {
		operation = http.MethodPost
	}

	canonical := fileUrlCanonical(operation, filePath, expires)
	if err := p.verifySignedRequest(ctx.Request.Context(), sandboxId, canonical, ctx.Query(SignatureQueryParam)); err != nil {
		return err
	}

	stripSignatureParams(ctx)
	return nil
}
