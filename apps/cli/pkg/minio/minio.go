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

func readIgnoreFile(rootPath, filename string) []string {
	ignoreFile := filepath.Join(rootPath, filename)
	content, err := os.ReadFile(ignoreFile)
	if err != nil {
		return nil
	}

	var patterns []string
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		patterns = append(patterns, line)
	}
	return patterns
}

func matchPattern(filePath string, patterns []string) bool {
	filePath = filepath.ToSlash(filePath)

	for _, pattern := range patterns {
		pattern = filepath.ToSlash(pattern)

		if strings.HasSuffix(pattern, "/") {
			pattern = strings.TrimSuffix(pattern, "/")
			if strings.HasPrefix(filePath, pattern+"/") || filePath == pattern {
				return true
			}
			continue
		}

		// Handle double star patterns (**) - matches any number of directories
		if strings.Contains(pattern, "**") {
			// Convert ** to a simpler pattern for basic matching
			parts := strings.Split(pattern, "**")
			if len(parts) == 2 {
				prefix := parts[0]
				suffix := parts[1]

				// Remove trailing/leading slashes from prefix/suffix
				prefix = strings.TrimSuffix(prefix, "/")
				suffix = strings.TrimPrefix(suffix, "/")

				if prefix == "" && suffix != "" {
					// Pattern like **/node_modules
					if strings.Contains(filePath, "/"+suffix) || strings.HasSuffix(filePath, suffix) || filePath == suffix {
						return true
					}
				} else if prefix != "" && suffix == "" {
					// Pattern like .git/**
					if strings.HasPrefix(filePath, prefix+"/") || filePath == prefix {
						return true
					}
				} else if prefix != "" && suffix != "" {
					// Pattern like src/**/test
					if strings.HasPrefix(filePath, prefix+"/") && (strings.Contains(filePath, "/"+suffix) || strings.HasSuffix(filePath, suffix)) {
						return true
					}
				}
			}
			continue
		}

		if strings.Contains(pattern, "*") {
			matched, err := filepath.Match(pattern, filepath.Base(filePath))
			if err == nil && matched {
				return true
			}
			// Also check full path for patterns like */node_modules
			matched, err = filepath.Match(pattern, filePath)
			if err == nil && matched {
				return true
			}
			continue
		}

		// Handle exact matches and prefix matches
		if filePath == pattern ||
			strings.HasPrefix(filePath, pattern+"/") ||
			filepath.Base(filePath) == pattern {
			return true
		}
	}
	return false
}

func shouldExcludeFile(filePath, rootPath string) bool {
	relPath, err := filepath.Rel(rootPath, filePath)
	if err != nil {
		return false
	}

	dockerignorePatterns := readIgnoreFile(rootPath, ".dockerignore")

	if len(dockerignorePatterns) == 0 {
		return false
	}

	return matchPattern(relPath, dockerignorePatterns)
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
	// Check if .dockerignore exists and provide helpful message if context is large
	dockerignoreExists := false
	if _, err := os.Stat(filepath.Join(dirPath, ".dockerignore")); err == nil {
		dockerignoreExists = true
	}

	tarFile, err := os.Create(CONTEXT_TAR_FILE_NAME)
	if err != nil {
		return nil, fmt.Errorf("failed to create tar file: %w", err)
	}
	defer tarFile.Close()

	tw := tar.NewWriter(tarFile)
	defer tw.Close()

	fileCount := 0
	totalSize := int64(0)

	warned := false
	err = filepath.Walk(dirPath, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if fi.Name() == CONTEXT_TAR_FILE_NAME {
			return nil
		}

		if shouldExcludeFile(file, dirPath) {
			relPath, _ := filepath.Rel(dirPath, file)
			if fi.IsDir() {
				fmt.Printf("Excluding directory: %s\n", relPath)
				return filepath.SkipDir
			}
			fmt.Printf("Excluding file: %s\n", relPath)
			return nil
		}

		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(dirPath, file)
		if err != nil {
			return err
		}
		header.Name = relPath

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		// Write file contents if regular file
		if fi.Mode().IsRegular() {
			fileCount++
			totalSize += fi.Size()

			if fileCount%1000 == 0 {
				fmt.Printf("Processing... %d files, %.2f MB total\n", fileCount, float64(totalSize)/(1024*1024))
			}

			// Warn if context is getting very large (only warn once)
			if totalSize > 100*1024*1024 && !dockerignoreExists && !warned {
				fmt.Printf("Warning: Context size exceeds 100MB. Consider adding a .dockerignore file to exclude unnecessary files.\n")
				warned = true
			}

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
		if strings.Contains(err.Error(), "write too long") {
			return nil, fmt.Errorf("context directory is too large for tar archive. Please create a .dockerignore file to exclude large directories like .git, node_modules, dist, etc. Original error: %w", err)
		}
		return nil, fmt.Errorf("failed to process directory: %w", err)
	}

	fmt.Printf("Context processing complete: %d files, %.2f MB total\n", fileCount, float64(totalSize)/(1024*1024))

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
