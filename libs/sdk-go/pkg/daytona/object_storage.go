// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package daytona

import (
	"archive/tar"
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/daytonaio/daytona/libs/sdk-go/pkg/errors"
)

// objectStorageConfig holds configuration for S3-compatible object storage.
type objectStorageConfig struct {
	EndpointURL     string
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    *string
	BucketName      string
}

// objectStorage handles S3-compatible object storage operations for uploading
// build contexts and other artifacts. This is used internally by the SDK.
type objectStorage struct {
	client     *s3.Client
	bucketName string
}

// NewObjectStorage creates a new objectStorage instance.
//
// This is used internally by the SDK for uploading build contexts.
func NewObjectStorage(config objectStorageConfig) *objectStorage {
	// Extract region from endpoint URL
	region := "us-east-1" // default
	if strings.Contains(config.EndpointURL, "amazonaws.com") {
		parts := strings.Split(config.EndpointURL, ".")
		for i, part := range parts {
			if part == "s3" && i+1 < len(parts) {
				region = parts[i+1]
				break
			}
		}
	}

	bucketName := config.BucketName
	if bucketName == "" {
		bucketName = "daytona-volume-builds"
	}

	// Create credentials
	var sessionToken *string
	if config.SessionToken != nil && *config.SessionToken != "" {
		sessionToken = config.SessionToken
	}

	creds := credentials.NewStaticCredentialsProvider(
		config.AccessKeyID,
		config.SecretAccessKey,
		aws.ToString(sessionToken),
	)

	// s3proxy does not implement AWS SDK v2 checksum headers — suppress unless required by protocol.
	client := s3.New(s3.Options{
		Region:                     region,
		Credentials:                creds,
		BaseEndpoint:               aws.String(config.EndpointURL),
		UsePathStyle:               true,
		RequestChecksumCalculation: aws.RequestChecksumCalculationWhenRequired,
		ResponseChecksumValidation: aws.ResponseChecksumValidationWhenRequired,
	})

	return &objectStorage{
		client:     client,
		bucketName: bucketName,
	}
}

// Upload uploads a file or directory to object storage as a tar archive.
//
// The content is hashed and deduplicated - if the same content exists, it returns
// the existing hash without re-uploading.
//
// Parameters:
//   - path: Local path to the file or directory
//   - organizationID: Organization identifier for the storage prefix
//   - archiveBasePath: Base path within the tar archive
//
// Returns the content hash (used as identifier) or an error.
func (objStorage *objectStorage) Upload(ctx context.Context, path, organizationID, archiveBasePath string) (string, error) {
	// Check if path exists
	if _, err := objStorage.statPath(path); err != nil {
		return "", errors.NewDaytonaError(fmt.Sprintf("Path does not exist: %s", path), 0, nil)
	}

	// Compute hash for the path
	pathHash, err := objStorage.computeHashForPath(path, archiveBasePath)
	if err != nil {
		return "", err
	}

	// Define the S3 prefix
	prefix := fmt.Sprintf("%s/%s/", organizationID, pathHash)
	s3Key := prefix + "context.tar"

	// Check if it already exists in S3
	exists, err := objStorage.folderExistsInS3(ctx, prefix)
	if err != nil {
		return "", err
	}
	if exists {
		return pathHash, nil
	}

	// Upload to S3
	if err := objStorage.uploadAsTar(ctx, s3Key, path, archiveBasePath); err != nil {
		return "", err
	}

	return pathHash, nil
}

// statPath gets file info for a path
func (objStorage *objectStorage) statPath(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

// computeHashForPath computes MD5 hash for a file or directory
func (objStorage *objectStorage) computeHashForPath(pathStr, archiveBasePath string) (string, error) {
	absPath, err := filepath.Abs(pathStr)
	if err != nil {
		return "", errors.NewDaytonaError(fmt.Sprintf("Failed to get absolute path: %v", err), 0, nil)
	}

	hasher := md5.New()

	fileInfo, err := os.Stat(absPath)
	if err != nil {
		return "", errors.NewDaytonaError(fmt.Sprintf("Failed to stat path: %v", err), 0, nil)
	}

	if fileInfo.IsDir() {
		// Hash directory recursively
		err = filepath.Walk(absPath, func(filePath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Get relative path for consistent hashing
			relPath, err := filepath.Rel(absPath, filePath)
			if err != nil {
				return err
			}

			// Combine archive base path with relative path
			archivePath := archiveBasePath
			if relPath != "." {
				archivePath = filepath.Join(archiveBasePath, relPath)
			}

			// Write path to hash
			hasher.Write([]byte(archivePath + "\n"))

			if !info.IsDir() {
				// Hash file contents
				f, err := os.Open(filePath)
				if err != nil {
					return err
				}
				defer f.Close()

				if _, err := io.Copy(hasher, f); err != nil {
					return err
				}
			}

			return nil
		})
		if err != nil {
			return "", errors.NewDaytonaError(fmt.Sprintf("Failed to hash directory: %v", err), 0, nil)
		}
	} else {
		// Hash single file
		hasher.Write([]byte(archiveBasePath + "\n"))
		f, err := os.Open(absPath)
		if err != nil {
			return "", errors.NewDaytonaError(fmt.Sprintf("Failed to open file: %v", err), 0, nil)
		}
		defer f.Close()

		if _, err := io.Copy(hasher, f); err != nil {
			return "", errors.NewDaytonaError(fmt.Sprintf("Failed to hash file: %v", err), 0, nil)
		}
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// folderExistsInS3 checks if a folder exists in S3
func (objStorage *objectStorage) folderExistsInS3(ctx context.Context, prefix string) (bool, error) {
	result, err := objStorage.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket:  aws.String(objStorage.bucketName),
		Prefix:  aws.String(prefix),
		MaxKeys: aws.Int32(1),
	})
	if err != nil {
		return false, errors.NewDaytonaError(fmt.Sprintf("Failed to check S3 folder: %v", err), 0, nil)
	}

	return len(result.Contents) > 0, nil
}

// uploadAsTar creates a tar archive and uploads it to S3
func (objStorage *objectStorage) uploadAsTar(ctx context.Context, s3Key, sourcePath, archiveBasePath string) error {
	// Create tar archive in memory
	var buf bytes.Buffer
	tarWriter := tar.NewWriter(&buf)

	absPath, err := filepath.Abs(sourcePath)
	if err != nil {
		return errors.NewDaytonaError(fmt.Sprintf("Failed to get absolute path: %v", err), 0, nil)
	}

	fileInfo, err := os.Stat(absPath)
	if err != nil {
		return errors.NewDaytonaError(fmt.Sprintf("Failed to stat path: %v", err), 0, nil)
	}

	if fileInfo.IsDir() {
		// Add directory contents to tar
		err = filepath.Walk(absPath, func(filePath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Get relative path
			relPath, err := filepath.Rel(absPath, filePath)
			if err != nil {
				return err
			}

			// Combine archive base path with relative path
			tarPath := archiveBasePath
			if relPath != "." {
				tarPath = filepath.Join(archiveBasePath, relPath)
			}

			// Create tar header
			header, err := tar.FileInfoHeader(info, "")
			if err != nil {
				return err
			}
			header.Name = tarPath

			// Write header
			if err := tarWriter.WriteHeader(header); err != nil {
				return err
			}

			// Write file contents if it's a file
			if !info.IsDir() {
				f, err := os.Open(filePath)
				if err != nil {
					return err
				}
				defer f.Close()

				if _, err := io.Copy(tarWriter, f); err != nil {
					return err
				}
			}

			return nil
		})
		if err != nil {
			return errors.NewDaytonaError(fmt.Sprintf("Failed to create tar archive: %v", err), 0, nil)
		}
	} else {
		// Add single file to tar
		f, err := os.Open(absPath)
		if err != nil {
			return errors.NewDaytonaError(fmt.Sprintf("Failed to open file: %v", err), 0, nil)
		}
		defer f.Close()

		header, err := tar.FileInfoHeader(fileInfo, "")
		if err != nil {
			return errors.NewDaytonaError(fmt.Sprintf("Failed to create tar header: %v", err), 0, nil)
		}
		header.Name = archiveBasePath

		if err := tarWriter.WriteHeader(header); err != nil {
			return errors.NewDaytonaError(fmt.Sprintf("Failed to write tar header: %v", err), 0, nil)
		}

		if _, err := io.Copy(tarWriter, f); err != nil {
			return errors.NewDaytonaError(fmt.Sprintf("Failed to write file to tar: %v", err), 0, nil)
		}
	}

	if err := tarWriter.Close(); err != nil {
		return errors.NewDaytonaError(fmt.Sprintf("Failed to close tar writer: %v", err), 0, nil)
	}

	// Upload to S3
	_, err = objStorage.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(objStorage.bucketName),
		Key:         aws.String(s3Key),
		Body:        bytes.NewReader(buf.Bytes()),
		ContentType: aws.String("application/x-tar"),
	})
	if err != nil {
		return errors.NewDaytonaError(fmt.Sprintf("Failed to upload to S3: %v", err), 0, nil)
	}

	return nil
}

// PushAccessCredentials holds temporary credentials for uploading to object storage.
//
// These credentials are obtained from the API and used for uploading build contexts
// when creating snapshots with custom [DockerImage] definitions.
type PushAccessCredentials struct {
	StorageURL     string `json:"storageUrl"`
	AccessKey      string `json:"accessKey"`
	Secret         string `json:"secret"`
	SessionToken   string `json:"sessionToken"`
	Bucket         string `json:"bucket"`
	OrganizationID string `json:"organizationId"`
}
