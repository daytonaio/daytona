// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"fmt"
	"strings"

	"github.com/daytonaio/daytona/cli/internal/clierr"
	apiclient "github.com/daytonaio/daytona/libs/api-client-go"
	"github.com/joho/godotenv"
)

// ParseKeyValuePairs parses KEY=VALUE entries collected from the given flag.
// Every entry must contain '=' and a non-empty key.
func ParseKeyValuePairs(entries []string, flag string) (map[string]string, error) {
	pairs := make(map[string]string, len(entries))
	for _, entry := range entries {
		key, value, found := strings.Cut(entry, "=")
		if !found || key == "" {
			return nil, clierr.Newf(clierr.CategoryUsage, "invalid --%s value %q: expected KEY=VALUE", flag, entry)
		}
		pairs[key] = value
	}
	return pairs, nil
}

// ParseVolumeSpecs parses VOLUME_ID:MOUNT_PATH entries. Both the volume ID
// and the mount path must be non-empty.
func ParseVolumeSpecs(entries []string) ([]apiclient.SandboxVolume, error) {
	volumes := make([]apiclient.SandboxVolume, 0, len(entries))
	for _, entry := range entries {
		volumeId, mountPath, found := strings.Cut(entry, ":")
		if !found || volumeId == "" || mountPath == "" {
			return nil, clierr.Newf(clierr.CategoryUsage, "invalid --volume value %q: expected VOLUME_ID:MOUNT_PATH", entry)
		}
		volumes = append(volumes, apiclient.SandboxVolume{
			VolumeId:  volumeId,
			MountPath: mountPath,
		})
	}
	return volumes, nil
}

// ReadKeyValueFile reads KEY=VALUE pairs from a dotenv-style file.
func ReadKeyValueFile(path string) (map[string]string, error) {
	values, err := godotenv.Read(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}
	return values, nil
}

// ResolveKeyValuePairs merges KEY=VALUE pairs from an optional file with
// entries given on the command line; CLI entries override file values per key.
func ResolveKeyValuePairs(entries []string, filePath, flag string) (map[string]string, error) {
	result := map[string]string{}
	if filePath != "" {
		fileValues, err := ReadKeyValueFile(filePath)
		if err != nil {
			return nil, err
		}
		result = fileValues
	}
	cliValues, err := ParseKeyValuePairs(entries, flag)
	if err != nil {
		return nil, err
	}
	for key, value := range cliValues {
		result[key] = value
	}
	return result, nil
}
