// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package proxy

import (
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mssola/useragent"
)

const (
	SKIP_PREVIEW_WARNING_HEADER      = "X-Daytona-Skip-Preview-Warning"
	PREVIEW_PAGE_ACCEPT_COOKIE_NAME  = "daytona-preview-page-accepted"
	PREVIEW_PAGE_COOKIE_MAX_AGE      = 1 * 24 * 60 * 60 // 1 day in seconds
	ACCEPT_PREVIEW_PAGE_WARNING_PATH = "/accept-daytona-preview-warning"
)

func handleAcceptProxyWarning(ctx *gin.Context, secure bool) {
	// Set SameSite attribute based on security context
	if secure {
		// For HTTPS, use SameSite=None to allow cross-origin iframe usage
		ctx.SetSameSite(http.SameSiteNoneMode)
	} else {
		// For HTTP (local dev), use SameSite=Lax
		ctx.SetSameSite(http.SameSiteLaxMode)
	}

	// Set the acceptance cookie
	ctx.SetCookie(
		PREVIEW_PAGE_ACCEPT_COOKIE_NAME,
		"true",
		PREVIEW_PAGE_COOKIE_MAX_AGE,
		"/",
		strings.Split(ctx.Request.Host, ":")[0],
		secure,
		true,
	)

	// Redirect to the original URL or root
	redirectURL := ctx.Query("redirect")
	if redirectURL == "" {
		redirectURL = "/"
	}

	ctx.Redirect(http.StatusFound, redirectURL)
}

// browserWarningMiddleware is the middleware that checks for browsers and shows warning
func (p *Proxy) browserWarningMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx.Request.Header.Get(SKIP_PREVIEW_WARNING_HEADER) == "true" {
			ctx.Next()
			return
		}

		userAgent := ctx.Request.UserAgent()

		// Skip warning for non-browser requests
		if !isBrowser(userAgent) {
			ctx.Next()
			return
		}

		// Skip warning for WebSocket requests
		if isWebSocketRequest(ctx.Request) {
			ctx.Next()
			return
		}

		// Skip warning if user has already accepted
		if hasAcceptedWarning(ctx) {
			ctx.Next()
			return
		}

		// Skip warning for the acceptance endpoint itself or auth callbacks
		targetPort, _, err := p.parseHost(ctx.Request.Host)
		if err != nil {
			switch ctx.Request.Method {
			case "GET":
				switch ctx.Request.URL.Path {
				case "/callback", "/health":
					ctx.Next()
					return
				}
			}
		}

		if ctx.Request.URL.Path == ACCEPT_PREVIEW_PAGE_WARNING_PATH || targetPort == TERMINAL_PORT {
			ctx.Next()
			return
		}

		// Serve the warning page
		serveWarningPage(ctx, p.config.EnableTLS)
		ctx.Abort() // Stop further processing
	}
}

// serveWarningPage serves the static HTML warning page
func serveWarningPage(c *gin.Context, https bool) {
	htmlContent := `<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>Daytona Preview - Warning</title>
    <style>
      * {
        margin: 0;
        padding: 0;
        box-sizing: border-box;
      }

      body {
        font-family:
          -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', 'Oxygen', 'Ubuntu', 'Cantarell', sans-serif;
        background: #0a0a0a;
        color: #ffffff;
        min-height: 100vh;
        display: flex;
        flex-direction: column;
      }

      .header {
        padding: 1rem 2rem;
        border-bottom: 1px solid #1a1a1a;
        background: #000;
      }

      .logo {
        display: flex;
        align-items: center;
        font-size: 1.5rem;
        font-weight: 600;
        color: #fff;
      }

      .logo::before {
        content: '⚡';
        margin-right: 0.5rem;
        font-size: 1.8rem;
      }

      .container {
        flex: 1;
        display: flex;
        align-items: center;
        justify-content: center;
        padding: 2rem;
      }

      .warning-card {
        background: #1a1a1a;
        border: 1px solid #333;
        border-radius: 12px;
        padding: 3rem 2.5rem;
        max-width: 600px;

        text-align: center;
        box-shadow: 0 20px 40px rgba(0, 0, 0, 0.5);
      }

      .warning-icon {
        font-size: 4rem;
        margin-bottom: 1.5rem;
        color: #ffa500;
      }

      .warning-title {
        font-size: 2rem;
        font-weight: 700;
        margin-bottom: 1rem;
        color: #fff;
      }

      .warning-subtitle {
        font-size: 1.1rem;

        margin-bottom: 2rem;
        line-height: 1.5;
      }

      .warning-text {
        font-size: 0.95rem;
        color: #ccc;
        line-height: 1.6;
        margin-bottom: 2.5rem;
        text-align: left;
        background: #0f0f0f;
        padding: 1.5rem;
        border-radius: 8px;
        border: 1px solid #2a2a2a;
      }

      .warning-text strong {
        color: #ffa500;
      }

      .button-container {
        display: flex;
        gap: 1rem;
        justify-content: center;
        flex-wrap: wrap;
      }

      .btn {
        padding: 0.875rem 2rem;
        border: none;
        border-radius: 8px;
        font-size: 1rem;
        font-weight: 600;
        cursor: pointer;
        transition: all 0.2s ease;
        text-decoration: none;
        display: inline-flex;
        align-items: center;
        gap: 0.5rem;
        min-width: 160px;
        justify-content: center;
      }

      .btn-primary {
        background: #0066ff;
        color: white;
      }

      .footer {
        padding: 1rem 2rem;
        text-align: center;
        font-size: 0.85rem;
        color: #666;
        border-top: 1px solid #1a1a1a;
      }

      a {
        color: #ffffff;
      }

      @media (max-width: 768px) {
        .container {
          padding: 1rem;
        }

        .warning-card {
          padding: 2rem 1.5rem;
        }

        .warning-title {
          font-size: 1.5rem;
        }

        .button-container {
          flex-direction: column;
        }
      }
    </style>
  </head>
  <body>
    <div class="container">
      <div class="warning-card">
        <div class="warning-icon">⚠️</div>
        <h1 class="warning-title">Preview URL Warning</h1>
        <p class="warning-subtitle">You are about to visit <strong>%s</strong></p>

        <div class="warning-text">
          • This website is served through <a href="https://daytona.io" target="_blank">daytona.io</a><br />
          • Content and functionality may change without notice<br />
          • You should only visit this website if you trust whoever sent the link to<br />
          • Be careful about disclosing personal or financial information like passwords, phone numbers, or credit cards<br />
          • To get rid of this warning for your organization, visit our docs: <a href="https://daytona.io/docs/en/preview-and-authentication" target="_blank">https://daytona.io/docs/en/preview-and-authentication</a>
        </div>

        <form action="%s" method="POST" style="margin: 0">
          <div class="button-container">
            <button type="submit" class="btn btn-primary">I Understand, Continue</button>
          </div>
        </form>
      </div>
    </div>

    <div class="footer">Powered by Daytona - Secure and Elastic Infrastructure for AI-Generated Code</div>
  </body>
</html>`

	protocol := "http://"
	if https {
		protocol = "https://"
	}

	redirectPath := protocol + c.Request.Host + c.Request.URL.String()
	redirectUrl := ACCEPT_PREVIEW_PAGE_WARNING_PATH + "?redirect=" + url.QueryEscape(redirectPath)

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, fmt.Sprintf(htmlContent, redirectPath, redirectUrl))
}

// isBrowser checks if the request is coming from a web browser
func isBrowser(userAgent string) bool {
	ua := useragent.New(userAgent)

	browser, _ := ua.Browser()
	browser = strings.ToLower(browser)

	browsers := []string{"chrome", "firefox", "safari", "edge", "brave", "vivaldi", "samsung", "opera"}

	return slices.Contains(browsers, browser)
}

// hasAcceptedWarning checks if the user has already accepted the warning
func hasAcceptedWarning(c *gin.Context) bool {
	cookie, err := c.Cookie(PREVIEW_PAGE_ACCEPT_COOKIE_NAME)
	return err == nil && cookie == "true"
}

// isWebSocketRequest checks if the request is a WebSocket upgrade request
func isWebSocketRequest(req *http.Request) bool {
	connection := strings.ToLower(req.Header.Get("Connection"))
	upgrade := strings.ToLower(req.Header.Get("Upgrade"))

	return upgrade == "websocket" && connection == "upgrade"
}
