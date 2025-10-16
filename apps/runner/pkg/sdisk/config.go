package sdisk

// Config contains all configuration for the SDisk volume manager
type Config struct {
	// Base directory for storing volumes
	DataDir string

	// S3 configuration
	S3 S3Config

	// QCOW2 configuration
	QCOW2 QCOW2Config

	// Pool configuration
	Pool PoolConfig
}

// S3Config contains S3-specific configuration
type S3Config struct {
	// S3 bucket name for storing volumes
	Bucket string

	// AWS region
	Region string

	// AWS access key ID (optional if using IAM roles)
	AccessKeyID string

	// AWS secret access key (optional if using IAM roles)
	SecretAccessKey string

	// Custom endpoint for S3-compatible services (optional)
	Endpoint string

	// Use path-style addressing for S3-compatible services
	UsePathStyle bool

	// Layer size threshold in bytes (default: 100MB)
	// If the last layer is smaller than this, it will be reused for the next push
	// If >= this size, a new layer will be created
	LayerSizeThresholdMB int64
}

// QCOW2Config contains QCOW2-specific configuration
type QCOW2Config struct {
	// Compression type for QCOW2 images: "zlib", "zstd", or "" for none
	Compression string

	// Cluster size for QCOW2 in bytes (default: 65536 = 64K)
	ClusterSize int

	// Lazy refcounts for better performance
	LazyRefcounts bool

	// Preallocation mode: "off", "metadata", "falloc", "full"
	// - "off": Thin provisioning - file only uses space for actual data (default)
	// - "metadata": Pre-allocate metadata structures (~1-2% of virtual size)
	// - "falloc": Pre-allocate full space using fallocate (fast)
	// - "full": Pre-allocate and zero full space (slow but secure)
	Preallocation string
}

// PoolConfig contains volume pool configuration
type PoolConfig struct {
	// Enable volume pooling (default: false for backward compatibility)
	Enabled bool

	// Maximum number of concurrently mounted volumes (default: 100)
	// When this limit is reached, the least recently used volume will be unmounted
	MaxMounted int
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.DataDir == "" {
		return ErrInvalidConfig
	}

	// S3 validation
	if c.S3.Bucket != "" && c.S3.Region == "" {
		return ErrInvalidConfig
	}

	// Set QCOW2 defaults
	if c.QCOW2.ClusterSize == 0 {
		c.QCOW2.ClusterSize = 65536 // 64K default
	}

	if c.QCOW2.Preallocation == "" {
		c.QCOW2.Preallocation = "off"
	}

	// Set Pool defaults
	if c.Pool.Enabled {
		if c.Pool.MaxMounted <= 0 {
			c.Pool.MaxMounted = 100 // Default to 100 mounted volumes
		}
	}

	// Set S3 layer size threshold default
	if c.S3.LayerSizeThresholdMB <= 0 {
		c.S3.LayerSizeThresholdMB = 100 // Default to 100MB
	}

	// Validate compression type
	if c.QCOW2.Compression != "" &&
		c.QCOW2.Compression != "zlib" &&
		c.QCOW2.Compression != "zstd" {
		return ErrInvalidConfig
	}

	// Validate preallocation mode
	if c.QCOW2.Preallocation != "off" &&
		c.QCOW2.Preallocation != "metadata" &&
		c.QCOW2.Preallocation != "falloc" &&
		c.QCOW2.Preallocation != "full" {
		return ErrInvalidConfig
	}

	return nil
}
