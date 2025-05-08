// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package storage

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/daytonaio/runner/cmd/runner/config"
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
