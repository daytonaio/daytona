package multiplexer

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/daytonaio/runner/pkg/volume"
)

// MockProvider implements a mock volume provider for testing
type MockProvider struct {
	files map[string][]byte
}

func NewMockProvider() *MockProvider {
	return &MockProvider{
		files: make(map[string][]byte),
	}
}

func (m *MockProvider) Connect(ctx context.Context, config volume.ProviderConfig) error {
	return nil
}

func (m *MockProvider) ReadFile(ctx context.Context, path string, offset int64, size int) ([]byte, error) {
	data, exists := m.files[path]
	if !exists {
		return nil, nil
	}

	end := offset + int64(size)
	if end > int64(len(data)) {
		end = int64(len(data))
	}

	return data[offset:end], nil
}

func (m *MockProvider) WriteFile(ctx context.Context, path string, data []byte, offset int64) error {
	if offset == 0 {
		m.files[path] = data
	} else {
		// Simple append for testing
		existing := m.files[path]
		m.files[path] = append(existing, data...)
	}
	return nil
}

func (m *MockProvider) DeleteFile(ctx context.Context, path string) error {
	delete(m.files, path)
	return nil
}

func (m *MockProvider) ListDir(ctx context.Context, path string) ([]volume.FileInfo, error) {
	return []volume.FileInfo{}, nil
}

func (m *MockProvider) CreateDir(ctx context.Context, path string) error {
	return nil
}

func (m *MockProvider) DeleteDir(ctx context.Context, path string) error {
	return nil
}

func (m *MockProvider) GetFileInfo(ctx context.Context, path string) (volume.FileInfo, error) {
	data, exists := m.files[path]
	if !exists {
		return volume.FileInfo{}, nil
	}

	return volume.FileInfo{
		Name:    path,
		Size:    int64(len(data)),
		Mode:    0666,
		ModTime: time.Now(),
	}, nil
}

func (m *MockProvider) Exists(ctx context.Context, path string) (bool, error) {
	_, exists := m.files[path]
	return exists, nil
}

func (m *MockProvider) Rename(ctx context.Context, oldPath, newPath string) error {
	if data, exists := m.files[oldPath]; exists {
		m.files[newPath] = data
		delete(m.files, oldPath)
	}
	return nil
}

func (m *MockProvider) Truncate(ctx context.Context, path string, size int64) error {
	if data, exists := m.files[path]; exists {
		if int64(len(data)) > size {
			m.files[path] = data[:size]
		}
	}
	return nil
}

func (m *MockProvider) Close() error {
	return nil
}

func TestMultiplexerDaemon(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()

	// Create daemon
	daemon := NewMultiplexerDaemon("/tmp/test-mount", "/tmp/test-cache", logger)

	// Test volume registration
	t.Run("RegisterVolume", func(t *testing.T) {
		config := volume.ProviderConfig{
			Type:       "mock",
			BucketName: "test-bucket",
		}

		err := daemon.RegisterVolume(ctx, "test-volume", config, false)
		if err == nil {
			t.Error("Expected error for unregistered provider type, got nil")
		}

		// TODO: Register mock provider factory and test successful registration
	})

	// Test reference counting
	t.Run("RefCounting", func(t *testing.T) {
		// First need to register a volume
		// Then test increment/decrement
	})

	// Test statistics
	t.Run("Statistics", func(t *testing.T) {
		stats := daemon.GetStats()
		if stats == nil {
			t.Error("Expected non-nil stats")
		}

		if stats.TotalVolumes != 0 {
			t.Errorf("Expected 0 volumes, got %d", stats.TotalVolumes)
		}
	})
}

func TestVolumeCache(t *testing.T) {
	cache := NewVolumeCache("/tmp/test-cache", 1024*1024) // 1MB cache

	t.Run("PutAndGet", func(t *testing.T) {
		testData := []byte("test data")
		cache.Put("test.txt", 0, testData)

		// Should be able to retrieve immediately
		data, ok := cache.Get("test.txt", 0, len(testData))
		if !ok {
			t.Error("Expected cache hit")
		}

		if string(data) != string(testData) {
			t.Errorf("Expected %s, got %s", testData, data)
		}
	})

	t.Run("Invalidate", func(t *testing.T) {
		testData := []byte("test data")
		cache.Put("test2.txt", 0, testData)

		// Invalidate the cache
		cache.Invalidate("test2.txt")

		// Should not be able to retrieve
		_, ok := cache.Get("test2.txt", 0, len(testData))
		if ok {
			t.Error("Expected cache miss after invalidation")
		}
	})

	t.Run("Clear", func(t *testing.T) {
		cache.Put("test3.txt", 0, []byte("data"))

		err := cache.Clear()
		if err != nil {
			t.Errorf("Clear failed: %v", err)
		}

		// Should not be able to retrieve
		_, ok := cache.Get("test3.txt", 0, 4)
		if ok {
			t.Error("Expected cache miss after clear")
		}
	})
}

func BenchmarkCache(b *testing.B) {
	cache := NewVolumeCache("/tmp/bench-cache", 10*1024*1024) // 10MB cache
	testData := make([]byte, 4096)                            // 4KB blocks

	b.Run("Put", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			cache.Put("bench.txt", int64(i*4096), testData)
		}
	})

	b.Run("Get", func(b *testing.B) {
		// Pre-populate cache
		cache.Put("bench-read.txt", 0, testData)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cache.Get("bench-read.txt", 0, 4096)
		}
	})
}
