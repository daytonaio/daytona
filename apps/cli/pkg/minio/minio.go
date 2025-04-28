// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package minio

import (
	"archive/tar"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const CONTEXT_TAR_FILE_NAME = "context.tar"

type Client struct {
	minioClient *minio.Client
	bucket      string
}

func NewClient(endpoint, accessKey, secretKey, bucket string, useSSL bool, sessionToken string) (*Client, error) {
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, sessionToken),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	return &Client{
		minioClient: minioClient,
		bucket:      bucket,
	}, nil
}

func (c *Client) UploadFile(ctx context.Context, objectName string, data []byte) error {
	exists, err := c.minioClient.BucketExists(ctx, c.bucket)
	if err != nil {
		return fmt.Errorf("error checking bucket existence: %w", err)
	}
	if !exists {
		err = c.minioClient.MakeBucket(ctx, c.bucket, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("error creating bucket: %w", err)
		}
	}

	reader := bytes.NewReader(data)
	objectSize := int64(len(data))

	_, err = c.minioClient.PutObject(ctx, c.bucket, objectName, reader, objectSize, minio.PutObjectOptions{
		ContentType: "application/octet-stream",
	})
	if err != nil {
		return fmt.Errorf("error uploading file: %w", err)
	}

	return nil
}

func (c *Client) ListObjects(ctx context.Context, prefix string) ([]string, error) {
	var objects []string

	objectCh := c.minioClient.ListObjects(ctx, c.bucket, minio.ListObjectsOptions{
		Prefix: prefix,
	})

	for object := range objectCh {
		if object.Err != nil {
			return nil, object.Err
		}
		objects = append(objects, object.Key)
	}

	return objects, nil
}

func (c *Client) ProcessDirectory(ctx context.Context, dirPath, orgID string, existingObjects map[string]bool) ([]string, error) {
	tarFile, err := os.Create(CONTEXT_TAR_FILE_NAME)
	if err != nil {
		return nil, fmt.Errorf("failed to create tar file: %w", err)
	}
	defer tarFile.Close()

	tw := tar.NewWriter(tarFile)
	defer tw.Close()

	err = filepath.Walk(dirPath, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(filepath.Dir(dirPath), file)
		if err != nil {
			return err
		}
		header.Name = relPath

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		// Write file contents if regular file
		if fi.Mode().IsRegular() {
			f, err := os.Open(file)
			if err != nil {
				return err
			}
			defer f.Close()

			_, err = io.Copy(tw, f)
			return err
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to process directory: %w", err)
	}

	hasher := sha256.New()
	hasher.Write([]byte(dirPath))
	if _, err := io.Copy(hasher, tarFile); err != nil {
		return nil, fmt.Errorf("failed to hash tar: %w", err)
	}
	hash := hex.EncodeToString(hasher.Sum(nil))

	objectName := fmt.Sprintf("%s/%s", orgID, hash)
	if _, exists := existingObjects[objectName]; !exists {
		err = c.CreateDirectory(ctx, objectName)
		if err != nil {
			return nil, err
		}

		if _, err := tarFile.Seek(0, 0); err != nil {
			return nil, fmt.Errorf("failed to seek tar file: %w", err)
		}

		tarContent, err := io.ReadAll(tarFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read tar file: %w", err)
		}

		err = c.UploadFile(ctx, fmt.Sprintf("%s/%s", objectName, CONTEXT_TAR_FILE_NAME), tarContent)
		if err != nil {
			return nil, fmt.Errorf("failed to upload tar: %w", err)
		}

		if err := os.Remove(CONTEXT_TAR_FILE_NAME); err != nil {
			return nil, fmt.Errorf("failed to remove tar file: %w", err)
		}
	} else {
		fmt.Printf("Directory %s with hash %s already exists in storage\n", dirPath, hash)
	}

	return []string{hash}, nil
}

func (c *Client) ProcessFile(ctx context.Context, filePath, orgID string, existingObjects map[string]bool) (string, error) {
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	hasher := sha256.New()
	hasher.Write(fileContent)
	hash := hex.EncodeToString(hasher.Sum(nil))

	objectName := fmt.Sprintf("%s/%s", orgID, hash)
	if _, exists := existingObjects[objectName]; !exists {
		var tarBuffer bytes.Buffer
		tarWriter := tar.NewWriter(&tarBuffer)

		fileName := filepath.Base(filePath)

		fileInfo, err := os.Stat(filePath)
		if err != nil {
			return "", fmt.Errorf("failed to stat file: %w", err)
		}

		header, err := tar.FileInfoHeader(fileInfo, "")
		if err != nil {
			return "", fmt.Errorf("failed to create tar header: %w", err)
		}
		header.Name = fileName

		if err := tarWriter.WriteHeader(header); err != nil {
			return "", fmt.Errorf("failed to write tar header: %w", err)
		}

		if _, err := tarWriter.Write(fileContent); err != nil {
			return "", fmt.Errorf("failed to write file to tar: %w", err)
		}

		if err := tarWriter.Close(); err != nil {
			return "", fmt.Errorf("failed to close tar writer: %w", err)
		}

		err = c.CreateDirectory(ctx, objectName)
		if err != nil {
			return "", err
		}

		// Upload tar file instead of raw content
		err = c.UploadFile(ctx, fmt.Sprintf("%s/%s", objectName, CONTEXT_TAR_FILE_NAME), tarBuffer.Bytes())
		if err != nil {
			return "", err
		}
	} else {
		fmt.Printf("File %s with hash %s already exists in storage - skipping\n", filePath, hash)
	}

	return hash, nil
}

func (c *Client) CreateDirectory(ctx context.Context, directoryPath string) error {
	exists, err := c.minioClient.BucketExists(ctx, c.bucket)
	if err != nil {
		return fmt.Errorf("error checking bucket existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("bucket does not exist")
	}

	// Ensure the directory path ends with a slash to represent a directory
	if !strings.HasSuffix(directoryPath, "/") {
		directoryPath = directoryPath + "/"
	}

	// Create an empty object to represent the directory
	emptyContent := []byte{}
	reader := bytes.NewReader(emptyContent)
	_, err = c.minioClient.PutObject(ctx, c.bucket, directoryPath, reader, 0, minio.PutObjectOptions{
		ContentType: "application/directory",
	})
	if err != nil {
		return fmt.Errorf("error creating directory marker: %w", err)
	}

	return nil
}
