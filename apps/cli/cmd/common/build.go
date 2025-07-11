// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

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
