// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

// Package api implements 'daytona api', an escape hatch for making raw
// authenticated requests to the Daytona API. It is its own package because
// package cmd cannot import config (config imports cmd for autocompletion
// setup, which would create an import cycle).
package api

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	apiclient_cli "github.com/daytonaio/daytona/cli/apiclient"
	"github.com/daytonaio/daytona/cli/auth"
	"github.com/daytonaio/daytona/cli/config"
	"github.com/daytonaio/daytona/cli/internal/clierr"
	"github.com/spf13/cobra"
)

var (
	apiMethodFlag string
	apiInputFlag  string
)

var ApiCmd = &cobra.Command{
	Use:   "api PATH",
	Short: "Make an authenticated request to the Daytona API",
	Long: `Make an authenticated HTTP request to the Daytona API and print the raw response body to stdout.

PATH is resolved against the active profile's API URL and the request is authenticated with the active profile's credentials. Responses with status 400 or above still print the body, then exit non-zero.`,
	Example: `  daytona api /sandbox
  daytona api /sandbox/my-sandbox -X DELETE
  daytona api /snapshots -X POST --input snapshot.json
  cat body.json | daytona api /sandbox -X POST --input -`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return clierr.New(clierr.CategoryUsage, "missing required argument: PATH")
		}
		if len(args) > 1 {
			return clierr.Newf(clierr.CategoryUsage, "accepts 1 argument, received %d", len(args))
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		method, err := normalizeMethod(apiMethodFlag)
		if err != nil {
			return err
		}

		body, hasBody, err := resolveBody(method, apiInputFlag, os.Stdin)
		if err != nil {
			return err
		}

		// Refresh the OAuth token up front: GetApiClient only triggers a
		// refresh once a client instance is already cached, so a fresh
		// process would build the raw request below with an expired token.
		// GetApiClient is dropped entirely — its only other effect here was
		// profile validation, which GetConfig/GetActiveProfile below already
		// perform (and RefreshTokenIfNeeded re-checks the credentials).
		if err := auth.RefreshTokenIfNeeded(cmd.Context()); err != nil {
			return err
		}

		c, err := config.GetConfig()
		if err != nil {
			return err
		}

		activeProfile, err := c.GetActiveProfile()
		if err != nil {
			return err
		}

		var bodyReader io.Reader
		if hasBody {
			bodyReader = bytes.NewReader(body)
		}

		req, err := http.NewRequestWithContext(cmd.Context(), method, joinURL(activeProfile.Api.Url, args[0]), bodyReader)
		if err != nil {
			return clierr.New(clierr.CategoryUsage, err.Error())
		}

		if activeProfile.Api.Key != nil {
			req.Header.Set("Authorization", "Bearer "+*activeProfile.Api.Key)
		} else if activeProfile.Api.Token != nil {
			req.Header.Set("Authorization", "Bearer "+activeProfile.Api.Token.AccessToken)
			if activeProfile.ActiveOrganizationId != nil {
				req.Header.Set("X-Daytona-Organization-ID", *activeProfile.ActiveOrganizationId)
			}
		}
		req.Header.Set(apiclient_cli.DaytonaSourceHeader, "cli")
		req.Header.Set("Accept", "application/json")
		if hasBody {
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err := (&http.Client{}).Do(req)
		if err != nil {
			return clierr.New(clierr.CategoryNetwork, err.Error())
		}
		defer resp.Body.Close()

		out := &trailingNewlineWriter{w: os.Stdout}
		if _, err := io.Copy(out, resp.Body); err != nil {
			return clierr.New(clierr.CategoryNetwork, err.Error())
		}
		if err := out.finish(); err != nil {
			return err
		}

		if resp.StatusCode >= 400 {
			return clierr.FromHTTPStatus(resp.StatusCode, fmt.Sprintf("HTTP %d %s", resp.StatusCode, http.StatusText(resp.StatusCode)))
		}

		return nil
	},
}

func init() {
	ApiCmd.Flags().StringVarP(&apiMethodFlag, "method", "X", http.MethodGet, "HTTP method (GET, POST, PUT, PATCH, DELETE, HEAD)")
	ApiCmd.Flags().StringVar(&apiInputFlag, "input", "", "Request body source: a file path or '-' for stdin (POST, PUT, and PATCH only)")
}

// joinURL resolves an API path against the profile base URL, tolerating a
// trailing slash on the base and a leading slash on the path.
func joinURL(base, path string) string {
	return strings.TrimSuffix(base, "/") + "/" + strings.TrimPrefix(path, "/")
}

var apiAllowedMethods = []string{
	http.MethodGet,
	http.MethodPost,
	http.MethodPut,
	http.MethodPatch,
	http.MethodDelete,
	http.MethodHead,
}

// normalizeMethod uppercases the --method value and validates it against the
// supported set.
func normalizeMethod(method string) (string, error) {
	normalized := strings.ToUpper(strings.TrimSpace(method))
	for _, m := range apiAllowedMethods {
		if normalized == m {
			return normalized, nil
		}
	}
	return "", clierr.Newf(clierr.CategoryUsage, "invalid --method value %q: must be one of GET, POST, PUT, PATCH, DELETE, HEAD", method)
}

// resolveBody resolves the --input flag into the request body bytes. input is
// a file path or "-" for stdin and is only valid for methods that carry a
// body (POST, PUT, PATCH).
func resolveBody(method, input string, stdin io.Reader) (body []byte, hasBody bool, err error) {
	if input == "" {
		return nil, false, nil
	}

	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch:
	default:
		return nil, false, clierr.Newf(clierr.CategoryUsage, "--input cannot be used with %s: only POST, PUT, and PATCH requests carry a body", method)
	}

	if input == "-" {
		data, err := io.ReadAll(stdin)
		if err != nil {
			return nil, false, clierr.Newf(clierr.CategoryUsage, "failed to read request body from stdin: %s", err)
		}
		return data, true, nil
	}

	data, err := os.ReadFile(input)
	if err != nil {
		return nil, false, clierr.Newf(clierr.CategoryUsage, "failed to read request body from %s: %s", input, err)
	}
	return data, true, nil
}

// trailingNewlineWriter passes bytes through to w, tracking whether any
// output was produced and whether it ended with a newline so finish can
// terminate a partial last line.
type trailingNewlineWriter struct {
	w     io.Writer
	wrote bool
	last  byte
}

func (t *trailingNewlineWriter) Write(p []byte) (int, error) {
	n, err := t.w.Write(p)
	if n > 0 {
		t.wrote = true
		t.last = p[n-1]
	}
	return n, err
}

// finish writes a trailing newline when output was produced without one.
func (t *trailingNewlineWriter) finish() error {
	if !t.wrote || t.last == '\n' {
		return nil
	}
	_, err := t.w.Write([]byte{'\n'})
	return err
}
