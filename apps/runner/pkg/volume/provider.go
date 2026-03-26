package volume

import (
	"context"
	"io/fs"
	"time"
)

// Provider defines the interface for volume storage backends
type Provider interface {
	// Initialize connection to storage backend
	Connect(ctx context.Context, config ProviderConfig) error

	// File operations
	ReadFile(ctx context.Context, path string, offset int64, size int) ([]byte, error)
	WriteFile(ctx context.Context, path string, data []byte, offset int64) error
	DeleteFile(ctx context.Context, path string) error

	// Directory operations
	ListDir(ctx context.Context, path string) ([]FileInfo, error)
	CreateDir(ctx context.Context, path string) error
	DeleteDir(ctx context.Context, path string) error

	// Metadata operations
	GetFileInfo(ctx context.Context, path string) (FileInfo, error)
	Exists(ctx context.Context, path string) (bool, error)

	// Advanced operations
	Rename(ctx context.Context, oldPath, newPath string) error
	Truncate(ctx context.Context, path string, size int64) error

	// Cleanup
	Close() error
}

// ProviderConfig holds configuration for storage providers
type ProviderConfig struct {
	Type       string            // "s3", "gcs", "azure"
	Endpoint   string            // Storage endpoint URL
	AccessKey  string            // Access credentials
	SecretKey  string            // Secret credentials
	Region     string            // Region for cloud providers
	BucketName string            // Bucket/container name
	Subpath    string            // Optional subpath within bucket
	Options    map[string]string // Provider-specific options
}

// FileInfo represents file or directory metadata
type FileInfo struct {
	Name    string      // Base name of the file
	Size    int64       // Length in bytes for regular files
	Mode    fs.FileMode // File mode bits
	ModTime time.Time   // Modification time
	IsDir   bool        // True if directory
	ETag    string      // Entity tag for caching
}

// WriteHandle represents an open file for writing
type WriteHandle interface {
	Write(ctx context.Context, data []byte, offset int64) error
	Flush(ctx context.Context) error
	Close(ctx context.Context) error
}

// ProviderFactory creates providers based on type
type ProviderFactory interface {
	CreateProvider(providerType string) (Provider, error)
}
