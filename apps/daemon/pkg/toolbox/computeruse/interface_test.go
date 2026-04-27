// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeComputerUse struct {
	screenshotReq       *ScreenshotRequest
	regionReq           *RegionScreenshotRequest
	compressedReq       *CompressedScreenshotRequest
	compressedRegionReq *CompressedRegionScreenshotRequest
}

func (f *fakeComputerUse) TakeScreenshot(req *ScreenshotRequest) (*ScreenshotResponse, error) {
	f.screenshotReq = req
	return &ScreenshotResponse{Screenshot: "ok"}, nil
}

func (f *fakeComputerUse) TakeRegionScreenshot(req *RegionScreenshotRequest) (*ScreenshotResponse, error) {
	f.regionReq = req
	return &ScreenshotResponse{Screenshot: "ok"}, nil
}

func (f *fakeComputerUse) TakeCompressedScreenshot(req *CompressedScreenshotRequest) (*ScreenshotResponse, error) {
	f.compressedReq = req
	return &ScreenshotResponse{Screenshot: "ok"}, nil
}

func (f *fakeComputerUse) TakeCompressedRegionScreenshot(req *CompressedRegionScreenshotRequest) (*ScreenshotResponse, error) {
	f.compressedRegionReq = req
	return &ScreenshotResponse{Screenshot: "ok"}, nil
}

func newScreenshotTestRouter(t *testing.T, fake *fakeComputerUse) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/computeruse/screenshot", WrapScreenshotHandler(fake.TakeScreenshot))
	r.GET("/computeruse/screenshot/region", WrapRegionScreenshotHandler(fake.TakeRegionScreenshot))
	r.GET("/computeruse/screenshot/compressed", WrapCompressedScreenshotHandler(fake.TakeCompressedScreenshot))
	r.GET("/computeruse/screenshot/region/compressed", WrapCompressedRegionScreenshotHandler(fake.TakeCompressedRegionScreenshot))
	return r
}

func performScreenshotRequest(router *gin.Engine, path string) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	router.ServeHTTP(rr, req)
	return rr
}

func TestWrapRegionScreenshotHandlerParsesLowercaseQueryParams(t *testing.T) {
	fake := &fakeComputerUse{}
	router := newScreenshotTestRouter(t, fake)

	rr := performScreenshotRequest(router, "/computeruse/screenshot/region?x=0&y=0&width=200&height=200")

	require.Equal(t, http.StatusOK, rr.Code)
	require.NotNil(t, fake.regionReq)
	assert.Equal(t, 0, fake.regionReq.X)
	assert.Equal(t, 0, fake.regionReq.Y)
	assert.Equal(t, 200, fake.regionReq.Width)
	assert.Equal(t, 200, fake.regionReq.Height)
}

func TestWrapCompressedRegionScreenshotHandlerParsesLowercaseQueryParams(t *testing.T) {
	fake := &fakeComputerUse{}
	router := newScreenshotTestRouter(t, fake)

	rr := performScreenshotRequest(router, "/computeruse/screenshot/region/compressed?x=0&y=0&width=200&height=200&format=png&quality=80&scale=1")

	require.Equal(t, http.StatusOK, rr.Code)
	require.NotNil(t, fake.compressedRegionReq)
	assert.Equal(t, 0, fake.compressedRegionReq.X)
	assert.Equal(t, 0, fake.compressedRegionReq.Y)
	assert.Equal(t, 200, fake.compressedRegionReq.Width)
	assert.Equal(t, 200, fake.compressedRegionReq.Height)
	assert.Equal(t, "png", fake.compressedRegionReq.Format)
	assert.Equal(t, 80, fake.compressedRegionReq.Quality)
	assert.Equal(t, 1.0, fake.compressedRegionReq.Scale)
}

func TestWrapRegionScreenshotHandlerRejectsMissingDimensions(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{
			name: "missing width",
			path: "/computeruse/screenshot/region?x=0&y=0&height=200",
		},
		{
			name: "missing height",
			path: "/computeruse/screenshot/region?x=0&y=0&width=200",
		},
		{
			name: "compressed missing width",
			path: "/computeruse/screenshot/region/compressed?x=0&y=0&height=200",
		},
		{
			name: "compressed missing height",
			path: "/computeruse/screenshot/region/compressed?x=0&y=0&width=200",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeComputerUse{}
			router := newScreenshotTestRouter(t, fake)

			rr := performScreenshotRequest(router, tt.path)

			assert.Equal(t, http.StatusBadRequest, rr.Code)
			assert.Nil(t, fake.regionReq)
			assert.Nil(t, fake.compressedRegionReq)
		})
	}
}

func TestWrapRegionScreenshotHandlerRejectsInvalidDimensions(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{
			name: "zero width",
			path: "/computeruse/screenshot/region?x=0&y=0&width=0&height=200",
		},
		{
			name: "zero height",
			path: "/computeruse/screenshot/region?x=0&y=0&width=200&height=0",
		},
		{
			name: "negative width",
			path: "/computeruse/screenshot/region?x=0&y=0&width=-1&height=200",
		},
		{
			name: "negative height",
			path: "/computeruse/screenshot/region?x=0&y=0&width=200&height=-1",
		},
		{
			name: "compressed zero width",
			path: "/computeruse/screenshot/region/compressed?x=0&y=0&width=0&height=200",
		},
		{
			name: "compressed zero height",
			path: "/computeruse/screenshot/region/compressed?x=0&y=0&width=200&height=0",
		},
		{
			name: "compressed negative width",
			path: "/computeruse/screenshot/region/compressed?x=0&y=0&width=-1&height=200",
		},
		{
			name: "compressed negative height",
			path: "/computeruse/screenshot/region/compressed?x=0&y=0&width=200&height=-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeComputerUse{}
			router := newScreenshotTestRouter(t, fake)

			rr := performScreenshotRequest(router, tt.path)

			assert.Equal(t, http.StatusBadRequest, rr.Code)
			assert.Nil(t, fake.regionReq)
			assert.Nil(t, fake.compressedRegionReq)
		})
	}
}

func TestWrapRegionScreenshotHandlerParsesShowCursor(t *testing.T) {
	fake := &fakeComputerUse{}
	router := newScreenshotTestRouter(t, fake)

	rr := performScreenshotRequest(router, "/computeruse/screenshot/region?x=0&y=0&width=200&height=200&showCursor=true")

	require.Equal(t, http.StatusOK, rr.Code)
	require.NotNil(t, fake.regionReq)
	assert.True(t, fake.regionReq.ShowCursor)
}

func TestWrapRegionScreenshotHandlerParsesShowCursorSnakeCaseAlias(t *testing.T) {
	fake := &fakeComputerUse{}
	router := newScreenshotTestRouter(t, fake)

	rr := performScreenshotRequest(router, "/computeruse/screenshot/region?x=0&y=0&width=200&height=200&show_cursor=true")

	require.Equal(t, http.StatusOK, rr.Code)
	require.NotNil(t, fake.regionReq)
	assert.True(t, fake.regionReq.ShowCursor)
}

func TestWrapScreenshotHandlerParsesShowCursor(t *testing.T) {
	fake := &fakeComputerUse{}
	router := newScreenshotTestRouter(t, fake)

	rr := performScreenshotRequest(router, "/computeruse/screenshot?showCursor=true")

	require.Equal(t, http.StatusOK, rr.Code)
	require.NotNil(t, fake.screenshotReq)
	assert.True(t, fake.screenshotReq.ShowCursor)
}

func TestWrapCompressedScreenshotHandlerParsesShowCursor(t *testing.T) {
	fake := &fakeComputerUse{}
	router := newScreenshotTestRouter(t, fake)

	rr := performScreenshotRequest(router, "/computeruse/screenshot/compressed?showCursor=true")

	require.Equal(t, http.StatusOK, rr.Code)
	require.NotNil(t, fake.compressedReq)
	assert.True(t, fake.compressedReq.ShowCursor)
}

func TestWrapCompressedRegionScreenshotHandlerParsesShowCursor(t *testing.T) {
	fake := &fakeComputerUse{}
	router := newScreenshotTestRouter(t, fake)

	rr := performScreenshotRequest(router, "/computeruse/screenshot/region/compressed?x=0&y=0&width=200&height=200&showCursor=true")

	require.Equal(t, http.StatusOK, rr.Code)
	require.NotNil(t, fake.compressedRegionReq)
	assert.True(t, fake.compressedRegionReq.ShowCursor)
}

func TestWrapScreenshotHandlerRejectsInvalidShowCursor(t *testing.T) {
	fake := &fakeComputerUse{}
	router := newScreenshotTestRouter(t, fake)

	rr := performScreenshotRequest(router, "/computeruse/screenshot?showCursor=yes")

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Nil(t, fake.screenshotReq)
}

func TestWrapScreenshotHandlerRejectsInvalidShowCursorSnakeCaseAlias(t *testing.T) {
	fake := &fakeComputerUse{}
	router := newScreenshotTestRouter(t, fake)

	rr := performScreenshotRequest(router, "/computeruse/screenshot?show_cursor=yes")

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Nil(t, fake.screenshotReq)
}

func TestWrapCompressedScreenshotHandlerRejectsInvalidShowCursor(t *testing.T) {
	fake := &fakeComputerUse{}
	router := newScreenshotTestRouter(t, fake)

	rr := performScreenshotRequest(router, "/computeruse/screenshot/compressed?showCursor=yes")

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Nil(t, fake.compressedReq)
}

func TestWrapRegionScreenshotHandlerAllowsNegativeCoordinates(t *testing.T) {
	fake := &fakeComputerUse{}
	router := newScreenshotTestRouter(t, fake)

	rr := performScreenshotRequest(router, "/computeruse/screenshot/region?x=-100&y=-100&width=100&height=100")

	require.Equal(t, http.StatusOK, rr.Code)
	require.NotNil(t, fake.regionReq)
	assert.Equal(t, -100, fake.regionReq.X)
	assert.Equal(t, -100, fake.regionReq.Y)
	assert.Equal(t, 100, fake.regionReq.Width)
	assert.Equal(t, 100, fake.regionReq.Height)
}

func TestWrapCompressedRegionScreenshotHandlerValidatesCompressionOptions(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{
			name: "quality non-integer",
			path: "/computeruse/screenshot/region/compressed?x=0&y=0&width=200&height=200&quality=abc",
		},
		{
			name: "quality too low",
			path: "/computeruse/screenshot/region/compressed?x=0&y=0&width=200&height=200&quality=0",
		},
		{
			name: "quality too high",
			path: "/computeruse/screenshot/region/compressed?x=0&y=0&width=200&height=200&quality=200",
		},
		{
			name: "scale non-float",
			path: "/computeruse/screenshot/region/compressed?x=0&y=0&width=200&height=200&scale=abc",
		},
		{
			name: "scale too low",
			path: "/computeruse/screenshot/region/compressed?x=0&y=0&width=200&height=200&scale=0",
		},
		{
			name: "scale too high",
			path: "/computeruse/screenshot/region/compressed?x=0&y=0&width=200&height=200&scale=2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeComputerUse{}
			router := newScreenshotTestRouter(t, fake)

			rr := performScreenshotRequest(router, tt.path)

			assert.Equal(t, http.StatusBadRequest, rr.Code)
			assert.Nil(t, fake.compressedRegionReq)
		})
	}
}

func TestWrapCompressedScreenshotHandlerValidatesCompressionOptions(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{
			name: "quality non-integer",
			path: "/computeruse/screenshot/compressed?quality=abc",
		},
		{
			name: "quality too low",
			path: "/computeruse/screenshot/compressed?quality=0",
		},
		{
			name: "quality too high",
			path: "/computeruse/screenshot/compressed?quality=200",
		},
		{
			name: "scale non-float",
			path: "/computeruse/screenshot/compressed?scale=abc",
		},
		{
			name: "scale too low",
			path: "/computeruse/screenshot/compressed?scale=0",
		},
		{
			name: "scale too high",
			path: "/computeruse/screenshot/compressed?scale=2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeComputerUse{}
			router := newScreenshotTestRouter(t, fake)

			rr := performScreenshotRequest(router, tt.path)

			assert.Equal(t, http.StatusBadRequest, rr.Code)
			assert.Nil(t, fake.compressedReq)
		})
	}
}

func TestWrapCompressedRegionScreenshotHandlerParsesValidCompressionOptions(t *testing.T) {
	fake := &fakeComputerUse{}
	router := newScreenshotTestRouter(t, fake)

	rr := performScreenshotRequest(router, "/computeruse/screenshot/region/compressed?x=0&y=0&width=200&height=200&quality=80&scale=0.5")

	require.Equal(t, http.StatusOK, rr.Code)
	require.NotNil(t, fake.compressedRegionReq)
	assert.Equal(t, "png", fake.compressedRegionReq.Format)
	assert.Equal(t, 80, fake.compressedRegionReq.Quality)
	assert.Equal(t, 0.5, fake.compressedRegionReq.Scale)
}

func TestWrapCompressedScreenshotHandlerParsesValidCompressionOptions(t *testing.T) {
	fake := &fakeComputerUse{}
	router := newScreenshotTestRouter(t, fake)

	rr := performScreenshotRequest(router, "/computeruse/screenshot/compressed?quality=80&scale=0.5")

	require.Equal(t, http.StatusOK, rr.Code)
	require.NotNil(t, fake.compressedReq)
	assert.Equal(t, "png", fake.compressedReq.Format)
	assert.Equal(t, 80, fake.compressedReq.Quality)
	assert.Equal(t, 0.5, fake.compressedReq.Scale)
}

func TestWrapRegionScreenshotHandlerRejectsTitleCaseQueryParams(t *testing.T) {
	fake := &fakeComputerUse{}
	router := newScreenshotTestRouter(t, fake)

	rr := performScreenshotRequest(router, "/computeruse/screenshot/region?X=0&Y=0&Width=200&Height=200")

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Nil(t, fake.regionReq)
}

func TestWrapCompressedRegionScreenshotHandlerRejectsTitleCaseQueryParams(t *testing.T) {
	fake := &fakeComputerUse{}
	router := newScreenshotTestRouter(t, fake)

	rr := performScreenshotRequest(router, "/computeruse/screenshot/region/compressed?X=0&Y=0&Width=200&Height=200")

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Nil(t, fake.compressedRegionReq)
}

func newComputerUseJSONContext(t *testing.T, path string, body any) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)

	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal test request: %v", err)
	}

	ctx.Request = httptest.NewRequest(http.MethodPost, path, bytes.NewReader(payload))
	ctx.Request.Header.Set("Content-Type", "application/json")

	return ctx, recorder
}

func TestWrapClickHandlerReturnsBadRequestForValidationErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := WrapClickHandler(func(req *MouseClickRequest) (*MouseClickResponse, error) {
		if req.Button == "wheel" {
			return nil, errors.New("unsupported mouse button")
		}

		return &MouseClickResponse{Position: Position{X: req.X, Y: req.Y}}, nil
	})

	ctx, recorder := newComputerUseJSONContext(t, "/computeruse/mouse/click", map[string]any{
		"x":      100,
		"y":      200,
		"button": "wheel",
		"double": false,
	})

	handler(ctx)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}

	if !strings.Contains(recorder.Body.String(), "unsupported mouse button") {
		t.Fatalf("expected validation error in response body, got %q", recorder.Body.String())
	}
}

func TestWrapScrollHandlerReturnsBadRequestForValidationErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := WrapScrollHandler(func(req *MouseScrollRequest) (*ScrollResponse, error) {
		if req.Direction == "left" {
			return nil, errors.New("unsupported scroll direction")
		}

		return &ScrollResponse{Success: true}, nil
	})

	ctx, recorder := newComputerUseJSONContext(t, "/computeruse/mouse/scroll", map[string]any{
		"x":         10,
		"y":         20,
		"direction": "left",
		"amount":    1,
	})

	handler(ctx)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}

	if !strings.Contains(recorder.Body.String(), "unsupported scroll direction") {
		t.Fatalf("expected validation error in response body, got %q", recorder.Body.String())
	}
}
