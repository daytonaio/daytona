// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"fmt"
	"os"
	"runtime"
	"strings"
)

const (
	runnerPlatformEnvVar         = "DAYTONA_RUNNER_PLATFORM"
	defaultSandboxPlatform       = "linux/amd64"
	nativeSandboxPlatformDefault = "native"
)

type sandboxPlatform struct {
	os           string
	architecture string
}

func getSandboxPlatform() sandboxPlatform {
	return parseSandboxPlatform(os.Getenv(runnerPlatformEnvVar))
}

func parseSandboxPlatform(raw string) sandboxPlatform {
	if strings.TrimSpace(raw) == "" {
		return parseSandboxPlatform(defaultSandboxPlatform)
	}

	platform := strings.ToLower(strings.TrimSpace(raw))
	if platform == nativeSandboxPlatformDefault {
		return sandboxPlatform{
			os:           "linux",
			architecture: runtimeArchitecture(runtime.GOARCH),
		}
	}

	if strings.HasPrefix(platform, "linux/") {
		parts := strings.Split(platform, "/")
		if len(parts) == 2 {
			normalizedArch, ok := normalizeArchitecture(parts[1])
			if !ok {
				return parseSandboxPlatform(defaultSandboxPlatform)
			}
			return sandboxPlatform{
				os:           parts[0],
				architecture: normalizedArch,
			}
		}

		return parseSandboxPlatform(defaultSandboxPlatform)
	}

	normalizedArch, ok := normalizeArchitecture(platform)
	if !ok {
		return parseSandboxPlatform(defaultSandboxPlatform)
	}

	// Keep backwards compatibility with arch-only inputs.
	return sandboxPlatform{
		os:           "linux",
		architecture: normalizedArch,
	}
}

func normalizeArchitecture(rawArch string) (string, bool) {
	rawArch = strings.ToLower(strings.TrimSpace(rawArch))

	switch strings.ToLower(strings.TrimSpace(rawArch)) {
	case "x86_64", "x64":
		return "amd64", true
	case "aarch64":
		return "arm64", true
	case "amd64":
		return "amd64", true
	case "arm64":
		return "arm64", true
	default:
		return rawArch, false
	}
}

func runtimeArchitecture(runtimeArch string) string {
	arch, _ := normalizeArchitecture(runtimeArch)
	return arch
}

func getSandboxContainerArchs() []string {
	platform := getSandboxPlatform()
	switch platform.architecture {
	case "arm64":
		return []string{"arm64", "aarch64"}
	case "amd64":
		return []string{"amd64", "x86_64"}
	default:
		return []string{platform.architecture}
	}
}

func sandboxImagePlatform() string {
	platform := getSandboxPlatform()
	return fmt.Sprintf("%s/%s", platform.os, platform.architecture)
}

func (p sandboxPlatform) String() string {
	return fmt.Sprintf("%s/%s", p.os, p.architecture)
}

func isImageArchSupported(imageArch string) bool {
	normalizedImageArch := strings.ToLower(strings.TrimSpace(imageArch))
	for _, supported := range getSandboxContainerArchs() {
		if normalizedImageArch == supported {
			return true
		}
	}
	return false
}
