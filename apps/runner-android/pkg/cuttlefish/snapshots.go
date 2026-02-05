// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package cuttlefish

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// SnapshotType indicates whether a snapshot is a base image or custom snapshot
type SnapshotType string

const (
	SnapshotTypeBase   SnapshotType = "base"
	SnapshotTypeCustom SnapshotType = "custom"
)

// SnapshotMetadata contains metadata about a snapshot
type SnapshotMetadata struct {
	Name         string       `json:"name"`
	Type         SnapshotType `json:"type"`
	OrgId        string       `json:"orgId,omitempty"` // Only for custom snapshots
	Description  string       `json:"description,omitempty"`
	BaseRef      string       `json:"baseRef,omitempty"` // For custom: which base image it was created from
	CreatedAt    time.Time    `json:"createdAt"`
	CreatedFrom  string       `json:"createdFrom,omitempty"` // SandboxId it was created from
	SizeBytes    int64        `json:"sizeBytes,omitempty"`
	Architecture string       `json:"architecture,omitempty"`
	FormFactor   string       `json:"formFactor,omitempty"` // phone, tablet, tv, auto, etc.
}

// SnapshotInfo represents a snapshot available for use
type SnapshotInfo struct {
	// Path is the full snapshot path (e.g., "aosp_cf_x86_64_phone" or "orgId/my-snapshot")
	Path     string           `json:"path"`
	Metadata SnapshotMetadata `json:"metadata"`
}

// uuidRegex matches UUID format (used to detect org IDs)
var uuidRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// IsBaseSnapshot checks if a snapshot path refers to a base system image
// Base snapshots have no parent folder (no "/" in the path)
func IsBaseSnapshot(snapshotPath string) bool {
	return !strings.Contains(snapshotPath, "/")
}

// IsCustomSnapshot checks if a snapshot path refers to a custom org snapshot
// Custom snapshots have a UUID org ID as parent folder
func IsCustomSnapshot(snapshotPath string) bool {
	parts := strings.SplitN(snapshotPath, "/", 2)
	if len(parts) < 2 {
		return false
	}
	return uuidRegex.MatchString(parts[0])
}

// ParseSnapshotPath parses a snapshot path into its components
// - "aosp_cf_x86_64_phone" -> (base, "", "aosp_cf_x86_64_phone")
// - "orgId/my-snapshot" -> (custom, "orgId", "my-snapshot")
func ParseSnapshotPath(snapshotPath string) (snapshotType SnapshotType, orgId string, name string) {
	parts := strings.SplitN(snapshotPath, "/", 2)

	// No "/" means it's a base snapshot
	if len(parts) < 2 {
		return SnapshotTypeBase, "", snapshotPath
	}

	// Check if first part is a UUID (org ID)
	if uuidRegex.MatchString(parts[0]) {
		return SnapshotTypeCustom, parts[0], parts[1]
	}

	// Unknown format, treat as base with the full path as name
	return SnapshotTypeBase, "", snapshotPath
}

// GetSnapshotsDir returns the base directory for custom org snapshots
func (c *Client) GetSnapshotsDir() string {
	return filepath.Join(c.config.ArtifactsPath, "snapshots")
}

// GetSnapshotDir returns the full path to a specific snapshot
// For base snapshots (e.g., "aosp_cf_x86_64_phone"): looks in CVDHome
// For custom snapshots (e.g., "orgId/my-snapshot"): looks in ArtifactsPath/snapshots/
func (c *Client) GetSnapshotDir(snapshotPath string) string {
	if IsBaseSnapshot(snapshotPath) {
		// Base snapshots are stored in CVDHome directory
		return filepath.Join(c.config.CVDHome, snapshotPath)
	}
	// Custom snapshots are stored in ArtifactsPath/snapshots/
	return filepath.Join(c.GetSnapshotsDir(), snapshotPath)
}

// ListSnapshots returns all available snapshots
// Base snapshots are found in CVDHome (e.g., /home/vsoc-01/aosp_cf_x86_64_phone/)
// Custom snapshots are found in ArtifactsPath/snapshots/orgId/snapshotName/
func (c *Client) ListSnapshots(ctx context.Context) ([]string, error) {
	var snapshots []string

	// 1. List base snapshots from CVDHome
	// These are directories like aosp_cf_x86_64_phone, aosp_cf_x86_64_tablet, etc.
	baseEntries, err := c.listDirectory(ctx, c.config.CVDHome)
	if err != nil {
		log.Warnf("Failed to list CVDHome directory: %v", err)
	} else {
		for _, entry := range baseEntries {
			// Skip non-AOSP directories and common Cuttlefish directories
			if !c.isAndroidSystemImage(ctx, entry) {
				continue
			}
			snapshots = append(snapshots, entry)
		}
	}

	// 2. List custom org snapshots from ArtifactsPath/snapshots/
	customSnapshotsDir := c.GetSnapshotsDir()
	orgEntries, err := c.listDirectory(ctx, customSnapshotsDir)
	if err != nil {
		log.Debugf("No custom snapshots directory or failed to list: %v", err)
	} else {
		for _, entry := range orgEntries {
			// Only process UUID directories (org IDs)
			if !uuidRegex.MatchString(entry) {
				continue
			}
			orgDir := filepath.Join(customSnapshotsDir, entry)
			orgSnapshots, err := c.listSnapshotsInDir(ctx, orgDir, entry)
			if err != nil {
				log.Warnf("Failed to list snapshots for org %s: %v", entry, err)
				continue
			}
			snapshots = append(snapshots, orgSnapshots...)
		}
	}

	return snapshots, nil
}

// isAndroidSystemImage checks if a directory looks like an Android system image
func (c *Client) isAndroidSystemImage(ctx context.Context, name string) bool {
	// Skip common non-image directories
	skipDirs := map[string]bool{
		"bin": true, "cuttlefish": true, "cuttlefish_runtime": true,
		".cache": true, ".config": true, ".local": true, ".android": true,
		".cursor": true, ".cursor-server": true, ".bashrc": true,
		"logs": true, "tmp": true, "android-cuttlefish": true,
	}
	if skipDirs[name] {
		return false
	}
	// Skip hidden directories
	if strings.HasPrefix(name, ".") {
		return false
	}

	// Check if it looks like an Android image directory
	dirPath := filepath.Join(c.config.CVDHome, name)

	// First verify it's a directory
	isDir, err := c.isDirectory(ctx, dirPath)
	if err != nil || !isDir {
		return false
	}

	// Check for typical Cuttlefish image files
	// Common patterns: aosp_cf_x86_64_phone, aosp_cf_arm64_phone, cf_vm, cuttlefish_export, etc.
	if strings.HasPrefix(name, "aosp_cf_") || strings.HasPrefix(name, "cf_") ||
		strings.HasPrefix(name, "cuttlefish_") {
		// Verify it actually contains images
		superImgPath := filepath.Join(dirPath, "super.img")
		if exists, _ := c.fileExists(ctx, superImgPath); exists {
			return true
		}
	}

	// Check for super.img (modern Cuttlefish) or system.img (legacy)
	superImgPath := filepath.Join(dirPath, "super.img")
	if exists, _ := c.fileExists(ctx, superImgPath); exists {
		return true
	}

	systemImgPath := filepath.Join(dirPath, "system.img")
	exists, _ := c.fileExists(ctx, systemImgPath)
	return exists
}

// ListSnapshotsWithInfo returns all snapshots with their metadata
func (c *Client) ListSnapshotsWithInfo(ctx context.Context) ([]SnapshotInfo, error) {
	paths, err := c.ListSnapshots(ctx)
	if err != nil {
		return nil, err
	}

	var infos []SnapshotInfo
	for _, path := range paths {
		metadata, err := c.GetSnapshotMetadata(ctx, path)
		if err != nil {
			// Create basic metadata if file doesn't exist
			snapshotType, orgId, name := ParseSnapshotPath(path)
			metadata = &SnapshotMetadata{
				Name:  name,
				Type:  snapshotType,
				OrgId: orgId,
			}
		}
		infos = append(infos, SnapshotInfo{
			Path:     path,
			Metadata: *metadata,
		})
	}

	return infos, nil
}

// ListOrgSnapshots returns snapshots for a specific organization
func (c *Client) ListOrgSnapshots(ctx context.Context, orgId string) ([]SnapshotInfo, error) {
	snapshotsDir := c.GetSnapshotsDir()
	orgDir := filepath.Join(snapshotsDir, orgId)

	paths, err := c.listSnapshotsInDir(ctx, orgDir, orgId)
	if err != nil {
		return nil, err
	}

	var infos []SnapshotInfo
	for _, path := range paths {
		metadata, err := c.GetSnapshotMetadata(ctx, path)
		if err != nil {
			metadata = &SnapshotMetadata{
				Name:  filepath.Base(path),
				Type:  SnapshotTypeCustom,
				OrgId: orgId,
			}
		}
		infos = append(infos, SnapshotInfo{
			Path:     path,
			Metadata: *metadata,
		})
	}

	return infos, nil
}

// GetSnapshotMetadata reads metadata for a snapshot
func (c *Client) GetSnapshotMetadata(ctx context.Context, snapshotPath string) (*SnapshotMetadata, error) {
	metadataPath := filepath.Join(c.GetSnapshotDir(snapshotPath), "manifest.json")

	data, err := c.readFile(ctx, metadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read snapshot metadata: %w", err)
	}

	var metadata SnapshotMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse snapshot metadata: %w", err)
	}

	return &metadata, nil
}

// SnapshotExists checks if a snapshot exists
func (c *Client) SnapshotExists(ctx context.Context, snapshotPath string) (bool, error) {
	snapshotDir := c.GetSnapshotDir(snapshotPath)
	exists, err := c.fileExists(ctx, snapshotDir)
	if err != nil {
		return false, err
	}

	if !exists {
		return false, nil
	}

	// For base snapshots, verify it contains Android system images
	if IsBaseSnapshot(snapshotPath) {
		// Check for super.img (modern Cuttlefish) or system.img (legacy)
		// Modern Cuttlefish uses super.img which contains system, vendor, product partitions
		superImgPath := filepath.Join(snapshotDir, "super.img")
		hasSuperImg, _ := c.fileExists(ctx, superImgPath)
		if hasSuperImg {
			return true, nil
		}

		// Fallback: check for system.img (legacy format)
		systemImgPath := filepath.Join(snapshotDir, "system.img")
		hasSystemImg, _ := c.fileExists(ctx, systemImgPath)
		if !hasSystemImg {
			log.Warnf("Snapshot directory %s exists but doesn't contain super.img or system.img", snapshotDir)
			return false, nil
		}
	}

	return true, nil
}

// GetSnapshotNotFoundHelp returns helpful information when a snapshot is not found
func (c *Client) GetSnapshotNotFoundHelp(snapshotPath string) string {
	if IsBaseSnapshot(snapshotPath) {
		return fmt.Sprintf(
			"Base snapshot '%s' not found. Android system images need to be downloaded first.\n"+
				"To download images, run on the Cuttlefish host:\n"+
				"  cvd fetch --target_directory=%s aosp-main-throttled/aosp_cf_x86_64_phone-trunk_staging-userdebug\n"+
				"Or download manually from https://ci.android.com and extract to %s",
			snapshotPath,
			c.GetSnapshotDir(snapshotPath),
			c.GetSnapshotDir(snapshotPath),
		)
	}
	return fmt.Sprintf("Custom snapshot '%s' not found", snapshotPath)
}

// CreateSnapshotFromInstance creates a custom snapshot from a running instance
func (c *Client) CreateSnapshotFromInstance(ctx context.Context, sandboxId string, orgId string, snapshotName string, description string) (*SnapshotInfo, error) {
	log.Infof("Creating snapshot '%s' for org '%s' from sandbox '%s'", snapshotName, orgId, sandboxId)

	// Get instance info
	info, exists := c.GetInstance(sandboxId)
	if !exists {
		return nil, fmt.Errorf("sandbox %s not found", sandboxId)
	}

	// Check instance is running
	state := c.getInstanceState(ctx, info.InstanceNum)
	if state != InstanceStateRunning {
		return nil, fmt.Errorf("sandbox must be running to create snapshot (current state: %s)", state)
	}

	// Create snapshot directory
	snapshotPath := fmt.Sprintf("%s/%s", orgId, snapshotName)
	snapshotDir := c.GetSnapshotDir(snapshotPath)

	if err := c.runCommand(ctx, "mkdir", "-p", snapshotDir); err != nil {
		return nil, fmt.Errorf("failed to create snapshot directory: %w", err)
	}

	// Use cvd snapshot_take to capture the instance state
	snapshotCmd := fmt.Sprintf(
		"cd %s && HOME=%s cvd snapshot_take --instance_nums=%d --snapshot_path=%s --force 2>&1",
		c.config.CVDHome,
		c.config.CVDHome,
		info.InstanceNum,
		filepath.Join(snapshotDir, "snapshot"),
	)

	output, err := c.runShellScript(ctx, snapshotCmd)
	if err != nil {
		// Cleanup on failure
		_ = c.runCommand(ctx, "rm", "-rf", snapshotDir)
		return nil, fmt.Errorf("failed to take snapshot: %w (output: %s)", err, output)
	}

	// Get the base image this instance was created from
	baseRef := ""
	if info.Metadata != nil {
		baseRef = info.Metadata["snapshot"]
	}

	// Create metadata
	metadata := SnapshotMetadata{
		Name:         snapshotName,
		Type:         SnapshotTypeCustom,
		OrgId:        orgId,
		Description:  description,
		BaseRef:      baseRef,
		CreatedAt:    time.Now(),
		CreatedFrom:  sandboxId,
		Architecture: "x86_64", // TODO: detect from base
		FormFactor:   "phone",  // TODO: detect from base
	}

	// Calculate size
	sizeCmd := fmt.Sprintf("du -sb %s | cut -f1", snapshotDir)
	sizeOutput, err := c.runShellScript(ctx, sizeCmd)
	if err == nil {
		var size int64
		fmt.Sscanf(strings.TrimSpace(sizeOutput), "%d", &size)
		metadata.SizeBytes = size
	}

	// Save metadata
	metadataPath := filepath.Join(snapshotDir, "manifest.json")
	metadataData, _ := json.MarshalIndent(metadata, "", "  ")
	if err := c.writeFile(ctx, metadataPath, metadataData); err != nil {
		log.Warnf("Failed to save snapshot metadata: %v", err)
	}

	log.Infof("Snapshot '%s' created successfully", snapshotPath)

	return &SnapshotInfo{
		Path:     snapshotPath,
		Metadata: metadata,
	}, nil
}

// DeleteSnapshot removes a snapshot
func (c *Client) DeleteSnapshot(ctx context.Context, snapshotPath string) error {
	log.Infof("Deleting snapshot '%s'", snapshotPath)

	// Don't allow deleting base snapshots
	if IsBaseSnapshot(snapshotPath) {
		return fmt.Errorf("cannot delete base snapshots")
	}

	snapshotDir := c.GetSnapshotDir(snapshotPath)

	exists, err := c.fileExists(ctx, snapshotDir)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("snapshot '%s' not found", snapshotPath)
	}

	if err := c.runCommand(ctx, "rm", "-rf", snapshotDir); err != nil {
		return fmt.Errorf("failed to delete snapshot: %w", err)
	}

	log.Infof("Snapshot '%s' deleted", snapshotPath)
	return nil
}

// GetSnapshotInfo returns information about a specific snapshot
func (c *Client) GetSnapshotInfo(ctx context.Context, snapshotPath string) (*SnapshotInfo, error) {
	exists, err := c.SnapshotExists(ctx, snapshotPath)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("snapshot '%s' not found", snapshotPath)
	}

	metadata, err := c.GetSnapshotMetadata(ctx, snapshotPath)
	if err != nil {
		// Create basic metadata
		snapshotType, orgId, name := ParseSnapshotPath(snapshotPath)
		metadata = &SnapshotMetadata{
			Name:  name,
			Type:  snapshotType,
			OrgId: orgId,
		}
	}

	return &SnapshotInfo{
		Path:     snapshotPath,
		Metadata: *metadata,
	}, nil
}

// listSnapshotsInDir lists snapshot directories under a parent directory
func (c *Client) listSnapshotsInDir(ctx context.Context, dir string, prefix string) ([]string, error) {
	entries, err := c.listDirectory(ctx, dir)
	if err != nil {
		return nil, err
	}

	var snapshots []string
	for _, entry := range entries {
		// Each entry is a snapshot name
		snapshotPath := fmt.Sprintf("%s/%s", prefix, entry)
		snapshots = append(snapshots, snapshotPath)
	}

	return snapshots, nil
}

// listDirectory lists directories in a path
func (c *Client) listDirectory(ctx context.Context, path string) ([]string, error) {
	if c.IsRemote() {
		cmd := fmt.Sprintf("ls -1 %s 2>/dev/null || true", path)
		output, err := c.runShellScript(ctx, cmd)
		if err != nil {
			return nil, err
		}

		var entries []string
		for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
			if line != "" {
				entries = append(entries, line)
			}
		}
		return entries, nil
	}

	// Local mode
	dirEntries, err := os.ReadDir(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var entries []string
	for _, entry := range dirEntries {
		if entry.IsDir() {
			entries = append(entries, entry.Name())
		}
	}
	return entries, nil
}

// isDirectory checks if a path is a directory
func (c *Client) isDirectory(ctx context.Context, path string) (bool, error) {
	if c.IsRemote() {
		cmd := fmt.Sprintf("test -d %s && echo yes || echo no", path)
		output, err := c.runShellScript(ctx, cmd)
		if err != nil {
			return false, err
		}
		return strings.TrimSpace(output) == "yes", nil
	}

	// Local mode
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return info.IsDir(), nil
}
