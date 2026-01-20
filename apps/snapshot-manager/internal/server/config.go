/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

package server

import (
	"fmt"
	"net/http"
	"os"

	"github.com/daytonaio/snapshot-manager/internal/config"
	"github.com/distribution/distribution/v3/configuration"
	"golang.org/x/crypto/bcrypt"
)

// BuildConfig creates a distribution registry configuration from the app config
func BuildConfig(cfg *config.Config) (*configuration.Configuration, error) {
	builder := NewConfigBuilder()

	builder.WithAddr(cfg.Addr)
	builder.WithLogLevel(cfg.LogLevel)

	switch cfg.StorageDriver {
	case "filesystem":
		builder.WithFilesystemStorage(cfg.StorageDir)
	case "s3":
		s3Params := make(map[string]interface{})
		if cfg.S3AccessKey != "" {
			s3Params["accesskey"] = cfg.S3AccessKey
		}
		if cfg.S3SecretKey != "" {
			s3Params["secretkey"] = cfg.S3SecretKey
		}
		if cfg.S3RegionEndpoint != "" {
			s3Params["regionendpoint"] = cfg.S3RegionEndpoint
		}
		if cfg.S3RootDirectory != "" {
			s3Params["rootdirectory"] = cfg.S3RootDirectory
		}
		s3Params["encrypt"] = cfg.S3Encrypt
		s3Params["secure"] = cfg.S3Secure
		builder.WithS3Storage(cfg.S3Region, cfg.S3Bucket, s3Params)
	}

	if cfg.StorageDeleteEnable {
		builder.WithDelete(true)
	}

	if cfg.CacheEnabled {
		builder.WithCache(cfg.CacheDriver)
	}

	if cfg.TLSCertificate != "" {
		builder.WithTLS(cfg.TLSCertificate, cfg.TLSKey)
	}

	// Configure authentication based on AuthType
	switch cfg.AuthType {
	case config.AuthTypeBasic:
		htpasswdPath, err := generateHtpasswdFile(cfg.AuthUsername, cfg.AuthPassword)
		if err != nil {
			return nil, fmt.Errorf("failed to generate htpasswd file: %w", err)
		}
		builder.WithAuth(htpasswdPath)
	case config.AuthTypeNone:
		// Check for legacy htpasswd path configuration
		if cfg.AuthHtpasswdPath != "" {
			builder.WithAuth(cfg.AuthHtpasswdPath)
		}
	}

	if cfg.HTTPSecret != "" {
		builder.WithHTTPSecret(cfg.HTTPSecret)
	}

	return builder.Build()
}

// generateHtpasswdFile creates a temporary htpasswd file with the given credentials
func generateHtpasswdFile(username, password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	htpasswdContent := fmt.Sprintf("%s:%s\n", username, string(hashedPassword))

	tmpFile, err := os.CreateTemp("", "htpasswd-*")
	if err != nil {
		return "", fmt.Errorf("failed to create htpasswd file: %w", err)
	}
	defer tmpFile.Close()

	if _, err := tmpFile.WriteString(htpasswdContent); err != nil {
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("failed to write htpasswd file: %w", err)
	}

	return tmpFile.Name(), nil
}

// ConfigBuilder provides a fluent interface for building registry configurations
type ConfigBuilder struct {
	config *configuration.Configuration
}

// NewConfigBuilder creates a new configuration builder with sensible defaults
func NewConfigBuilder() *ConfigBuilder {
	return &ConfigBuilder{
		config: &configuration.Configuration{
			Version: "0.1",
			Log: configuration.Log{
				Level:     configuration.Loglevel("info"),
				Formatter: "text",
			},
			HTTP: configuration.HTTP{
				Addr: ":5000",
				Headers: http.Header{
					"X-Content-Type-Options": []string{"nosniff"},
				},
			},
			Storage: make(configuration.Storage),
			Catalog: configuration.Catalog{
				MaxEntries: 1000, // Enable catalog with default max entries
			},
		},
	}
}

// WithFilesystemStorage configures filesystem storage
func (cb *ConfigBuilder) WithFilesystemStorage(rootDir string) *ConfigBuilder {
	cb.config.Storage["filesystem"] = configuration.Parameters{
		"rootdirectory": rootDir,
	}
	return cb
}

// WithS3Storage configures S3 storage
func (cb *ConfigBuilder) WithS3Storage(region, bucket string, params map[string]interface{}) *ConfigBuilder {
	s3Params := configuration.Parameters{
		"region": region,
		"bucket": bucket,
	}
	// Merge additional params
	for k, v := range params {
		s3Params[k] = v
	}
	cb.config.Storage["s3"] = s3Params
	return cb
}

// WithAddr sets the HTTP address
func (cb *ConfigBuilder) WithAddr(addr string) *ConfigBuilder {
	cb.config.HTTP.Addr = addr
	return cb
}

// WithLogLevel sets the log level
func (cb *ConfigBuilder) WithLogLevel(level string) *ConfigBuilder {
	cb.config.Log.Level = configuration.Loglevel(level)
	return cb
}

// WithTLS enables TLS with the given certificate and key files
func (cb *ConfigBuilder) WithTLS(certFile, keyFile string) *ConfigBuilder {
	cb.config.HTTP.TLS = configuration.TLS{
		Certificate: certFile,
		Key:         keyFile,
	}
	return cb
}

// WithAuth enables htpasswd authentication
func (cb *ConfigBuilder) WithAuth(htpasswdPath string) *ConfigBuilder {
	cb.config.Auth = configuration.Auth{
		"htpasswd": configuration.Parameters{
			"realm": "Registry Realm",
			"path":  htpasswdPath,
		},
	}
	return cb
}

// WithDelete enables or disables blob deletion
func (cb *ConfigBuilder) WithDelete(enabled bool) *ConfigBuilder {
	cb.config.Storage["delete"] = configuration.Parameters{
		"enabled": enabled,
	}
	return cb
}

// WithCache enables in-memory blob descriptor caching
func (cb *ConfigBuilder) WithCache(driver string) *ConfigBuilder {
	cb.config.Storage["cache"] = configuration.Parameters{
		"blobdescriptor": driver,
	}
	return cb
}

// WithCatalog configures the catalog endpoint with a maximum number of entries
func (cb *ConfigBuilder) WithCatalog(maxEntries int) *ConfigBuilder {
	cb.config.Catalog = configuration.Catalog{
		MaxEntries: maxEntries,
	}
	return cb
}

// WithHTTPSecret sets a shared secret for multi-instance deployments
// This is REQUIRED when running multiple registry instances behind a load balancer
func (cb *ConfigBuilder) WithHTTPSecret(secret string) *ConfigBuilder {
	cb.config.HTTP.Secret = secret
	return cb
}

// WithRawConfig allows direct access to the underlying configuration
// This is useful for advanced configurations not covered by the builder
func (cb *ConfigBuilder) WithRawConfig(fn func(*configuration.Configuration)) *ConfigBuilder {
	fn(cb.config)
	return cb
}

// Build returns the built configuration
func (cb *ConfigBuilder) Build() (*configuration.Configuration, error) {
	// Validate that at least one storage driver is configured
	hasDriver := false
	for key := range cb.config.Storage {
		if key != "delete" && key != "maintenance" && key != "cache" {
			hasDriver = true
			break
		}
	}
	if !hasDriver {
		return nil, fmt.Errorf("no storage driver configured")
	}
	return cb.config, nil
}

// Config returns the raw configuration (for advanced users who want to modify it directly)
func (cb *ConfigBuilder) Config() *configuration.Configuration {
	return cb.config
}
