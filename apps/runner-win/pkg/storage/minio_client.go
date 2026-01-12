// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package storage

import (
	"context"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/daytonaio/runner-win/cmd/runner/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
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

	endpoint = strings.TrimPrefix(endpoint, "http://")
	endpoint = strings.TrimPrefix(endpoint, "https://")

	useSSL := strings.Contains(endpoint, "https")

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

// getSnapshotPath returns the full object path for a snapshot
func (m *minioClient) getSnapshotPath(snapshotName string) string {
	// Strip "snapshots/" prefix if already present (snapshot.ref contains S3 path)
	snapshotName = strings.TrimPrefix(snapshotName, SNAPSHOTS_PREFIX+"/")

	// Ensure snapshot name has .qcow2 extension
	if !strings.HasSuffix(snapshotName, ".qcow2") {
		snapshotName = snapshotName + ".qcow2"
	}
	return path.Join(SNAPSHOTS_PREFIX, snapshotName)
}

// PutSnapshot uploads a snapshot file to the snapshot store
func (m *minioClient) PutSnapshot(ctx context.Context, snapshotName string, reader io.Reader, size int64) (string, error) {
	objectPath := m.getSnapshotPath(snapshotName)

	_, err := m.client.PutObject(ctx, m.bucketName, objectPath, reader, size, minio.PutObjectOptions{
		ContentType: "application/octet-stream",
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload snapshot to storage: %w", err)
	}

	return objectPath, nil
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

	return obj, stat.Size, nil
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
