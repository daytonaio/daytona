// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package cuttlefish

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	log "github.com/sirupsen/logrus"
)

// Snapshot storage prefix in the bucket (matches runner-win/runner-ch convention)
const SNAPSHOTS_PREFIX = "snapshots"

// S3Config holds S3 configuration
type S3Config struct {
	Region          string
	EndpointUrl     string
	AccessKeyId     string
	SecretAccessKey string
	Bucket          string
}

// S3Client handles uploading/downloading snapshots to/from S3
type S3Client struct {
	client     *s3.Client
	downloader *manager.Downloader
	uploader   *manager.Uploader
	bucket     string
	cvdClient  *Client // Reference to CVD client for remote file access
	configured bool
}

// NewS3Client creates a new S3 client for snapshot management
func NewS3Client(ctx context.Context, cfg S3Config, cvdClient *Client) (*S3Client, error) {
	if cfg.Bucket == "" || cfg.AccessKeyId == "" || cfg.SecretAccessKey == "" {
		log.Info("S3 not configured - snapshots will only be stored locally")
		return &S3Client{configured: false, cvdClient: cvdClient}, nil
	}

	opts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyId,
			cfg.SecretAccessKey,
			"",
		)),
	}

	if cfg.Region != "" {
		opts = append(opts, awsconfig.WithRegion(cfg.Region))
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	var s3Client *s3.Client
	if cfg.EndpointUrl != "" {
		s3Client = s3.NewFromConfig(awsCfg, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(cfg.EndpointUrl)
			o.UsePathStyle = true
		})
	} else {
		s3Client = s3.NewFromConfig(awsCfg)
	}

	uploader := manager.NewUploader(s3Client, func(u *manager.Uploader) {
		u.PartSize = 64 * 1024 * 1024 // 64 MB parts
		u.Concurrency = 5
	})

	downloader := manager.NewDownloader(s3Client, func(d *manager.Downloader) {
		d.PartSize = 64 * 1024 * 1024
		d.Concurrency = 5
	})

	log.Infof("S3 configured: bucket=%s, region=%s", cfg.Bucket, cfg.Region)

	return &S3Client{
		client:     s3Client,
		uploader:   uploader,
		downloader: downloader,
		bucket:     cfg.Bucket,
		cvdClient:  cvdClient,
		configured: true,
	}, nil
}

// IsConfigured returns true if S3 is properly configured
func (c *S3Client) IsConfigured() bool {
	return c.configured
}

// UploadSnapshotResult contains the result of a snapshot upload
type UploadSnapshotResult struct {
	S3Path        string           `json:"s3Path"`
	UploadedFiles map[string]int64 `json:"uploadedFiles"`
	TotalSize     int64            `json:"totalSize"`
	Duration      time.Duration    `json:"duration"`
}

// UploadSnapshot uploads all files in a snapshot directory to S3
// Path format: bucket/snapshots/{organizationId}/{snapshotName}/
func (c *S3Client) UploadSnapshot(ctx context.Context, snapshotPath, organizationId, snapshotName string) (*UploadSnapshotResult, error) {
	if !c.configured {
		return nil, fmt.Errorf("S3 is not configured")
	}

	startTime := time.Now()
	s3Prefix := fmt.Sprintf("%s/%s/%s", SNAPSHOTS_PREFIX, organizationId, snapshotName)

	log.Infof("Uploading snapshot to s3://%s/%s", c.bucket, s3Prefix)

	// List files in snapshot directory (follow symlinks to get actual files)
	files, err := c.listSnapshotFiles(ctx, snapshotPath)
	if err != nil {
		return nil, fmt.Errorf("failed to list snapshot files: %w", err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no files found in snapshot directory: %s", snapshotPath)
	}

	log.Infof("Found %d files to upload in snapshot", len(files))

	result := &UploadSnapshotResult{
		S3Path:        fmt.Sprintf("s3://%s/%s", c.bucket, s3Prefix),
		UploadedFiles: make(map[string]int64),
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	errChan := make(chan error, len(files))

	// Limit concurrency for large files
	sem := make(chan struct{}, 3)

	for _, file := range files {
		wg.Add(1)
		go func(filename string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			s3Key := fmt.Sprintf("%s/%s", s3Prefix, filename)
			localPath := filepath.Join(snapshotPath, filename)

			// Resolve symlinks to get the real file
			realPath, err := c.resolveSymlink(ctx, localPath)
			if err != nil {
				errChan <- fmt.Errorf("failed to resolve %s: %w", filename, err)
				return
			}

			size, err := c.uploadFile(ctx, realPath, s3Key)
			if err != nil {
				errChan <- fmt.Errorf("failed to upload %s: %w", filename, err)
				return
			}

			mu.Lock()
			result.UploadedFiles[filename] = size
			result.TotalSize += size
			mu.Unlock()

			log.Infof("Uploaded %s (%.1f MB)", filename, float64(size)/(1024*1024))
		}(file)
	}

	wg.Wait()
	close(errChan)

	var errs []string
	for err := range errChan {
		errs = append(errs, err.Error())
	}
	if len(errs) > 0 {
		return nil, fmt.Errorf("upload failed: %s", strings.Join(errs, "; "))
	}

	result.Duration = time.Since(startTime)
	log.Infof("Snapshot upload complete: %d files, %.1f MB in %v",
		len(result.UploadedFiles), float64(result.TotalSize)/(1024*1024), result.Duration)

	return result, nil
}

// DownloadSnapshotResult contains information about a downloaded snapshot
type DownloadSnapshotResult struct {
	LocalPath string        `json:"localPath"`
	TotalSize int64         `json:"totalSize"`
	FileCount int           `json:"fileCount"`
	Duration  time.Duration `json:"duration"`
}

// DownloadSnapshot downloads a snapshot from S3 to the local filesystem
func (c *S3Client) DownloadSnapshot(ctx context.Context, snapshotsPath, organizationId, snapshotName string) (*DownloadSnapshotResult, error) {
	if !c.configured {
		return nil, fmt.Errorf("S3 is not configured")
	}

	startTime := time.Now()
	s3Prefix := fmt.Sprintf("%s/%s/%s/", SNAPSHOTS_PREFIX, organizationId, snapshotName)
	localDir := filepath.Join(snapshotsPath, organizationId, snapshotName)

	log.Infof("Downloading snapshot from s3://%s/%s to %s", c.bucket, s3Prefix, localDir)

	// Create local directory
	if c.cvdClient.IsRemote() {
		if err := c.cvdClient.runCommand(ctx, "mkdir", "-p", localDir); err != nil {
			return nil, fmt.Errorf("failed to create directory: %w", err)
		}
	} else {
		if err := os.MkdirAll(localDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// List all objects with the prefix
	paginator := s3.NewListObjectsV2Paginator(c.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(c.bucket),
		Prefix: aws.String(s3Prefix),
	})

	var filesToDownload []string
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list objects: %w", err)
		}
		for _, obj := range page.Contents {
			filesToDownload = append(filesToDownload, *obj.Key)
		}
	}

	if len(filesToDownload) == 0 {
		return nil, fmt.Errorf("snapshot not found in S3: %s/%s", organizationId, snapshotName)
	}

	log.Infof("Found %d files to download", len(filesToDownload))

	var totalSize int64
	var downloadErrors []error
	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, 3)

	for _, s3Key := range filesToDownload {
		wg.Add(1)
		go func(key string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			relativePath := strings.TrimPrefix(key, s3Prefix)
			localPath := filepath.Join(localDir, relativePath)

			size, err := c.downloadFile(ctx, key, localPath)
			if err != nil {
				mu.Lock()
				downloadErrors = append(downloadErrors, fmt.Errorf("failed to download %s: %w", key, err))
				mu.Unlock()
				return
			}

			atomic.AddInt64(&totalSize, size)
			log.Debugf("Downloaded %s (%.1f MB)", relativePath, float64(size)/(1024*1024))
		}(s3Key)
	}

	wg.Wait()

	if len(downloadErrors) > 0 {
		return nil, fmt.Errorf("failed to download snapshot: %v", downloadErrors[0])
	}

	duration := time.Since(startTime)
	log.Infof("Downloaded snapshot %s/%s: %d files, %.1f MB in %v",
		organizationId, snapshotName, len(filesToDownload), float64(totalSize)/(1024*1024), duration)

	return &DownloadSnapshotResult{
		LocalPath: localDir,
		TotalSize: totalSize,
		FileCount: len(filesToDownload),
		Duration:  duration,
	}, nil
}

// SnapshotExistsInS3 checks if a snapshot exists in S3
func (c *S3Client) SnapshotExistsInS3(ctx context.Context, organizationId, snapshotName string) (bool, error) {
	if !c.configured {
		return false, nil
	}

	// Check for manifest.json to confirm snapshot exists
	s3Key := fmt.Sprintf("%s/%s/%s/manifest.json", SNAPSHOTS_PREFIX, organizationId, snapshotName)

	_, err := c.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(s3Key),
	})
	if err != nil {
		return false, nil
	}

	return true, nil
}

// DeleteSnapshotFromS3 deletes a snapshot from S3
func (c *S3Client) DeleteSnapshotFromS3(ctx context.Context, organizationId, snapshotName string) error {
	if !c.configured {
		return fmt.Errorf("S3 is not configured")
	}

	s3Prefix := fmt.Sprintf("%s/%s/%s/", SNAPSHOTS_PREFIX, organizationId, snapshotName)

	log.Infof("Deleting snapshot from s3://%s/%s", c.bucket, s3Prefix)

	paginator := s3.NewListObjectsV2Paginator(c.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(c.bucket),
		Prefix: aws.String(s3Prefix),
	})

	var deleted int
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to list objects: %w", err)
		}
		for _, obj := range page.Contents {
			_, err := c.client.DeleteObject(ctx, &s3.DeleteObjectInput{
				Bucket: aws.String(c.bucket),
				Key:    obj.Key,
			})
			if err != nil {
				log.Warnf("Failed to delete %s: %v", *obj.Key, err)
			} else {
				deleted++
			}
		}
	}

	log.Infof("Deleted %d objects from S3", deleted)
	return nil
}

// resolveSymlink resolves a symlink to its real path
func (c *S3Client) resolveSymlink(ctx context.Context, path string) (string, error) {
	if c.cvdClient.IsRemote() {
		cmd := fmt.Sprintf("readlink -f %s", path)
		output, err := c.cvdClient.runShellScript(ctx, cmd)
		if err != nil {
			return path, nil // If readlink fails, use original path
		}
		resolved := strings.TrimSpace(output)
		if resolved != "" {
			return resolved, nil
		}
		return path, nil
	}

	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return path, nil
	}
	return resolved, nil
}

// listSnapshotFiles returns files in the snapshot directory (resolving symlinks, skipping dirs)
func (c *S3Client) listSnapshotFiles(ctx context.Context, snapshotPath string) ([]string, error) {
	// List only regular files (follow symlinks)
	listCmd := fmt.Sprintf("find -L %s -maxdepth 1 -type f -printf '%%f\\n' 2>/dev/null", snapshotPath)
	output, err := c.cvdClient.runShellScript(ctx, listCmd)
	if err != nil {
		return nil, fmt.Errorf("failed to list directory: %w", err)
	}

	var files []string
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, ".") {
			files = append(files, line)
		}
	}

	return files, nil
}

// uploadFile uploads a single file to S3
func (c *S3Client) uploadFile(ctx context.Context, localPath, s3Key string) (int64, error) {
	if c.cvdClient.IsRemote() {
		return c.uploadFileRemote(ctx, localPath, s3Key)
	}

	file, err := os.Open(localPath)
	if err != nil {
		return 0, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return 0, fmt.Errorf("failed to stat file: %w", err)
	}

	_, err = c.uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(c.bucket),
		Key:           aws.String(s3Key),
		Body:          file,
		ContentLength: aws.Int64(stat.Size()),
	})
	if err != nil {
		return 0, fmt.Errorf("failed to upload: %w", err)
	}

	return stat.Size(), nil
}

// uploadFileRemote uploads a file from a remote host to S3 by streaming via SSH
func (c *S3Client) uploadFileRemote(ctx context.Context, remotePath, s3Key string) (int64, error) {
	sizeCmd := fmt.Sprintf("stat -c %%s %s", remotePath)
	sizeOutput, err := c.cvdClient.runShellScript(ctx, sizeCmd)
	if err != nil {
		return 0, fmt.Errorf("failed to get file size: %w", err)
	}

	var fileSize int64
	fmt.Sscanf(strings.TrimSpace(sizeOutput), "%d", &fileSize)

	reader, writer := io.Pipe()

	cmd := exec.CommandContext(ctx, "ssh",
		"-i", c.cvdClient.SSHKeyPath,
		"-o", "StrictHostKeyChecking=accept-new",
		"-o", "BatchMode=yes",
		c.cvdClient.SSHHost,
		fmt.Sprintf("cat %s", remotePath),
	)
	cmd.Stdout = writer

	if err := cmd.Start(); err != nil {
		writer.Close()
		return 0, fmt.Errorf("failed to start SSH stream: %w", err)
	}

	uploadErr := make(chan error, 1)
	go func() {
		_, err := c.uploader.Upload(ctx, &s3.PutObjectInput{
			Bucket:        aws.String(c.bucket),
			Key:           aws.String(s3Key),
			Body:          reader,
			ContentLength: aws.Int64(fileSize),
		})
		uploadErr <- err
	}()

	go func() {
		cmd.Wait()
		writer.Close()
	}()

	if err := <-uploadErr; err != nil {
		return 0, fmt.Errorf("failed to upload: %w", err)
	}

	return fileSize, nil
}

// downloadFile downloads a single file from S3
func (c *S3Client) downloadFile(ctx context.Context, s3Key, localPath string) (int64, error) {
	if c.cvdClient.IsRemote() {
		return c.downloadFileRemote(ctx, s3Key, localPath)
	}

	file, err := os.Create(localPath)
	if err != nil {
		return 0, fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	result, err := c.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(s3Key),
	})
	if err != nil {
		os.Remove(localPath)
		return 0, fmt.Errorf("failed to get object: %w", err)
	}
	defer result.Body.Close()

	written, err := io.Copy(file, result.Body)
	if err != nil {
		os.Remove(localPath)
		return 0, fmt.Errorf("failed to write file: %w", err)
	}

	return written, nil
}

// downloadFileRemote downloads a file from S3 to a remote host via SSH pipe
func (c *S3Client) downloadFileRemote(ctx context.Context, s3Key, remotePath string) (int64, error) {
	result, err := c.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(s3Key),
	})
	if err != nil {
		return 0, fmt.Errorf("failed to get object: %w", err)
	}
	defer result.Body.Close()

	cmd := exec.CommandContext(ctx, "ssh",
		"-i", c.cvdClient.SSHKeyPath,
		"-o", "StrictHostKeyChecking=accept-new",
		"-o", "BatchMode=yes",
		c.cvdClient.SSHHost,
		fmt.Sprintf("cat > %s", remotePath),
	)
	cmd.Stdin = result.Body

	if err := cmd.Run(); err != nil {
		return 0, fmt.Errorf("failed to write to remote: %w", err)
	}

	var size int64
	if result.ContentLength != nil {
		size = *result.ContentLength
	}

	return size, nil
}
