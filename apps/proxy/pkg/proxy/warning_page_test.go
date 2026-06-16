// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package proxy

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestServeWarningPage_EscapesXSSPayloadInBody(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/?><ScRiPt>alert(1)</ScRiPt>", nil)
	c.Request.Host = "daytonaproxy01.net"

	serveWarningPage(c, false)

	body := w.Body.String()
	if strings.Contains(body, "<ScRiPt>") || strings.Contains(body, "<script>") {
		t.Fatalf("XSS payload was rendered raw in warning page body:\n%s", body)
	}
	if !strings.Contains(body, "&lt;ScRiPt&gt;alert(1)&lt;/ScRiPt&gt;") {
		t.Fatalf("expected the script tag to be HTML-escaped; body did not contain the escaped form:\n%s", body)
	}
}

func TestServeWarningPage_EscapesHostHeader(t *testing.T) {
	// Host header is technically attacker-controllable on forged requests. Even
	// though browsers won't normally let an attacker forge it through a victim,
	// escape defensively.
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	c.Request.Host = "evil\"<img src=x>.example.com"

	serveWarningPage(c, false)

	body := w.Body.String()
	if strings.Contains(body, "<img src=x>") {
		t.Fatalf("Host header was rendered raw in warning page body:\n%s", body)
	}
}

func TestServeWarningPage_RendersBenignURLReadably(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/dashboard/index.html", nil)
	c.Request.Host = "3000-abc.daytonaproxy01.net"

	serveWarningPage(c, true)

	body := w.Body.String()
	if !strings.Contains(body, "https://3000-abc.daytonaproxy01.net/dashboard/index.html") {
		t.Fatalf("expected benign redirect path to be rendered as readable text; got:\n%s", body)
	}
}

func TestServeWarningPage_SetsCSPAndNosniffHeaders(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	c.Request.Host = "daytonaproxy01.net"

	serveWarningPage(c, false)

	csp := w.Header().Get("Content-Security-Policy")
	if !strings.Contains(csp, "default-src 'none'") {
		t.Fatalf("expected restrictive CSP on warning page; got %q", csp)
	}
	if got := w.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("expected X-Content-Type-Options: nosniff; got %q", got)
	}
}

func TestServeWarningPage_FormActionUrlEncodesPayload(t *testing.T) {
	// The redirectUrl is built via url.QueryEscape; confirm dangerous chars do
	// not reach the action attribute as a literal '"' or '<'.
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/?\"><script>x</script>", nil)
	c.Request.Host = "daytonaproxy01.net"

	serveWarningPage(c, false)

	body := w.Body.String()
	if strings.Contains(body, `action="`) && strings.Contains(body, `"><script>`) {
		t.Fatalf("form action attribute appears to allow attribute breakout:\n%s", body)
	}
}

func TestHandleAcceptProxyWarning_AllowsSameHostAbsoluteRedirect(t *testing.T) {
	// This is exactly what serveWarningPage produces: an absolute URL on the
	// same host the request arrived on. It must be honored, not downgraded.
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(
		"POST",
		ACCEPT_PREVIEW_PAGE_WARNING_PATH+"?redirect=https%3A%2F%2F3000-abc.daytonaproxy01.net%2Fdashboard%2Findex.html",
		nil,
	)
	c.Request.Host = "3000-abc.daytonaproxy01.net"

	handleAcceptProxyWarning(c, true)

	// gin's ResponseWriter buffers the status; http.Redirect writes no body for
	// POST, so assert the buffered status rather than the recorder's Code.
	if c.Writer.Status() != http.StatusFound {
		t.Fatalf("expected 302 Found; got %d", c.Writer.Status())
	}
	if loc := w.Header().Get("Location"); loc != "https://3000-abc.daytonaproxy01.net/dashboard/index.html" {
		t.Fatalf("expected same-host redirect to be honored; got Location %q", loc)
	}
}

func TestHandleAcceptProxyWarning_AllowsSameHostWithPortAndQuery(t *testing.T) {
	// Dev http://localhost:port and host:port must pass: serveWarningPage and the
	// validator both read the full raw Request.Host (incl. port).
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(
		"POST",
		ACCEPT_PREVIEW_PAGE_WARNING_PATH+"?redirect=http%3A%2F%2Flocalhost%3A8080%2Fdashboard%3Ftab%3Dlogs",
		nil,
	)
	c.Request.Host = "localhost:8080"

	handleAcceptProxyWarning(c, false)

	// gin's ResponseWriter buffers the status; http.Redirect writes no body for
	// POST, so assert the buffered status rather than the recorder's Code.
	if c.Writer.Status() != http.StatusFound {
		t.Fatalf("expected 302 Found; got %d", c.Writer.Status())
	}
	if loc := w.Header().Get("Location"); loc != "http://localhost:8080/dashboard?tab=logs" {
		t.Fatalf("expected same-host:port redirect to be honored; got Location %q", loc)
	}
}

func TestHandleAcceptProxyWarning_AllowsSameHostWithForwardedHostHeader(t *testing.T) {
	// Custom Preview Proxy: the consent POST reaches Daytona on a daytona proxy
	// host; X-Forwarded-Host is never read for host resolution, so the same-host
	// redirect must still be honored.
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(
		"POST",
		ACCEPT_PREVIEW_PAGE_WARNING_PATH+"?redirect=https%3A%2F%2F3000-abc.daytonaproxy01.net%2Fapp",
		nil,
	)
	c.Request.Host = "3000-abc.daytonaproxy01.net"
	c.Request.Header.Set("X-Forwarded-Host", "preview.yourcompany.com")

	handleAcceptProxyWarning(c, true)

	if loc := w.Header().Get("Location"); loc != "https://3000-abc.daytonaproxy01.net/app" {
		t.Fatalf("expected redirect to be unaffected by X-Forwarded-Host; got Location %q", loc)
	}
}

func TestHandleAcceptProxyWarning_AllowsSafeRelativePath(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", ACCEPT_PREVIEW_PAGE_WARNING_PATH+"?redirect=%2Fdashboard", nil)
	c.Request.Host = "3000-abc.daytonaproxy01.net"

	handleAcceptProxyWarning(c, true)

	if loc := w.Header().Get("Location"); loc != "/dashboard" {
		t.Fatalf("expected safe relative path to be honored; got Location %q", loc)
	}
}

func TestHandleAcceptProxyWarning_RejectsOpenRedirectTargets(t *testing.T) {
	cases := []struct {
		name     string
		redirect string
	}{
		{"absolute-cross-host", "https://evil.com/phish"},
		{"protocol-relative", "//evil.com/x"},
		{"backslash-slash", "/\\evil.com"},
		{"double-backslash", "\\\\evil.com"},
		{"userinfo-confusion", "https://3000-abc.daytonaproxy01.net@evil.com"},
		{"subdomain-lookalike", "https://host.evil.com"},
		{"javascript-scheme", "javascript:alert(1)"},
		{"data-scheme", "data:text/html,<script>alert(1)</script>"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("POST", ACCEPT_PREVIEW_PAGE_WARNING_PATH, nil)
			q := c.Request.URL.Query()
			q.Set("redirect", tc.redirect)
			c.Request.URL.RawQuery = q.Encode()
			c.Request.Host = "3000-abc.daytonaproxy01.net"

			handleAcceptProxyWarning(c, true)

			if loc := w.Header().Get("Location"); loc != "/" {
				t.Fatalf("expected unsafe redirect %q to fall back to \"/\"; got Location %q", tc.redirect, loc)
			}
		})
	}
}

func TestHandleAcceptProxyWarning_EmptyRedirectFallsBackToRoot(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", ACCEPT_PREVIEW_PAGE_WARNING_PATH, nil)
	c.Request.Host = "3000-abc.daytonaproxy01.net"

	handleAcceptProxyWarning(c, true)

	if loc := w.Header().Get("Location"); loc != "/" {
		t.Fatalf("expected empty redirect to fall back to \"/\"; got Location %q", loc)
	}
}
