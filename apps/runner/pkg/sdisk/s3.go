package sdisk

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// S3Client handles S3 operations for disks
type S3Client struct {
	s3Client *s3.Client
	bucket   string
}

// S3Metadata represents disk metadata stored in S3
type S3Metadata struct {
	Name     string    `json:"name"`
	SizeGB   int64     `json:"size_gb"`
	Created  time.Time `json:"created"`
	Modified time.Time `json:"modified"`
	Checksum string    `json:"checksum"`

	// Layer/snapshot information for incremental uploads
	Layers      []S3LayerInfo `json:"layers,omitempty"`        // List of all layers in order (base to top)
	BaseLayerID string        `json:"base_layer_id,omitempty"` // ID of the base layer
	TopLayerID  string        `json:"top_layer_id,omitempty"`  // ID of the current top layer
}

// LayerInfo represents information about a single QCOW2 layer
type S3LayerInfo struct {
	ID          string    `json:"id"`                    // Unique layer ID (e.g., timestamp-based)
	ParentID    string    `json:"parent_id,omitempty"`   // ID of parent layer (empty for base)
	Created     time.Time `json:"created"`               // When this layer was created
	Size        int64     `json:"size"`                  // Actual size of this layer file
	Checksum    string    `json:"checksum"`              // SHA256 of this layer
	Description string    `json:"description,omitempty"` // Optional description
}

// NewClient creates a new S3 client
func NewS3Client(ctx context.Context, cfg S3Config) (*S3Client, error) {
	if cfg.Bucket == "" {
		return nil, fmt.Errorf("bucket name is required")
	}

	if cfg.Region == "" {
		return nil, fmt.Errorf("region is required")
	}

	// Build AWS config
	var opts []func(*config.LoadOptions) error

	opts = append(opts, config.WithRegion(cfg.Region))

	// Add credentials if provided
	if cfg.AccessKeyID != "" && cfg.SecretAccessKey != "" {
		opts = append(opts, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				cfg.AccessKeyID,
				cfg.SecretAccessKey,
				"",
			),
		))
	}

	awsConfig, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client
	s3Opts := []func(*s3.Options){}

	if cfg.Endpoint != "" {
		s3Opts = append(s3Opts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
			o.UsePathStyle = cfg.UsePathStyle
		})
	}

	s3Client := s3.NewFromConfig(awsConfig, s3Opts...)

	return &S3Client{
		s3Client: s3Client,
		bucket:   cfg.Bucket,
	}, nil
}

// diskKey returns the S3 key for a disk's QCOW2 file (legacy/non-layered)
func (c *S3Client) diskKey(diskName string) string {
	return fmt.Sprintf("disks/%s/disk.qcow2", diskName)
}

// layerKey returns the S3 key for a specific layer
func (c *S3Client) layerKey(diskName, layerID string) string {
	return fmt.Sprintf("disks/%s/layers/%s", diskName, layerID)
}

// metadataKey returns the S3 key for a disk's metadata
func (c *S3Client) metadataKey(diskName string) string {
	return fmt.Sprintf("disks/%s/metadata.json", diskName)
}

// UploadDisk uploads a disk to S3
func (c *S3Client) UploadDisk(ctx context.Context, diskName, localPath string, metadata S3Metadata) error {
	// Open the local file
	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %w", err)
	}
	defer file.Close()

	// Get file size
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	// Upload the disk file
	key := c.diskKey(diskName)
	_, err = c.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(c.bucket),
		Key:           aws.String(key),
		Body:          file,
		ContentLength: aws.Int64(fileInfo.Size()),
		ContentType:   aws.String("application/octet-stream"),
	})
	if err != nil {
		return fmt.Errorf("failed to upload disk: %w", err)
	}

	// Upload metadata
	if err := c.UploadMetadata(ctx, diskName, metadata); err != nil {
		return fmt.Errorf("failed to upload metadata: %w", err)
	}

	return nil
}

// DownloadDisk downloads a disk from S3
func (c *S3Client) DownloadDisk(ctx context.Context, diskName, localPath string) error {
	// Create directory for the disk if it doesn't exist
	dir := filepath.Dir(localPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Download the disk file
	key := c.diskKey(diskName)
	result, err := c.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to download disk: %w", err)
	}
	defer result.Body.Close()

	// Create local file
	file, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create local file: %w", err)
	}
	defer file.Close()

	// Copy data
	if _, err := io.Copy(file, result.Body); err != nil {
		return fmt.Errorf("failed to write disk data: %w", err)
	}

	return nil
}

// UploadMetadata uploads disk metadata to S3
func (c *S3Client) UploadMetadata(ctx context.Context, diskName string, metadata S3Metadata) error {
	// Marshal metadata to JSON
	data, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Upload metadata
	key := c.metadataKey(diskName)
	_, err = c.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String("application/json"),
	})
	if err != nil {
		return fmt.Errorf("failed to upload metadata: %w", err)
	}

	return nil
}

// DownloadMetadata downloads disk metadata from S3
func (c *S3Client) DownloadMetadata(ctx context.Context, diskName string) (*S3Metadata, error) {
	key := c.metadataKey(diskName)
	result, err := c.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download metadata: %w", err)
	}
	defer result.Body.Close()

	// Read and unmarshal metadata
	data, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata: %w", err)
	}

	var metadata S3Metadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &metadata, nil
}

// DiskExists checks if a disk exists in S3
func (c *S3Client) DiskExists(ctx context.Context, diskName string) (bool, error) {
	key := c.diskKey(diskName)
	_, err := c.s3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		// Check if it's a "not found" error
		var notFound *types.NotFound
		if ok := errors.As(err, &notFound); ok {
			return false, nil
		}
		return false, fmt.Errorf("failed to check disk existence: %w", err)
	}

	return true, nil
}

// DeleteDisk deletes a disk and all its layers from S3
func (c *S3Client) DeleteDisk(ctx context.Context, diskName string) error {
	// List all objects under this disk's prefix
	prefix := fmt.Sprintf("disks/%s/", diskName)
	result, err := c.s3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(c.bucket),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		return fmt.Errorf("failed to list disk objects: %w", err)
	}

	// Delete all objects (layers, metadata, etc.)
	for _, obj := range result.Contents {
		_, err := c.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(c.bucket),
			Key:    obj.Key,
		})
		if err != nil {
			return fmt.Errorf("failed to delete %s: %w", *obj.Key, err)
		}
	}

	return nil
}

// ListDisks lists all disks in S3
func (c *S3Client) ListDisks(ctx context.Context) ([]string, error) {
	prefix := "disks/"
	result, err := c.s3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket:    aws.String(c.bucket),
		Prefix:    aws.String(prefix),
		Delimiter: aws.String("/"),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list disks: %w", err)
	}

	var disks []string
	for _, commonPrefix := range result.CommonPrefixes {
		if commonPrefix.Prefix != nil {
			// Extract disk name from prefix: "disks/name/" -> "name"
			parts := strings.Split(strings.TrimSuffix(*commonPrefix.Prefix, "/"), "/")
			if len(parts) >= 2 {
				disks = append(disks, parts[len(parts)-1])
			}
		}
	}

	return disks, nil
}

// UploadLayer uploads a single QCOW2 layer to S3 using sparse-aware compression
func (c *S3Client) UploadLayer(ctx context.Context, diskName, layerID, localPath string) error {
	// Verify the source file exists
	if _, err := os.Stat(localPath); err != nil {
		return fmt.Errorf("source file does not exist: %w", err)
	}

	// Create a temporary file for the tar output
	tmpFile, err := os.CreateTemp("", "layer-*.tar.gz")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Create tar.gz with sparse support
	dir := filepath.Dir(localPath)
	base := filepath.Base(localPath)

	// Check if tar is available
	if _, err := exec.LookPath("tar"); err != nil {
		return fmt.Errorf("tar command not found: %w", err)
	}

	// Add a timeout to prevent hanging
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(timeoutCtx, "tar", "-czS", "-f", tmpFile.Name(), "-C", dir, base)

	// Capture stderr for debugging
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create tar archive: %w, stderr: %s", err, stderr.String())
	}

	// Verify the tar file was created and has content
	fileInfo, err := tmpFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat tar file: %w", err)
	}
	if fileInfo.Size() == 0 {
		return fmt.Errorf("tar file is empty")
	}

	// Read the tar file into memory for S3 upload
	tarData, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		return fmt.Errorf("failed to read tar file: %w", err)
	}

	// Upload the tar file to S3
	key := c.layerKey(diskName, layerID) + ".tar.gz"
	_, err = c.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(c.bucket),
		Key:           aws.String(key),
		Body:          bytes.NewReader(tarData),
		ContentLength: aws.Int64(int64(len(tarData))),
		ContentType:   aws.String("application/x-tar+gzip"),
	})
	if err != nil {
		return fmt.Errorf("failed to upload layer: %w", err)
	}

	return nil
}

// DownloadLayer downloads a single QCOW2 layer from S3 using sparse-aware extraction
func (c *S3Client) DownloadLayer(ctx context.Context, diskName, layerID, localPath string) error {
	// Create directory for the layer if it doesn't exist
	dir := filepath.Dir(localPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Download the layer file
	key := c.layerKey(diskName, layerID) + ".tar.gz"
	result, err := c.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to download layer: %w", err)
	}
	defer result.Body.Close()

	// Extract with sparse support
	cmd := exec.CommandContext(ctx, "tar", "-xzS", "-f", "-", "-C", dir)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create tar pipe: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start tar: %w", err)
	}

	// Stream S3 data to tar
	if _, err := io.Copy(stdin, result.Body); err != nil {
		stdin.Close()
		cmd.Process.Kill()
		return fmt.Errorf("failed to extract layer: %w", err)
	}
	stdin.Close()

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("failed to complete extraction: %w", err)
	}

	return nil
}

// LayerExists checks if a specific layer exists in S3
func (c *S3Client) LayerExists(ctx context.Context, diskName, layerID string) (bool, error) {
	key := c.layerKey(diskName, layerID) + ".tar.gz"
	_, err := c.s3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		// Check if it's a "not found" error
		var notFound *types.NotFound
		if ok := errors.As(err, &notFound); ok {
			return false, nil
		}
		return false, fmt.Errorf("failed to check layer existence: %w", err)
	}

	return true, nil
}

// DeleteLayer deletes a specific layer from S3
func (c *S3Client) DeleteLayer(ctx context.Context, diskName, layerID string) error {
	key := c.layerKey(diskName, layerID) + ".tar.gz"
	_, err := c.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete layer: %w", err)
	}

	return nil
}
