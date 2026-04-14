// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package proxy_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/daytonaio/common-go/pkg/proxy"
	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// TestRawParam_PreservesPercentEncoding verifies that RawParam returns the
// original percent-encoded path suffix rather than the decoded form that
// ctx.Param() would return.
func TestRawParam_PreservesPercentEncoding(t *testing.T) {
	cases := []struct {
		name     string
		route    string
		param    string
		reqPath  string
		expected string
	}{
		{
			name:     "encoded @ preserved",
			route:    "/*path",
			param:    "path",
			reqPath:  "/%40topbar/page.js",
			expected: "/%40topbar/page.js",
		},
		{
			name:     "encoded brackets preserved",
			route:    "/*path",
			param:    "path",
			reqPath:  "/%5B%5Bslug%5D%5D/page.js",
			expected: "/%5B%5Bslug%5D%5D/page.js",
		},
		{
			name:    "reproduction case: (local) unencoded, %40 and %5B%5D encoded",
			route:   "/*path",
			param:   "path",
			reqPath: "/_next/static/chunks/app/(local)/%40topbar/%5B%5B...slug%5D%5D/page-fbf946cc1263adfe.js",
			// (local) was never encoded by the client — must stay as-is.
			// %40 and %5B%5D must not be decoded.
			expected: "/_next/static/chunks/app/(local)/%40topbar/%5B%5B...slug%5D%5D/page-fbf946cc1263adfe.js",
		},
		{
			name:     "no special chars unchanged",
			route:    "/*path",
			param:    "path",
			reqPath:  "/static/js/main.js",
			expected: "/static/js/main.js",
		},
		{
			name:     "wildcard with prefix: encoded suffix extracted correctly",
			route:    "/:port/*path",
			param:    "path",
			reqPath:  "/3000/_next/static/chunks/app/(local)/%40topbar/page.js",
			expected: "/_next/static/chunks/app/(local)/%40topbar/page.js",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			router := gin.New()
			var got string
			router.GET(tc.route, func(ctx *gin.Context) {
				got = proxy.RawParam(ctx, tc.param)
				ctx.Status(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, tc.reqPath, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if got != tc.expected {
				t.Errorf("RawParam(%q):\n  got  %q\n  want %q", tc.reqPath, got, tc.expected)
			}
		})
	}
}

// TestNewProxyRequestHandler_PreservesPathEncoding is an end-to-end test that
// verifies the proxy Director propagates RawPath so that the backend receives
// the original percent-encoding, not a re-encoded mutation.
//
// Regression for: https://github.com/daytonaio/daytona/issues/4448
// Mutation observed before fix:
//
//	(local)    → %28local%29   (parentheses incorrectly encoded)
//	%40topbar  → @topbar       (encoded @ incorrectly decoded)
func TestNewProxyRequestHandler_PreservesPathEncoding(t *testing.T) {
	cases := []struct {
		name        string
		incomingURL string // raw URL as sent by the browser
		wantPath    string // exact path the backend must receive
	}{
		{
			name:        "reproduction: (local) stays unencoded, %40 stays encoded",
			incomingURL: "/_next/static/chunks/app/(local)/%40topbar/%5B%5B...slug%5D%5D/page-fbf946cc1263adfe.js",
			wantPath:    "/_next/static/chunks/app/(local)/%40topbar/%5B%5B...slug%5D%5D/page-fbf946cc1263adfe.js",
		},
		{
			name:        "normal static path unchanged",
			incomingURL: "/static/js/bundle.js",
			wantPath:    "/static/js/bundle.js",
		},
		{
			name:        "encoded slash-like chars preserved",
			incomingURL: "/files/%2Ftmp%2Ffoo.txt",
			wantPath:    "/files/%2Ftmp%2Ffoo.txt",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Backend server records the raw request URI it receives.
			var backendReceivedURI string
			backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				backendReceivedURI = r.RequestURI
				w.WriteHeader(http.StatusOK)
			}))
			defer backend.Close()

			backendURL, _ := url.Parse(backend.URL)

			// getProxyTarget returns a target URL that mirrors the incoming path.
			getTarget := func(ctx *gin.Context) (*url.URL, map[string]string, error) {
				rawPath := proxy.RawParam(ctx, "path")
				if rawPath == "" {
					rawPath = "/"
				}
				target, err := url.Parse(backend.URL + rawPath)
				if err != nil {
					return nil, nil, err
				}
				// Preserve host so Director sets it correctly
				target.Host = backendURL.Host
				target.Scheme = backendURL.Scheme
				return target, nil, nil
			}

			router := gin.New()
			router.GET("/*path", proxy.NewProxyRequestHandler(getTarget, nil))

			// Use a real HTTP server for the proxy side: httptest.ResponseRecorder
			// does not implement http.CloseNotifier, which httputil.ReverseProxy
			// requires when flushing through gin's responseWriter.
			proxyServer := httptest.NewServer(router)
			defer proxyServer.Close()

			resp, err := http.Get(proxyServer.URL + tc.incomingURL)
			if err != nil {
				t.Fatalf("proxy request failed: %v", err)
			}
			resp.Body.Close()

			if backendReceivedURI != tc.wantPath {
				t.Errorf("backend received URI:\n  got  %q\n  want %q", backendReceivedURI, tc.wantPath)
			}
		})
	}
}
