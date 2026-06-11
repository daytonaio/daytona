// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"runtime"
	"testing"
)

func TestParseSandboxPlatformDefaultsToAmd64(t *testing.T) {
	t.Setenv("DAYTONA_RUNNER_PLATFORM", "")

	platform := getSandboxPlatform()
	if platform.String() != "linux/amd64" {
		t.Fatalf("expected linux/amd64, got %s", platform.String())
	}
}

func TestParseSandboxPlatformAllowsNative(t *testing.T) {
	t.Setenv("DAYTONA_RUNNER_PLATFORM", "native")
	platform := getSandboxPlatform()
	expectPlatform := "linux/" + expectedRuntimeSandboxArchitecture()

	if platform.String() != expectPlatform {
		t.Fatalf("expected %s, got %s", expectPlatform, platform.String())
	}
}

func TestParseSandboxPlatformAcceptsArm64(t *testing.T) {
	t.Setenv("DAYTONA_RUNNER_PLATFORM", "arm64")
	platform := getSandboxPlatform()
	if platform.String() != "linux/arm64" {
		t.Fatalf("expected linux/arm64, got %s", platform.String())
	}

	if got := isImageArchSupported("aarch64"); !got {
		t.Fatalf("expected aarch64 image arch to be supported for configured platform %s", platform.String())
	}

	if got := isImageArchSupported("amd64"); got {
		t.Fatalf("expected amd64 image arch not to be supported for configured platform %s", platform.String())
	}
}

func TestImageArchSupportedAcceptsAmd64Aliases(t *testing.T) {
	t.Setenv("DAYTONA_RUNNER_PLATFORM", "amd64")

	for _, imageArch := range []string{"amd64", "x86_64", "x64"} {
		if got := isImageArchSupported(imageArch); !got {
			t.Fatalf("expected %s image arch to be supported for amd64 platform", imageArch)
		}
	}
}

func TestSandboxPlatformFromDockerPlatformPrefersOriginalContainerPlatform(t *testing.T) {
	t.Setenv("DAYTONA_RUNNER_PLATFORM", "arm64")

	platform := sandboxPlatformFromDockerPlatform("linux/amd64")
	if platform.String() != "linux/amd64" {
		t.Fatalf("expected original container platform linux/amd64, got %s", platform.String())
	}

	fallback := sandboxPlatformFromDockerPlatform("")
	if fallback.String() != "linux/arm64" {
		t.Fatalf("expected configured fallback linux/arm64, got %s", fallback.String())
	}
}

func TestParseSandboxPlatformFallbackForInvalidInput(t *testing.T) {
	t.Setenv("DAYTONA_RUNNER_PLATFORM", "not-a-platform")
	platform := getSandboxPlatform()
	if platform.String() != "linux/amd64" {
		t.Fatalf("expected linux/amd64, got %s", platform.String())
	}
}

func expectedRuntimeSandboxArchitecture() string {
	switch runtime.GOARCH {
	case "amd64", "x86_64", "x64":
		return "amd64"
	case "arm64", "aarch64":
		return "arm64"
	default:
		return runtime.GOARCH
	}
}
