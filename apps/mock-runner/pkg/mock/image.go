// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package mock

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/daytonaio/mock-runner/pkg/api/dto"
	log "github.com/sirupsen/logrus"
)

// PullImage mocks pulling an image by tracking it in memory
func (m *MockClient) PullImage(ctx context.Context, imageName string, reg *dto.RegistryDTO) error {
	log.Infof("Mock: Pulling image %s", imageName)

	// Normalize image name
	imageName = strings.Replace(imageName, "docker.io/", "", 1)

	// Extract tag
	tag := "latest"
	lastColonIndex := strings.LastIndex(imageName, ":")
	if lastColonIndex != -1 {
		tag = imageName[lastColonIndex+1:]
	}

	// Check if image already exists
	if exists, _ := m.ImageExists(ctx, imageName, true); exists {
		log.Infof("Mock: Image %s already exists", imageName)
		return nil
	}

	// Add image to memory
	hash := generateMockHash(imageName)
	img := &ImageInfo{
		Name:       imageName,
		Tag:        tag,
		Size:       1024 * 1024 * 100, // Mock 100MB
		Entrypoint: []string{"/bin/sh"},
		Cmd:        []string{"-c", "sleep infinity"},
		Hash:       hash,
	}
	m.setImage(img)

	log.Infof("Mock: Image %s pulled successfully", imageName)
	return nil
}

// PushImage mocks pushing an image (no-op, just log)
func (m *MockClient) PushImage(ctx context.Context, imageName string, reg *dto.RegistryDTO) error {
	log.Infof("Mock: Pushing image %s", imageName)

	// Check if image exists
	if exists, _ := m.ImageExists(ctx, imageName, true); !exists {
		return fmt.Errorf("image %s not found", imageName)
	}

	log.Infof("Mock: Image %s pushed successfully", imageName)
	return nil
}

// BuildImage mocks building an image
func (m *MockClient) BuildImage(ctx context.Context, buildDto dto.BuildSnapshotRequestDTO) error {
	log.Infof("Mock: Building image %s", buildDto.Snapshot)

	if !strings.Contains(buildDto.Snapshot, ":") || strings.HasSuffix(buildDto.Snapshot, ":") {
		return fmt.Errorf("invalid image format: must contain exactly one colon (e.g., 'myimage:1.0')")
	}

	// Write mock build logs
	if m.logWriter != nil {
		m.logWriter.Write([]byte("Mock: Building image...\n"))
		m.logWriter.Write([]byte("Mock: Step 1/1 : FROM ubuntu:22.04\n"))
		m.logWriter.Write([]byte("Mock: Successfully built mock-image\n"))
	}

	// Check if image already exists
	if exists, _ := m.ImageExists(ctx, buildDto.Snapshot, true); exists {
		if m.logWriter != nil {
			m.logWriter.Write([]byte("Mock: Image already built\n"))
		}
		return nil
	}

	// Add image to memory
	hash := generateMockHash(buildDto.Snapshot)
	tag := "latest"
	lastColonIndex := strings.LastIndex(buildDto.Snapshot, ":")
	if lastColonIndex != -1 {
		tag = buildDto.Snapshot[lastColonIndex+1:]
	}

	img := &ImageInfo{
		Name:       buildDto.Snapshot,
		Tag:        tag,
		Size:       1024 * 1024 * 200, // Mock 200MB
		Entrypoint: []string{"/bin/sh"},
		Cmd:        []string{"-c", "sleep infinity"},
		Hash:       hash,
	}
	m.setImage(img)

	// Write to build log file
	if err := writeMockBuildLog(buildDto.Snapshot); err != nil {
		log.Warnf("Failed to write build log: %v", err)
	}

	if m.logWriter != nil {
		m.logWriter.Write([]byte("Mock: Image built successfully\n"))
	}

	log.Infof("Mock: Image %s built successfully", buildDto.Snapshot)
	return nil
}

// TagImage mocks tagging an image
func (m *MockClient) TagImage(ctx context.Context, sourceImage string, targetImage string) error {
	log.Infof("Mock: Tagging image %s as %s", sourceImage, targetImage)

	// Check if source image exists
	srcImg, ok := m.getImage(sourceImage)
	if !ok {
		return fmt.Errorf("source image %s not found", sourceImage)
	}

	// Validate target image format
	lastColonIndex := strings.LastIndex(targetImage, ":")
	if lastColonIndex == -1 {
		return fmt.Errorf("invalid target image format: %s", targetImage)
	}

	tag := targetImage[lastColonIndex+1:]
	if tag == "" {
		return fmt.Errorf("invalid target image format: %s", targetImage)
	}

	// Create new image entry with new tag
	newImg := &ImageInfo{
		Name:       targetImage,
		Tag:        tag,
		Size:       srcImg.Size,
		Entrypoint: srcImg.Entrypoint,
		Cmd:        srcImg.Cmd,
		Hash:       srcImg.Hash,
	}
	m.setImage(newImg)

	log.Infof("Mock: Image tagged successfully: %s → %s", sourceImage, targetImage)
	return nil
}

// ImageExists checks if an image exists in memory
func (m *MockClient) ImageExists(ctx context.Context, imageName string, includeLatest bool) (bool, error) {
	imageName = strings.Replace(imageName, "docker.io/", "", 1)

	if strings.HasSuffix(imageName, ":latest") && !includeLatest {
		return false, nil
	}

	_, ok := m.getImage(imageName)
	if ok {
		log.Infof("Mock: Image %s exists", imageName)
	}
	return ok, nil
}

// GetImageInfo returns mock image information
func (m *MockClient) GetImageInfo(ctx context.Context, imageName string) (*ImageInfo, error) {
	img, ok := m.getImage(imageName)
	if !ok {
		return nil, fmt.Errorf("image %s not found", imageName)
	}

	return img, nil
}

// RemoveImage removes an image from memory
func (m *MockClient) RemoveImage(ctx context.Context, imageName string, force bool) error {
	log.Infof("Mock: Removing image %s", imageName)

	_, ok := m.getImage(imageName)
	if !ok {
		log.Infof("Mock: Image %s already removed or not found", imageName)
		return nil
	}

	m.deleteImage(imageName)
	log.Infof("Mock: Image %s removed successfully", imageName)
	return nil
}

// Helper functions

func generateMockHash(imageName string) string {
	hash := sha256.Sum256([]byte(imageName))
	return fmt.Sprintf("sha256:%x", hash)
}

func writeMockBuildLog(snapshotRef string) error {
	buildId := snapshotRef
	if colonIndex := strings.Index(snapshotRef, ":"); colonIndex != -1 {
		buildId = snapshotRef[:colonIndex]
	}

	// Create log directory
	logDir := filepath.Join(os.TempDir(), "mock-runner", "builds")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	logPath := filepath.Join(logDir, buildId)
	logContent := fmt.Sprintf("Mock build log for %s\nBuild completed successfully.\n", snapshotRef)

	return os.WriteFile(logPath, []byte(logContent), 0644)
}
