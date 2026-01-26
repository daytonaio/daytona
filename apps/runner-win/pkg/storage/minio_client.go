// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package storage

import (
	"context"
	"fmt"
	"io"
	"path"
	"strings"
	"sync/atomic"
	"time"

	"github.com/daytonaio/runner-win/cmd/runner/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	log "github.com/sirupsen/logrus"
)

const CONTEXT_TAR_FILE_NAME = "context.tar"

// Snapshot storage prefix in the bucket
const SNAPSHOTS_PREFIX = "snapshots"

type minioClient struct {
	client     *minio.Client
	bucketName string
}

var instance ObjectStorageClient

func GetObjectStorageClient() (ObjectStorageClient, error) {
	if instance != nil {
		return instance, nil
	}

	runnerConfig, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	endpoint := runnerConfig.AWSEndpointUrl
	accessKeyId := runnerConfig.AWSAccessKeyId
	secretKey := runnerConfig.AWSSecretAccessKey
	bucketName := runnerConfig.AWSDefaultBucket
	region := runnerConfig.AWSRegion

	// Detect SSL before stripping the protocol prefix
	useSSL := strings.HasPrefix(endpoint, "https://")

	endpoint = strings.TrimPrefix(endpoint, "http://")
	endpoint = strings.TrimPrefix(endpoint, "https://")

	if endpoint == "" || accessKeyId == "" || secretKey == "" || bucketName == "" || region == "" {
		return nil, fmt.Errorf("missing S3 configuration - endpoint, access key, secret key, region, or bucket name not provided")
	}

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyId, secretKey, ""),
		Secure: useSSL,
		Region: region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create S3 client: %w", err)
	}

	instance = &minioClient{
		client:     client,
		bucketName: bucketName,
	}

	return instance, nil
}

func (m *minioClient) GetObject(ctx context.Context, organizationId, hash string) ([]byte, error) {
	objectPath := fmt.Sprintf("%s/%s/%s", organizationId, hash, CONTEXT_TAR_FILE_NAME)
	obj, err := m.client.GetObject(ctx, m.bucketName, objectPath, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object from storage: %w", err)
	}
	defer obj.Close()

	data, err := io.ReadAll(obj)
	if err != nil {
		return nil, fmt.Errorf("failed to read object data: %w", err)
	}

	return data, nil
}

// getSnapshotPath returns the full object path for a snapshot in S3
// Input: snapshot ref like "myapp.qcow2" or "orgId/myapp.qcow2"
// Output: full S3 path like "snapshots/myapp.qcow2" or "snapshots/orgId/myapp.qcow2"
func (m *minioClient) getSnapshotPath(snapshotRef string) string {
	// Ensure snapshot name has .qcow2 extension
	if !strings.HasSuffix(snapshotRef, ".qcow2") {
		snapshotRef = snapshotRef + ".qcow2"
	}
	return path.Join(SNAPSHOTS_PREFIX, snapshotRef)
}

// getSnapshotPathWithOrg returns the full object path for a snapshot with organization namespacing
// The path format is: snapshots/{organizationId}/{snapshotName}.qcow2
func (m *minioClient) getSnapshotPathWithOrg(organizationId, snapshotName string) string {
	// Ensure snapshot name has .qcow2 extension
	if !strings.HasSuffix(snapshotName, ".qcow2") {
		snapshotName = snapshotName + ".qcow2"
	}
	return path.Join(SNAPSHOTS_PREFIX, organizationId, snapshotName)
}

// getSnapshotRef returns the snapshot reference (without snapshots/ prefix)
// Input: snapshot name like "myapp" or "myapp.qcow2"
// Output: "myapp.qcow2"
func (m *minioClient) getSnapshotRef(snapshotName string) string {
	if !strings.HasSuffix(snapshotName, ".qcow2") {
		snapshotName = snapshotName + ".qcow2"
	}
	return snapshotName
}

// getSnapshotRefWithOrg returns the snapshot reference with org namespace (without snapshots/ prefix)
// Input: organizationId and snapshot name
// Output: "{organizationId}/{snapshotName}.qcow2"
func (m *minioClient) getSnapshotRefWithOrg(organizationId, snapshotName string) string {
	if !strings.HasSuffix(snapshotName, ".qcow2") {
		snapshotName = snapshotName + ".qcow2"
	}
	return path.Join(organizationId, snapshotName)
}

// progressReader wraps an io.Reader to log upload progress
type progressReader struct {
	reader      io.Reader
	totalSize   int64
	bytesRead   int64
	lastPercent int
	lastLogTime time.Time
	name        string
}

func newProgressReader(reader io.Reader, size int64, name string) *progressReader {
	return &progressReader{
		reader:      reader,
		totalSize:   size,
		lastLogTime: time.Now(),
		name:        name,
	}
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	if n > 0 {
		atomic.AddInt64(&pr.bytesRead, int64(n))
		currentBytes := atomic.LoadInt64(&pr.bytesRead)

		// Calculate progress
		percent := int(float64(currentBytes) / float64(pr.totalSize) * 100)

		// Log every 10% or every 30 seconds, whichever comes first
		if percent >= pr.lastPercent+10 || time.Since(pr.lastLogTime) > 30*time.Second {
			mbRead := float64(currentBytes) / (1024 * 1024)
			mbTotal := float64(pr.totalSize) / (1024 * 1024)
			log.Infof("Upload progress '%s': %.1f%% (%.1f MB / %.1f MB)", pr.name, float64(percent), mbRead, mbTotal)
			pr.lastPercent = percent
			pr.lastLogTime = time.Now()
		}
	}
	return n, err
}

// PutSnapshot uploads a snapshot file to the snapshot store (legacy, without org namespacing)
// Returns the snapshot name (without the snapshots/ prefix) for use as a reference
func (m *minioClient) PutSnapshot(ctx context.Context, snapshotName string, reader io.Reader, size int64) (string, error) {
	objectPath := m.getSnapshotPath(snapshotName)

	// Wrap reader with progress tracking
	progressReader := newProgressReader(reader, size, snapshotName)

	log.Infof("Starting upload of '%s' (%.1f MB) to %s", snapshotName, float64(size)/(1024*1024), objectPath)

	_, err := m.client.PutObject(ctx, m.bucketName, objectPath, progressReader, size, minio.PutObjectOptions{
		ContentType: "application/octet-stream",
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload snapshot to storage: %w", err)
	}

	log.Infof("Completed upload of '%s'", snapshotName)

	// Return just the snapshot name (without snapshots/ prefix) - the runner handles the prefix internally
	return m.getSnapshotRef(snapshotName), nil
}

// PutSnapshotWithOrg uploads a snapshot file with organization namespacing
// Path format in S3: snapshots/{organizationId}/{snapshotName}.qcow2
// Returns: {organizationId}/{snapshotName}.qcow2 (without snapshots/ prefix)
func (m *minioClient) PutSnapshotWithOrg(ctx context.Context, organizationId, snapshotName string, reader io.Reader, size int64) (string, error) {
	objectPath := m.getSnapshotPathWithOrg(organizationId, snapshotName)

	// Wrap reader with progress tracking
	progressReader := newProgressReader(reader, size, snapshotName)

	log.Infof("Starting upload of '%s' (%.1f MB) to %s (org: %s)", snapshotName, float64(size)/(1024*1024), objectPath, organizationId)

	_, err := m.client.PutObject(ctx, m.bucketName, objectPath, progressReader, size, minio.PutObjectOptions{
		ContentType: "application/octet-stream",
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload snapshot to storage: %w", err)
	}

	log.Infof("Completed upload of '%s' to org %s", snapshotName, organizationId)

	// Return the snapshot ref (without snapshots/ prefix) - the runner handles the prefix internally
	return m.getSnapshotRefWithOrg(organizationId, snapshotName), nil
}

// progressReadCloser wraps an io.ReadCloser to log download progress
type progressReadCloser struct {
	reader      io.ReadCloser
	totalSize   int64
	bytesRead   int64
	lastPercent int
	lastLogTime time.Time
	name        string
}

func (prc *progressReadCloser) Read(p []byte) (int, error) {
	n, err := prc.reader.Read(p)
	if n > 0 {
		atomic.AddInt64(&prc.bytesRead, int64(n))
		currentBytes := atomic.LoadInt64(&prc.bytesRead)

		// Calculate progress
		percent := int(float64(currentBytes) / float64(prc.totalSize) * 100)

		// Log every 10% or every 30 seconds
		if percent >= prc.lastPercent+10 || time.Since(prc.lastLogTime) > 30*time.Second {
			mbRead := float64(currentBytes) / (1024 * 1024)
			mbTotal := float64(prc.totalSize) / (1024 * 1024)
			log.Infof("Download progress '%s': %.1f%% (%.1f MB / %.1f MB)", prc.name, float64(percent), mbRead, mbTotal)
			prc.lastPercent = percent
			prc.lastLogTime = time.Now()
		}
	}
	return n, err
}

func (prc *progressReadCloser) Close() error {
	return prc.reader.Close()
}

// GetSnapshot retrieves a snapshot file from the snapshot store
func (m *minioClient) GetSnapshot(ctx context.Context, snapshotName string) (io.ReadCloser, int64, error) {
	objectPath := m.getSnapshotPath(snapshotName)

	obj, err := m.client.GetObject(ctx, m.bucketName, objectPath, minio.GetObjectOptions{})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get snapshot from storage: %w", err)
	}

	// Get object info for size
	stat, err := obj.Stat()
	if err != nil {
		obj.Close()
		return nil, 0, fmt.Errorf("failed to get snapshot info: %w", err)
	}

	log.Infof("Starting download of '%s' (%.1f MB)", snapshotName, float64(stat.Size)/(1024*1024))

	// Wrap with progress tracking
	progressReader := &progressReadCloser{
		reader:      obj,
		totalSize:   stat.Size,
		lastLogTime: time.Now(),
		name:        snapshotName,
	}

	return progressReader, stat.Size, nil
}

// DeleteSnapshot removes a snapshot from the snapshot store
func (m *minioClient) DeleteSnapshot(ctx context.Context, snapshotName string) error {
	objectPath := m.getSnapshotPath(snapshotName)

	err := m.client.RemoveObject(ctx, m.bucketName, objectPath, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete snapshot from storage: %w", err)
	}

	return nil
}

// SnapshotExists checks if a snapshot exists in the store
func (m *minioClient) SnapshotExists(ctx context.Context, snapshotName string) (bool, error) {
	objectPath := m.getSnapshotPath(snapshotName)

	_, err := m.client.StatObject(ctx, m.bucketName, objectPath, minio.StatObjectOptions{})
	if err != nil {
		// Check if it's a "not found" error
		errResp := minio.ToErrorResponse(err)
		if errResp.Code == "NoSuchKey" {
			return false, nil
		}
		return false, fmt.Errorf("failed to check snapshot existence: %w", err)
	}

	return true, nil
}
