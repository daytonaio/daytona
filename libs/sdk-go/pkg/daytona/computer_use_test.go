// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package daytona

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/daytonaio/daytona/libs/sdk-go/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComputerUseServiceCreation(t *testing.T) {
	cu := NewComputerUseService(nil, nil)
	require.NotNil(t, cu)
}

func TestComputerUseLazyInit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer server.Close()

	client := createTestToolboxClient(server)
	cu := NewComputerUseService(client, nil)

	assert.NotNil(t, cu.Mouse())
	assert.NotNil(t, cu.Keyboard())
	assert.NotNil(t, cu.Screenshot())
	assert.NotNil(t, cu.Display())
	assert.NotNil(t, cu.Recording())

	assert.Same(t, cu.Mouse(), cu.Mouse())
	assert.Same(t, cu.Keyboard(), cu.Keyboard())
	assert.Same(t, cu.Screenshot(), cu.Screenshot())
	assert.Same(t, cu.Display(), cu.Display())
	assert.Same(t, cu.Recording(), cu.Recording())
}

func TestComputerUseStart(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{})
	}))
	defer server.Close()

	client := createTestToolboxClient(server)
	cu := NewComputerUseService(client, nil)

	ctx := context.Background()
	err := cu.Start(ctx)
	assert.NoError(t, err)
}

func TestComputerUseStop(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{})
	}))
	defer server.Close()

	client := createTestToolboxClient(server)
	cu := NewComputerUseService(client, nil)

	ctx := context.Background()
	err := cu.Stop(ctx)
	assert.NoError(t, err)
}

func TestComputerUseGetStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": "running"})
	}))
	defer server.Close()

	client := createTestToolboxClient(server)
	cu := NewComputerUseService(client, nil)

	ctx := context.Background()
	status, err := cu.GetStatus(ctx)
	require.NoError(t, err)
	assert.Equal(t, "running", status["status"])
}

func TestMouseGetPosition(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"x": float64(100), "y": float64(200)})
	}))
	defer server.Close()

	client := createTestToolboxClient(server)
	mouse := NewMouseService(client, nil)

	ctx := context.Background()
	pos, err := mouse.GetPosition(ctx)
	require.NoError(t, err)
	assert.NotNil(t, pos["x"])
	assert.NotNil(t, pos["y"])
}

func TestMouseMove(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"x": float64(500), "y": float64(300)})
	}))
	defer server.Close()

	client := createTestToolboxClient(server)
	mouse := NewMouseService(client, nil)

	ctx := context.Background()
	pos, err := mouse.Move(ctx, 500, 300)
	require.NoError(t, err)
	assert.NotNil(t, pos)
}

func TestMouseClick(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"x": float64(100), "y": float64(200)})
	}))
	defer server.Close()

	client := createTestToolboxClient(server)
	mouse := NewMouseService(client, nil)

	ctx := context.Background()
	pos, err := mouse.Click(ctx, 100, 200, nil, nil)
	require.NoError(t, err)
	assert.NotNil(t, pos)
}

func TestMouseClickWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"x": float64(100), "y": float64(200)})
	}))
	defer server.Close()

	client := createTestToolboxClient(server)
	mouse := NewMouseService(client, nil)

	button := "right"
	doubleClick := true
	ctx := context.Background()
	pos, err := mouse.Click(ctx, 100, 200, &button, &doubleClick)
	require.NoError(t, err)
	assert.NotNil(t, pos)
}

func TestMouseDrag(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"x": float64(300), "y": float64(300)})
	}))
	defer server.Close()

	client := createTestToolboxClient(server)
	mouse := NewMouseService(client, nil)

	ctx := context.Background()
	pos, err := mouse.Drag(ctx, 100, 100, 300, 300, nil)
	require.NoError(t, err)
	assert.NotNil(t, pos)
}

func TestMouseScroll(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
	}))
	defer server.Close()

	client := createTestToolboxClient(server)
	mouse := NewMouseService(client, nil)

	ctx := context.Background()
	success, err := mouse.Scroll(ctx, 500, 400, "down", nil)
	require.NoError(t, err)
	assert.True(t, success)
}

func TestMouseScrollWithAmount(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
	}))
	defer server.Close()

	client := createTestToolboxClient(server)
	mouse := NewMouseService(client, nil)

	amount := 5
	ctx := context.Background()
	success, err := mouse.Scroll(ctx, 500, 400, "up", &amount)
	require.NoError(t, err)
	assert.True(t, success)
}

func TestKeyboardType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{})
	}))
	defer server.Close()

	client := createTestToolboxClient(server)
	kb := NewKeyboardService(client, nil)

	ctx := context.Background()
	err := kb.Type(ctx, "Hello, World!", nil)
	assert.NoError(t, err)
}

func TestKeyboardTypeWithDelay(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{})
	}))
	defer server.Close()

	client := createTestToolboxClient(server)
	kb := NewKeyboardService(client, nil)

	delay := 50
	ctx := context.Background()
	err := kb.Type(ctx, "slow", &delay)
	assert.NoError(t, err)
}

func TestKeyboardPress(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{})
	}))
	defer server.Close()

	client := createTestToolboxClient(server)
	kb := NewKeyboardService(client, nil)

	ctx := context.Background()
	err := kb.Press(ctx, "Enter", nil)
	assert.NoError(t, err)
}

func TestKeyboardPressWithModifiers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{})
	}))
	defer server.Close()

	client := createTestToolboxClient(server)
	kb := NewKeyboardService(client, nil)

	ctx := context.Background()
	err := kb.Press(ctx, "s", []string{"ctrl"})
	assert.NoError(t, err)
}

func TestKeyboardHotkey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{})
	}))
	defer server.Close()

	client := createTestToolboxClient(server)
	kb := NewKeyboardService(client, nil)

	ctx := context.Background()
	err := kb.Hotkey(ctx, "ctrl+c")
	assert.NoError(t, err)
}

func TestDisplayGetInfo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"displays": []map[string]interface{}{
				{"width": 1920, "height": 1080},
			},
		})
	}))
	defer server.Close()

	client := createTestToolboxClient(server)
	display := NewDisplayService(client, nil)

	ctx := context.Background()
	info, err := display.GetInfo(ctx)
	require.NoError(t, err)
	assert.NotNil(t, info["displays"])
}

func TestDisplayGetWindowsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "desktop not running"})
	}))
	defer server.Close()

	client := createTestToolboxClient(server)
	display := NewDisplayService(client, nil)

	ctx := context.Background()
	_, err := display.GetWindows(ctx)
	require.Error(t, err)
}

func TestConvertInt32PtrToIntPtr(t *testing.T) {
	t.Run("nil input", func(t *testing.T) {
		assert.Nil(t, convertInt32PtrToIntPtr(nil))
	})

	t.Run("non-nil input", func(t *testing.T) {
		val := int32(42)
		result := convertInt32PtrToIntPtr(&val)
		require.NotNil(t, result)
		assert.Equal(t, 42, *result)
	})
}

func TestComputerUseErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "desktop not running"})
	}))
	defer server.Close()

	client := createTestToolboxClient(server)
	cu := NewComputerUseService(client, nil)

	ctx := context.Background()
	err := cu.Start(ctx)
	require.Error(t, err)
}

func TestScreenshotAndDisplayServices(t *testing.T) {
	t.Run("full screen and region screenshots map response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.URL.Path, "region") {
				writeJSONResponse(t, w, http.StatusOK, map[string]any{"screenshot": "region-image", "sizeBytes": 9})
				return
			}
			writeJSONResponse(t, w, http.StatusOK, map[string]any{"screenshot": "full-image", "sizeBytes": 12})
		}))
		defer server.Close()

		ss := NewScreenshotService(createTestToolboxClient(server), nil)
		full, err := ss.TakeFullScreen(context.Background(), boolPtr(true))
		require.NoError(t, err)
		assert.Equal(t, "full-image", full.Image)
		require.NotNil(t, full.SizeBytes)
		region, err := ss.TakeRegion(context.Background(), types.ScreenshotRegion{X: 1, Y: 2, Width: 3, Height: 4}, nil)
		require.NoError(t, err)
		assert.Equal(t, 3, region.Width)
		assert.Equal(t, "region-image", region.Image)
	})

	t.Run("display windows success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			writeJSONResponse(t, w, http.StatusOK, map[string]any{"windows": []map[string]any{{"title": "Editor"}}})
		}))
		defer server.Close()

		display := NewDisplayService(createTestToolboxClient(server), nil)
		windows, err := display.GetWindows(context.Background())
		require.NoError(t, err)
		assert.NotEmpty(t, windows["windows"])
	})
}

func TestRecordingServiceOperations(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/download"):
			_, _ = w.Write([]byte("video-data"))
		case r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusOK)
		default:
			writeJSONResponse(t, w, http.StatusOK, map[string]any{"id": "rec-1", "fileName": "video.mp4", "status": "completed", "filePath": "/tmp/video.mp4", "startTime": "2025-01-01T00:00:00Z"})
		}
	}))
	defer server.Close()

	recording := NewRecordingService(createTestToolboxClient(server))
	ctx := context.Background()
	localPath := filepath.Join(t.TempDir(), "nested", "video.mp4")
	require.NoError(t, recording.Download(ctx, "rec-1", localPath))
	data, err := os.ReadFile(localPath)
	require.NoError(t, err)
	assert.Equal(t, "video-data", string(data))
	require.NoError(t, recording.Delete(ctx, "rec-1"))
}
