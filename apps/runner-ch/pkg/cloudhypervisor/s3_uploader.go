// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package cloudhypervisor

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
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	log "github.com/sirupsen/logrus"
)

// Snapshot storage prefix in the bucket (matches runner-win convention)
const SNAPSHOTS_PREFIX = "snapshots"

// S3Config holds S3 configuration
type S3Config struct {
	Region          string
	EndpointUrl     string
	AccessKeyId     string
	SecretAccessKey string
	Bucket          string
}

// S3Uploader handles uploading snapshots to S3
type S3Uploader struct {
	client     *s3.Client
	uploader   *manager.Uploader
	bucket     string
	chClient   *Client // Reference to CH client for remote file access
	configured bool
}

// NewS3Uploader creates a new S3 uploader
func NewS3Uploader(ctx context.Context, cfg S3Config, chClient *Client) (*S3Uploader, error) {
	// Check if S3 is configured
	if cfg.Bucket == "" || cfg.AccessKeyId == "" || cfg.SecretAccessKey == "" {
		log.Info("S3 not configured - snapshots will only be stored locally")
		return &S3Uploader{configured: false, chClient: chClient}, nil
	}

	// Build AWS config options
	opts := []func(*config.LoadOptions) error{
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyId,
			cfg.SecretAccessKey,
			"",
		)),
	}

	if cfg.Region != "" {
		opts = append(opts, config.WithRegion(cfg.Region))
	}

	awsCfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client with optional custom endpoint
	var s3Client *s3.Client
	if cfg.EndpointUrl != "" {
		s3Client = s3.NewFromConfig(awsCfg, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(cfg.EndpointUrl)
			o.UsePathStyle = true // Required for MinIO and other S3-compatible services
		})
	} else {
		s3Client = s3.NewFromConfig(awsCfg)
	}

	uploader := manager.NewUploader(s3Client, func(u *manager.Uploader) {
		u.PartSize = 64 * 1024 * 1024 // 64 MB parts for large files
		u.Concurrency = 5             // Parallel uploads
	})

	return &S3Uploader{
		client:     s3Client,
		uploader:   uploader,
		bucket:     cfg.Bucket,
		chClient:   chClient,
		configured: true,
	}, nil
}

// IsConfigured returns true if S3 is properly configured
func (u *S3Uploader) IsConfigured() bool {
	return u.configured
}

// UploadSnapshotResult contains the result of a snapshot upload
type UploadSnapshotResult struct {
	S3Path        string           // Full S3 path (bucket/org/name)
	UploadedFiles map[string]int64 // Filename -> size in bytes
	TotalSize     int64            // Total bytes uploaded
	Duration      time.Duration    // Time taken for upload
}

// UploadSnapshot uploads all files in a snapshot directory to S3
// Path format: bucket/snapshots/{organizationId}/{snapshotName}/
// This matches the runner-win convention of prefixing all snapshots with "snapshots/"
func (u *S3Uploader) UploadSnapshot(ctx context.Context, snapshotPath, organizationId, snapshotName string) (*UploadSnapshotResult, error) {
	if !u.configured {
		return nil, fmt.Errorf("S3 is not configured")
	}

	startTime := time.Now()
	s3Prefix := fmt.Sprintf("%s/%s/%s", SNAPSHOTS_PREFIX, organizationId, snapshotName)

	log.Infof("Uploading snapshot to s3://%s/%s", u.bucket, s3Prefix)

	// First, flatten the disk image to remove backing file dependencies
	// This is REQUIRED - the backing file won't exist when downloading to another host
	diskPath := filepath.Join(snapshotPath, "disk.qcow2")
	if err := u.flattenDiskImage(ctx, diskPath); err != nil {
		return nil, fmt.Errorf("failed to flatten disk image (required for S3 upload): %w", err)
	}

	// List files in snapshot directory
	files, err := u.listSnapshotFiles(ctx, snapshotPath)
	if err != nil {
		return nil, fmt.Errorf("failed to list snapshot files: %w", err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no files found in snapshot directory: %s", snapshotPath)
	}

	log.Infof("Found %d files to upload in snapshot", len(files))

	// Upload files concurrently
	result := &UploadSnapshotResult{
		S3Path:        fmt.Sprintf("s3://%s/%s", u.bucket, s3Prefix),
		UploadedFiles: make(map[string]int64),
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	errChan := make(chan error, len(files))

	for _, file := range files {
		wg.Add(1)
		go func(filename string) {
			defer wg.Done()

			s3Key := fmt.Sprintf("%s/%s", s3Prefix, filename)
			localPath := filepath.Join(snapshotPath, filename)

			size, err := u.uploadFile(ctx, localPath, s3Key)
			if err != nil {
				errChan <- fmt.Errorf("failed to upload %s: %w", filename, err)
				return
			}

			mu.Lock()
			result.UploadedFiles[filename] = size
			result.TotalSize += size
			mu.Unlock()

			log.Infof("Uploaded %s (%d bytes)", filename, size)
		}(file)
	}

	wg.Wait()
	close(errChan)

	// Check for errors
	var errs []string
	for err := range errChan {
		errs = append(errs, err.Error())
	}
	if len(errs) > 0 {
		return nil, fmt.Errorf("upload failed: %s", strings.Join(errs, "; "))
	}

	result.Duration = time.Since(startTime)
	log.Infof("Snapshot upload complete: %d files, %d bytes in %v",
		len(result.UploadedFiles), result.TotalSize, result.Duration)

	return result, nil
}

// flattenDiskImage converts a qcow2 with backing file to a standalone image
// This is essential for S3 upload - the backing file won't exist when downloading to another host
func (u *S3Uploader) flattenDiskImage(ctx context.Context, diskPath string) error {
	log.Infof("Checking disk image for backing file: %s", diskPath)

	// Check if the disk has a backing file using qemu-img info
	// The output contains "backing file:" line if there's a backing file
	checkCmd := fmt.Sprintf("qemu-img info %s | grep -q 'backing file:'", diskPath)

	_, err := u.chClient.runShellScript(ctx, checkCmd)
	if err != nil {
		// grep returned non-zero = no "backing file:" found = already standalone
		log.Info("Disk image has no backing file - skipping flatten")
		return nil
	}

	// Disk has a backing file - need to flatten it
	log.Info("Disk has backing file - flattening (this may take several minutes for large disks)...")

	// Get the current disk size for progress indication
	sizeCmd := fmt.Sprintf("qemu-img info %s | grep 'virtual size' | awk '{print $3}'", diskPath)
	sizeOutput, _ := u.chClient.runShellScript(ctx, sizeCmd)
	log.Infof("Virtual disk size: %s", strings.TrimSpace(sizeOutput))

	// Create a flattened copy using qemu-img convert
	// -O qcow2: output format
	// -c: compress the output (optional, reduces size but takes longer)
	// Using a temp file then moving to avoid partial files on failure
	flattenedPath := diskPath + ".flattened"

	// Remove any existing temp file first
	u.chClient.runShellScript(ctx, fmt.Sprintf("rm -f %s", flattenedPath))

	// Convert to standalone image (this reads the full backing chain)
	// -U (--force-share) bypasses lock check - safe because VM should be paused during snapshot
	// -p shows progress (useful for large disks)
	flattenCmd := fmt.Sprintf("qemu-img convert -U -O qcow2 -p %s %s", diskPath, flattenedPath)
	log.Infof("Running: %s", flattenCmd)

	output, err := u.chClient.runShellScript(ctx, flattenCmd)
	if err != nil {
		// Clean up temp file on failure
		u.chClient.runShellScript(ctx, fmt.Sprintf("rm -f %s", flattenedPath))
		return fmt.Errorf("failed to flatten disk: %w (output: %s)", err, output)
	}

	// Verify the flattened image has no backing file
	verifyCmd := fmt.Sprintf("qemu-img info %s | grep -q 'backing file:'", flattenedPath)
	_, err = u.chClient.runShellScript(ctx, verifyCmd)
	if err == nil {
		// Still has backing file - something went wrong
		u.chClient.runShellScript(ctx, fmt.Sprintf("rm -f %s", flattenedPath))
		return fmt.Errorf("flattened image still has backing file reference")
	}

	// Replace original with flattened version
	replaceCmd := fmt.Sprintf("mv %s %s", flattenedPath, diskPath)
	_, err = u.chClient.runShellScript(ctx, replaceCmd)
	if err != nil {
		return fmt.Errorf("failed to replace disk with flattened version: %w", err)
	}

	// Log the new size
	newSizeCmd := fmt.Sprintf("ls -lh %s | awk '{print $5}'", diskPath)
	newSize, _ := u.chClient.runShellScript(ctx, newSizeCmd)
	log.Infof("Disk image flattened successfully (new size: %s)", strings.TrimSpace(newSize))

	return nil
}

// listSnapshotFiles returns the list of files in a snapshot directory
func (u *S3Uploader) listSnapshotFiles(ctx context.Context, snapshotPath string) ([]string, error) {
	listCmd := fmt.Sprintf("ls -1 %s", snapshotPath)
	output, err := u.chClient.runShellScript(ctx, listCmd)
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
func (u *S3Uploader) uploadFile(ctx context.Context, localPath, s3Key string) (int64, error) {
	// For remote CH hosts, we need to stream the file via SSH
	if u.chClient.IsRemote() {
		return u.uploadFileRemote(ctx, localPath, s3Key)
	}

	// Local mode: read file directly
	file, err := os.Open(localPath)
	if err != nil {
		return 0, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return 0, fmt.Errorf("failed to stat file: %w", err)
	}

	filename := filepath.Base(localPath)
	log.Infof("Starting upload of '%s' (%.1f MB) to %s", filename, float64(stat.Size())/(1024*1024), s3Key)

	// Wrap with progress tracking (matches runner-win)
	progressReader := newProgressReader(file, stat.Size(), filename)

	_, err = u.uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(u.bucket),
		Key:           aws.String(s3Key),
		Body:          progressReader,
		ContentLength: aws.Int64(stat.Size()),
	})
	if err != nil {
		return 0, fmt.Errorf("failed to upload to S3: %w", err)
	}

	log.Infof("Completed upload of '%s'", filename)
	return stat.Size(), nil
}

// uploadFileRemote uploads a file from a remote host to S3 by streaming via SSH
func (u *S3Uploader) uploadFileRemote(ctx context.Context, remotePath, s3Key string) (int64, error) {
	// Get file size first
	sizeCmd := fmt.Sprintf("stat -c %%s %s", remotePath)
	sizeOutput, err := u.chClient.runShellScript(ctx, sizeCmd)
	if err != nil {
		return 0, fmt.Errorf("failed to get file size: %w", err)
	}

	var fileSize int64
	fmt.Sscanf(strings.TrimSpace(sizeOutput), "%d", &fileSize)

	filename := filepath.Base(remotePath)
	log.Infof("Starting upload of '%s' (%.1f MB) to %s (streaming from remote)",
		filename, float64(fileSize)/(1024*1024), s3Key)

	// Create a pipe to stream data from SSH to S3
	reader, writer := io.Pipe()

	// Start SSH cat command to stream the file
	cmd := exec.CommandContext(ctx, "ssh",
		"-i", u.chClient.SSHKeyPath,
		"-o", "StrictHostKeyChecking=accept-new",
		"-o", "BatchMode=yes",
		u.chClient.SSHHost,
		fmt.Sprintf("cat %s", remotePath),
	)
	cmd.Stdout = writer

	// Start the SSH command
	if err := cmd.Start(); err != nil {
		writer.Close()
		return 0, fmt.Errorf("failed to start SSH stream: %w", err)
	}

	// Upload in a goroutine with progress tracking
	uploadErr := make(chan error, 1)
	go func() {
		progressReader := newProgressReader(reader, fileSize, filename)
		_, err := u.uploader.Upload(ctx, &s3.PutObjectInput{
			Bucket:        aws.String(u.bucket),
			Key:           aws.String(s3Key),
			Body:          progressReader,
			ContentLength: aws.Int64(fileSize),
		})
		uploadErr <- err
	}()

	// Wait for SSH to complete and close the pipe
	go func() {
		cmd.Wait()
		writer.Close()
	}()

	// Wait for upload to complete
	if err := <-uploadErr; err != nil {
		return 0, fmt.Errorf("failed to upload to S3: %w", err)
	}

	log.Infof("Completed upload of '%s'", filename)
	return fileSize, nil
}

// progressReader wraps an io.Reader to log upload progress (matches runner-win convention)
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

		// Log every 10% or every 30 seconds, whichever comes first (matches runner-win)
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

// DeleteSnapshot deletes a snapshot from S3
func (u *S3Uploader) DeleteSnapshot(ctx context.Context, organizationId, snapshotName string) error {
	if !u.configured {
		return fmt.Errorf("S3 is not configured")
	}

	s3Prefix := fmt.Sprintf("%s/%s/%s/", SNAPSHOTS_PREFIX, organizationId, snapshotName)

	log.Infof("Deleting snapshot from s3://%s/%s", u.bucket, s3Prefix)

	// List all objects with the prefix
	paginator := s3.NewListObjectsV2Paginator(u.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(u.bucket),
		Prefix: aws.String(s3Prefix),
	})

	var objectsToDelete []string
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to list objects: %w", err)
		}
		for _, obj := range page.Contents {
			objectsToDelete = append(objectsToDelete, *obj.Key)
		}
	}

	// Delete all objects
	for _, key := range objectsToDelete {
		_, err := u.client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(u.bucket),
			Key:    aws.String(key),
		})
		if err != nil {
			log.Warnf("Failed to delete %s: %v", key, err)
		}
	}

	log.Infof("Deleted %d objects from S3", len(objectsToDelete))
	return nil
}

// SnapshotExists checks if a snapshot exists in S3
func (u *S3Uploader) SnapshotExists(ctx context.Context, organizationId, snapshotName string) (bool, error) {
	if !u.configured {
		return false, fmt.Errorf("S3 is not configured")
	}

	// Check for the disk.qcow2 file to confirm snapshot exists
	s3Key := fmt.Sprintf("%s/%s/%s/disk.qcow2", SNAPSHOTS_PREFIX, organizationId, snapshotName)

	_, err := u.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(u.bucket),
		Key:    aws.String(s3Key),
	})
	if err != nil {
		// Check if it's a "not found" error
		return false, nil
	}

	return true, nil
}

// DownloadSnapshotResult contains information about a downloaded snapshot
type DownloadSnapshotResult struct {
	LocalPath string        // Local path where snapshot was downloaded
	TotalSize int64         // Total bytes downloaded
	FileCount int           // Number of files downloaded
	Duration  time.Duration // Time taken to download
}

// DownloadSnapshot downloads a snapshot from S3 to the local filesystem
// The snapshot will be downloaded to: {snapshotsPath}/{organizationId}/{snapshotName}/
func (u *S3Uploader) DownloadSnapshot(ctx context.Context, snapshotsPath, organizationId, snapshotName string) (*DownloadSnapshotResult, error) {
	if !u.configured {
		return nil, fmt.Errorf("S3 is not configured")
	}

	startTime := time.Now()
	s3Prefix := fmt.Sprintf("%s/%s/%s/", SNAPSHOTS_PREFIX, organizationId, snapshotName)
	localDir := filepath.Join(snapshotsPath, organizationId, snapshotName)

	log.Infof("Downloading snapshot from s3://%s/%s to %s", u.bucket, s3Prefix, localDir)

	// Create local directory structure
	if err := os.MkdirAll(localDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create local directory: %w", err)
	}

	// List all objects with the prefix
	paginator := s3.NewListObjectsV2Paginator(u.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(u.bucket),
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

	// Download files with concurrency
	var totalSize int64
	var downloadErrors []error
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Use a semaphore to limit concurrent downloads
	sem := make(chan struct{}, 3) // Max 3 concurrent downloads

	for _, s3Key := range filesToDownload {
		wg.Add(1)
		go func(key string) {
			defer wg.Done()
			sem <- struct{}{}        // Acquire
			defer func() { <-sem }() // Release

			// Calculate local file path
			relativePath := strings.TrimPrefix(key, s3Prefix)
			localPath := filepath.Join(localDir, relativePath)

			// Create parent directory if needed
			if dir := filepath.Dir(localPath); dir != localDir {
				if err := os.MkdirAll(dir, 0755); err != nil {
					mu.Lock()
					downloadErrors = append(downloadErrors, fmt.Errorf("failed to create directory %s: %w", dir, err))
					mu.Unlock()
					return
				}
			}

			// Download the file
			size, err := u.downloadFile(ctx, key, localPath)
			if err != nil {
				mu.Lock()
				downloadErrors = append(downloadErrors, fmt.Errorf("failed to download %s: %w", key, err))
				mu.Unlock()
				return
			}

			atomic.AddInt64(&totalSize, size)
			log.Debugf("Downloaded %s (%d bytes)", relativePath, size)
		}(s3Key)
	}

	wg.Wait()

	if len(downloadErrors) > 0 {
		// Clean up on failure
		os.RemoveAll(localDir)
		return nil, fmt.Errorf("failed to download snapshot: %v", downloadErrors[0])
	}

	duration := time.Since(startTime)
	log.Infof("Downloaded snapshot %s/%s: %d files, %d bytes in %v",
		organizationId, snapshotName, len(filesToDownload), totalSize, duration)

	return &DownloadSnapshotResult{
		LocalPath: localDir,
		TotalSize: totalSize,
		FileCount: len(filesToDownload),
		Duration:  duration,
	}, nil
}

// downloadFile downloads a single file from S3
func (u *S3Uploader) downloadFile(ctx context.Context, s3Key, localPath string) (int64, error) {
	// Create the local file
	file, err := os.Create(localPath)
	if err != nil {
		return 0, fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Download using the S3 client
	result, err := u.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(u.bucket),
		Key:    aws.String(s3Key),
	})
	if err != nil {
		os.Remove(localPath)
		return 0, fmt.Errorf("failed to get object: %w", err)
	}
	defer result.Body.Close()

	// Copy the content to the file
	written, err := io.Copy(file, result.Body)
	if err != nil {
		os.Remove(localPath)
		return 0, fmt.Errorf("failed to write file: %w", err)
	}

	return written, nil
}
