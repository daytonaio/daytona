// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/daytonaio/apiclient"
	apiclient_cli "github.com/daytonaio/daytona/cli/apiclient"
	"github.com/daytonaio/daytona/cli/pkg/minio"
)

// Create MinIO client from access parameters
func CreateMinioClient(accessParams *apiclient.StorageAccessDto) (*minio.Client, error) {
	storageURL, err := url.Parse(accessParams.StorageUrl)
	if err != nil {
		return nil, fmt.Errorf("invalid storage URL: %w", err)
	}

	minioClient, err := minio.NewClient(
		storageURL.Host,
		accessParams.AccessKey,
		accessParams.Secret,
		accessParams.Bucket,
		storageURL.Scheme == "https",
		accessParams.SessionToken,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage client: %w", err)
	}

	return minioClient, nil
}

// List existing objects in MinIO
func ListExistingObjects(ctx context.Context, minioClient *minio.Client, orgID string) (map[string]bool, error) {
	objects, err := minioClient.ListObjects(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to list objects: %w", err)
	}

	existingObjects := make(map[string]bool)
	for _, obj := range objects {
		existingObjects[obj] = true
	}
	return existingObjects, nil
}

// getContextHashes processes context paths and returns their hashes
func getContextHashes(ctx context.Context, apiClient *apiclient.APIClient, contextPaths []string) ([]string, error) {
	contextHashes := []string{}
	if len(contextPaths) == 0 {
		return contextHashes, nil
	}

	// Get storage access parameters
	accessParams, res, err := apiClient.ObjectStorageAPI.GetPushAccess(ctx).Execute()
	if err != nil {
		return nil, apiclient_cli.HandleErrorResponse(res, err)
	}

	// Create MinIO client
	minioClient, err := CreateMinioClient(accessParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage client: %w", err)
	}

	// List existing objects to avoid re-uploading
	existingObjects, err := ListExistingObjects(ctx, minioClient, accessParams.OrganizationId)
	if err != nil {
		return nil, fmt.Errorf("failed to list existing objects: %w", err)
	}

	// Process each context path
	for _, contextPath := range contextPaths {
		absPath, err := filepath.Abs(contextPath)
		if err != nil {
			return nil, fmt.Errorf("invalid context path %s: %w", contextPath, err)
		}

		fileInfo, err := os.Stat(absPath)
		if err != nil {
			return nil, fmt.Errorf("failed to access context path %s: %w", contextPath, err)
		}

		if fileInfo.IsDir() {
			// Process directory
			dirHashes, err := minioClient.ProcessDirectory(ctx, absPath, accessParams.OrganizationId, existingObjects)
			if err != nil {
				return nil, fmt.Errorf("failed to process directory %s: %w", absPath, err)
			}
			contextHashes = append(contextHashes, dirHashes...)
		} else {
			// Process single file
			hash, err := minioClient.ProcessFile(ctx, absPath, accessParams.OrganizationId, existingObjects)
			if err != nil {
				return nil, fmt.Errorf("failed to process file %s: %w", absPath, err)
			}
			contextHashes = append(contextHashes, hash)
		}
	}

	return contextHashes, nil
}

func parseDockerfileForSources(dockerfileContent string, dockerfileDir string) ([]string, error) {
	var sources []string
	lines := strings.Split(dockerfileContent, "\n")

	copyRegex := regexp.MustCompile(`^\s*COPY\s+(.+)`)
	addRegex := regexp.MustCompile(`^\s*ADD\s+(.+)`)

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		var matches []string
		if copyRegex.MatchString(line) {
			// Skip COPY commands with --from= flag (multi-stage builds)
			if !strings.Contains(line, "--from=") {
				matches = copyRegex.FindStringSubmatch(line)
			}
		} else if addRegex.MatchString(line) {
			matches = addRegex.FindStringSubmatch(line)
		}

		if len(matches) > 1 {
			sourcePaths := parseCopyAddCommand(matches[1])
			for _, srcPath := range sourcePaths {
				// Skip if it's a URL (ADD command can use URLs)
				if strings.HasPrefix(srcPath, "http://") || strings.HasPrefix(srcPath, "https://") {
					continue
				}

				// Convert relative paths to absolute paths relative to Dockerfile directory
				if !filepath.IsAbs(srcPath) {
					srcPath = filepath.Join(dockerfileDir, srcPath)
				}

				srcPath = filepath.Clean(srcPath)

				// Check if path exists and add to sources
				if _, err := os.Stat(srcPath); err == nil {
					sources = append(sources, srcPath)
				} else {
					// If exact path doesn't exist, try to match glob patterns
					matches, err := filepath.Glob(srcPath)
					if err == nil && len(matches) > 0 {
						sources = append(sources, matches...)
					}
				}
			}
		}
	}

	// Remove duplicates and optimize paths
	sourceMap := make(map[string]bool)
	var uniqueSources []string

	// Check if we have the current directory (.) in our sources
	hasCurrentDir := false
	currentDirPath := dockerfileDir

	for _, src := range sources {
		if src == currentDirPath {
			hasCurrentDir = true
			break
		}
	}

	// If we have the current directory, we only need that (it includes everything)
	if hasCurrentDir {
		return []string{currentDirPath}, nil
	}

	// Otherwise, remove duplicates normally
	for _, src := range sources {
		if !sourceMap[src] {
			sourceMap[src] = true
			uniqueSources = append(uniqueSources, src)
		}
	}

	return uniqueSources, nil
}

func parseCopyAddCommand(args string) []string {
	args = strings.TrimSpace(args)
	var sources []string

	// Handle JSON array format: ["src1", "src2", "dest"]
	if strings.HasPrefix(args, "[") && strings.HasSuffix(args, "]") {
		// Remove brackets and parse as space-separated values with quotes
		content := strings.Trim(args, "[]")
		parts := parseQuotedArguments(content)
		if len(parts) >= 2 {
			// All but the last argument are sources
			sources = parts[:len(parts)-1]
		}
		return sources
	}

	// Handle regular format with possible flags
	parts := parseQuotedArguments(args)

	// Skip flags like --chown, --chmod, --from
	sourcesStartIdx := 0
	for i := 0; i < len(parts); i++ {
		part := parts[i]
		if strings.HasPrefix(part, "--") {
			// Skip the flag and its value if it has one
			if !strings.Contains(part, "=") && i+1 < len(parts) && !strings.HasPrefix(parts[i+1], "--") {
				sourcesStartIdx = i + 2
			} else {
				sourcesStartIdx = i + 1
			}
		} else {
			break
		}
	}

	// After skipping flags, we need at least one source and one destination
	if len(parts)-sourcesStartIdx >= 2 {
		sources = parts[sourcesStartIdx : len(parts)-1]
	}

	return sources
}

func parseQuotedArguments(input string) []string {
	var args []string
	var current strings.Builder
	inQuotes := false
	quoteChar := byte(0)

	input = strings.TrimSpace(input)

	for i := 0; i < len(input); i++ {
		char := input[i]

		if !inQuotes && (char == '"' || char == '\'') {
			inQuotes = true
			quoteChar = char
		} else if inQuotes && char == quoteChar {
			inQuotes = false
			quoteChar = 0
		} else if !inQuotes && (char == ' ' || char == '\t') {
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
			// Skip consecutive whitespace
			for i+1 < len(input) && (input[i+1] == ' ' || input[i+1] == '\t') {
				i++
			}
		} else {
			current.WriteByte(char)
		}
	}

	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args
}

func GetCreateBuildInfoDto(ctx context.Context, dockerfilePath string, contextPaths []string) (*apiclient.CreateBuildInfo, error) {
	dockerfileAbsPath, err := filepath.Abs(dockerfilePath)
	if err != nil {
		return nil, fmt.Errorf("invalid dockerfile path: %w", err)
	}

	if _, err := os.Stat(dockerfileAbsPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("dockerfile does not exist: %s", dockerfileAbsPath)
	}

	dockerfileContent, err := os.ReadFile(dockerfileAbsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read dockerfile: %w", err)
	}

	// If no context paths are provided, automatically parse the Dockerfile to find them
	if len(contextPaths) == 0 {
		dockerfileDir := filepath.Dir(dockerfileAbsPath)
		autoContextPaths, err := parseDockerfileForSources(string(dockerfileContent), dockerfileDir)
		if err != nil {
			return nil, fmt.Errorf("failed to parse dockerfile for context: %w", err)
		}
		contextPaths = autoContextPaths
	}

	apiClient, err := apiclient_cli.GetApiClient(nil, nil)
	if err != nil {
		return nil, err
	}

	contextHashes, err := getContextHashes(ctx, apiClient, contextPaths)
	if err != nil {
		return nil, err
	}

	return &apiclient.CreateBuildInfo{
		DockerfileContent: string(dockerfileContent),
		ContextHashes:     contextHashes,
	}, nil
}
