// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package proxy

import (
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
