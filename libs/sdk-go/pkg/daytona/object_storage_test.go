// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package daytona

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewObjectStorage(t *testing.T) {
	tests := []struct {
		name            string
		config          objectStorageConfig
		expectedBucket  string
	}{
		{
			name: "basic config",
			config: objectStorageConfig{
				EndpointURL:     "https://s3.us-east-1.amazonaws.com",
				AccessKeyID:     "AKID",
				SecretAccessKey: "SECRET",
				BucketName:      "my-bucket",
			},
			expectedBucket: "my-bucket",
		},
		{
			name: "default bucket name",
			config: objectStorageConfig{
				EndpointURL:     "https://s3.us-west-2.amazonaws.com",
				AccessKeyID:     "AKID",
				SecretAccessKey: "SECRET",
				BucketName:      "",
			},
			expectedBucket: "daytona-volume-builds",
		},
		{
			name: "with session token",
			config: objectStorageConfig{
				EndpointURL:     "https://s3.eu-west-1.amazonaws.com",
				AccessKeyID:     "AKID",
				SecretAccessKey: "SECRET",
				SessionToken:    strPtr("session-token"),
				BucketName:      "test-bucket",
			},
			expectedBucket: "test-bucket",
		},
		{
			name: "non-aws endpoint",
			config: objectStorageConfig{
				EndpointURL:     "https://minio.local:9000",
				AccessKeyID:     "minioadmin",
				SecretAccessKey: "minioadmin",
				BucketName:      "builds",
			},
			expectedBucket: "builds",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			objStorage := NewObjectStorage(tt.config)
			require.NotNil(t, objStorage)
			assert.Equal(t, tt.expectedBucket, objStorage.bucketName)
			assert.NotNil(t, objStorage.client)
		})
	}
}

func TestComputeHashForFile(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(tmpFile, []byte("hello world"), 0644)
	require.NoError(t, err)

	objStorage := NewObjectStorage(objectStorageConfig{
		EndpointURL:     "https://s3.us-east-1.amazonaws.com",
		AccessKeyID:     "AKID",
		SecretAccessKey: "SECRET",
		BucketName:      "test",
	})

	hash, err := objStorage.computeHashForPath(tmpFile, "test.txt")
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.Len(t, hash, 32)
}

func TestComputeHashForDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "subdir")
	err := os.Mkdir(subDir, 0755)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(subDir, "file1.txt"), []byte("content1"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(subDir, "file2.txt"), []byte("content2"), 0644)
	require.NoError(t, err)

	objStorage := NewObjectStorage(objectStorageConfig{
		EndpointURL:     "https://s3.us-east-1.amazonaws.com",
		AccessKeyID:     "AKID",
		SecretAccessKey: "SECRET",
		BucketName:      "test",
	})

	hash, err := objStorage.computeHashForPath(subDir, "context")
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.Len(t, hash, 32)
}

func TestComputeHashDeterministic(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(tmpFile, []byte("deterministic content"), 0644)
	require.NoError(t, err)

	objStorage := NewObjectStorage(objectStorageConfig{
		EndpointURL:     "https://s3.us-east-1.amazonaws.com",
		AccessKeyID:     "AKID",
		SecretAccessKey: "SECRET",
		BucketName:      "test",
	})

	hash1, err := objStorage.computeHashForPath(tmpFile, "test.txt")
	require.NoError(t, err)

	hash2, err := objStorage.computeHashForPath(tmpFile, "test.txt")
	require.NoError(t, err)

	assert.Equal(t, hash1, hash2)
}

func TestComputeHashDifferentContent(t *testing.T) {
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")
	err := os.WriteFile(file1, []byte("content A"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(file2, []byte("content B"), 0644)
	require.NoError(t, err)

	objStorage := NewObjectStorage(objectStorageConfig{
		EndpointURL:     "https://s3.us-east-1.amazonaws.com",
		AccessKeyID:     "AKID",
		SecretAccessKey: "SECRET",
		BucketName:      "test",
	})

	hash1, err := objStorage.computeHashForPath(file1, "file.txt")
	require.NoError(t, err)

	hash2, err := objStorage.computeHashForPath(file2, "file.txt")
	require.NoError(t, err)

	assert.NotEqual(t, hash1, hash2)
}

func TestStatPathNonExistent(t *testing.T) {
	objStorage := NewObjectStorage(objectStorageConfig{
		EndpointURL:     "https://s3.us-east-1.amazonaws.com",
		AccessKeyID:     "AKID",
		SecretAccessKey: "SECRET",
		BucketName:      "test",
	})

	_, err := objStorage.statPath("/nonexistent/path")
	require.Error(t, err)
}

func TestStatPathExistent(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "exists.txt")
	err := os.WriteFile(tmpFile, []byte("exists"), 0644)
	require.NoError(t, err)

	objStorage := NewObjectStorage(objectStorageConfig{
		EndpointURL:     "https://s3.us-east-1.amazonaws.com",
		AccessKeyID:     "AKID",
		SecretAccessKey: "SECRET",
		BucketName:      "test",
	})

	info, err := objStorage.statPath(tmpFile)
	require.NoError(t, err)
	assert.Equal(t, "exists.txt", info.Name())
}

func TestPushAccessCredentialsStruct(t *testing.T) {
	creds := PushAccessCredentials{
		StorageURL:     "https://s3.us-east-1.amazonaws.com",
		AccessKey:      "AKID",
		Secret:         "SECRET",
		SessionToken:   "TOKEN",
		Bucket:         "my-bucket",
		OrganizationID: "org-1",
	}

	assert.Equal(t, "https://s3.us-east-1.amazonaws.com", creds.StorageURL)
	assert.Equal(t, "AKID", creds.AccessKey)
	assert.Equal(t, "SECRET", creds.Secret)
	assert.Equal(t, "TOKEN", creds.SessionToken)
	assert.Equal(t, "my-bucket", creds.Bucket)
	assert.Equal(t, "org-1", creds.OrganizationID)
}
