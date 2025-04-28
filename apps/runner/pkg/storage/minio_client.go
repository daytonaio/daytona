// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package storage

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const CONTEXT_TAR_FILE_NAME = "context.tar"

type minioClient struct {
	client     *minio.Client
	bucketName string
}

var instance ObjectStorageClient

func GetObjectStorageClient() (ObjectStorageClient, error) {
	if instance != nil {
		return instance, nil
	}

	// TODO: fix to use env vars
	// endpoint := os.Getenv("S3_ENDPOINT")
	// accessKey := os.Getenv("S3_ACCESS_KEY")
	// secretKey := os.Getenv("S3_SECRET_KEY")
	// bucketName := os.Getenv("S3_DEFAULT_BUCKET")
	endpoint := "minio:9000"
	accessKey := "minioadmin"
	secretKey := "minioadmin"
	bucketName := "daytona"
	useSSL := os.Getenv("MINIO_USE_SSL") == "true"

	// Set defaults if not provided
	if endpoint == "" || accessKey == "" || secretKey == "" || bucketName == "" {
		return nil, fmt.Errorf("missing Minio configuration - endpoint, access key, secret key, or bucket name not provided")
	}

	// Initialize Minio client
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Minio client: %w", err)
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
