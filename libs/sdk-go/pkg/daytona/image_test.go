// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package daytona

import (
	"strings"
	"testing"

	"github.com/daytonaio/daytona/libs/sdk-go/pkg/options"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBase tests the Base constructor
func TestBase(t *testing.T) {
	tests := []struct {
		name      string
		baseImage string
		expected  string
	}{
		{
			name:      "ubuntu base",
			baseImage: "ubuntu:22.04",
			expected:  "FROM ubuntu:22.04",
		},
		{
			name:      "python base",
			baseImage: "python:3.11-slim",
			expected:  "FROM python:3.11-slim",
		},
		{
			name:      "alpine base",
			baseImage: "alpine:latest",
			expected:  "FROM alpine:latest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img := Base(tt.baseImage)
			require.NotNil(t, img)
			assert.Equal(t, tt.expected, img.Dockerfile())
			assert.Empty(t, img.Contexts())
		})
	}
}

// TestDebianSlim tests the DebianSlim constructor
func TestDebianSlim(t *testing.T) {
	tests := []struct {
		name     string
		version  *string
		expected string
	}{
		{
			name:     "default version",
			version:  nil,
			expected: "FROM python:3.12-slim-bookworm",
		},
		{
			name:     "custom version",
			version:  strPtr("3.10"),
			expected: "FROM python:3.10-slim-bookworm",
		},
		{
			name:     "another version",
			version:  strPtr("3.11"),
			expected: "FROM python:3.11-slim-bookworm",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img := DebianSlim(tt.version)
			require.NotNil(t, img)
			assert.Equal(t, tt.expected, img.Dockerfile())
		})
	}
}

// TestFromDockerfile tests the FromDockerfile constructor
func TestFromDockerfile(t *testing.T) {
	dockerfile := "FROM python:3.11\nRUN pip install numpy\nWORKDIR /app"
	img := FromDockerfile(dockerfile)
	require.NotNil(t, img)
	assert.Equal(t, dockerfile, img.Dockerfile())
}

// TestPipInstall tests the PipInstall method
func TestPipInstall(t *testing.T) {
	tests := []struct {
		name     string
		packages []string
		opts     []func(*options.PipInstall)
		contains []string
	}{
		{
			name:     "basic install",
			packages: []string{"numpy", "pandas"},
			contains: []string{"RUN pip install numpy pandas"},
		},
		{
			name:     "empty packages does nothing",
			packages: []string{},
			contains: []string{}, // should only have FROM line
		},
		{
			name:     "with index URL",
			packages: []string{"torch"},
			opts: []func(*options.PipInstall){
				options.WithIndexURL("https://download.pytorch.org/whl/cpu"),
			},
			contains: []string{"--index-url", "https://download.pytorch.org/whl/cpu", "torch"},
		},
		{
			name:     "with find links",
			packages: []string{"mypackage"},
			opts: []func(*options.PipInstall){
				options.WithFindLinks("/path/to/wheels"),
			},
			contains: []string{"--find-links", "/path/to/wheels"},
		},
		{
			name:     "with pre-release",
			packages: []string{"mypackage"},
			opts: []func(*options.PipInstall){
				options.WithPre(),
			},
			contains: []string{"--pre"},
		},
		{
			name:     "with extra options",
			packages: []string{"mypackage"},
			opts: []func(*options.PipInstall){
				options.WithExtraOptions("--no-cache-dir"),
			},
			contains: []string{"--no-cache-dir"},
		},
		{
			name:     "with extra index URLs",
			packages: []string{"mypackage"},
			opts: []func(*options.PipInstall){
				options.WithExtraIndexURLs("https://private.example.com/simple/"),
			},
			contains: []string{"--extra-index-url", "https://private.example.com/simple/"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img := Base("python:3.11").PipInstall(tt.packages, tt.opts...)
			dockerfile := img.Dockerfile()

			for _, expected := range tt.contains {
				assert.Contains(t, dockerfile, expected)
			}
		})
	}
}

// TestAptGet tests the AptGet method
func TestAptGet(t *testing.T) {
	t.Run("installs packages", func(t *testing.T) {
		img := Base("ubuntu:22.04").AptGet([]string{"git", "curl"})
		dockerfile := img.Dockerfile()
		assert.Contains(t, dockerfile, "apt-get update")
		assert.Contains(t, dockerfile, "apt-get install -y git curl")
		assert.Contains(t, dockerfile, "rm -rf /var/lib/apt/lists/*")
	})

	t.Run("empty packages does nothing", func(t *testing.T) {
		img := Base("ubuntu:22.04").AptGet([]string{})
		lines := strings.Split(img.Dockerfile(), "\n")
		assert.Len(t, lines, 1) // only FROM line
	})
}

// TestImageRun tests the Run method
func TestImageRun(t *testing.T) {
	img := Base("ubuntu:22.04").Run("mkdir -p /app/data")
	assert.Contains(t, img.Dockerfile(), "RUN mkdir -p /app/data")
}

// TestImageEnv tests the Env method
func TestImageEnv(t *testing.T) {
	img := Base("python:3.11").Env("PYTHONUNBUFFERED", "1")
	assert.Contains(t, img.Dockerfile(), "ENV PYTHONUNBUFFERED=1")
}

// TestImageWorkdir tests the Workdir method
func TestImageWorkdir(t *testing.T) {
	img := Base("python:3.11").Workdir("/app")
	assert.Contains(t, img.Dockerfile(), "WORKDIR /app")
}

// TestImageEntrypoint tests the Entrypoint method
func TestImageEntrypoint(t *testing.T) {
	img := Base("python:3.11").Entrypoint([]string{"python", "-m", "myapp"})
	dockerfile := img.Dockerfile()
	assert.Contains(t, dockerfile, `ENTRYPOINT ["python","-m","myapp"]`)
}

// TestImageCmd tests the Cmd method
func TestImageCmd(t *testing.T) {
	img := Base("python:3.11").Cmd([]string{"python", "app.py"})
	assert.Contains(t, img.Dockerfile(), `CMD ["python", "app.py"]`)
}

// TestImageUser tests the User method
func TestImageUser(t *testing.T) {
	img := Base("python:3.11").User("appuser")
	assert.Contains(t, img.Dockerfile(), "USER appuser")
}

// TestImageCopy tests the Copy method
func TestImageCopy(t *testing.T) {
	img := Base("python:3.11").Copy("requirements.txt", "/app/requirements.txt")
	assert.Contains(t, img.Dockerfile(), "COPY requirements.txt /app/requirements.txt")
}

// TestImageAdd tests the Add method
func TestImageAdd(t *testing.T) {
	img := Base("ubuntu:22.04").Add("https://example.com/app.tar.gz", "/app/")
	assert.Contains(t, img.Dockerfile(), "ADD https://example.com/app.tar.gz /app/")
}

// TestImageExpose tests the Expose method
func TestImageExpose(t *testing.T) {
	img := Base("python:3.11").Expose([]int{8080, 8443})
	assert.Contains(t, img.Dockerfile(), "EXPOSE 8080 8443")
}

// TestImageLabel tests the Label method
func TestImageLabel(t *testing.T) {
	img := Base("python:3.11").Label("maintainer", "team@example.com")
	assert.Contains(t, img.Dockerfile(), `LABEL maintainer="team@example.com"`)
}

// TestImageVolume tests the Volume method
func TestImageVolume(t *testing.T) {
	img := Base("python:3.11").Volume([]string{"/data", "/logs"})
	assert.Contains(t, img.Dockerfile(), "VOLUME [/data /logs]")
}

// TestImageChaining tests method chaining
func TestImageChaining(t *testing.T) {
	img := Base("python:3.11").
		AptGet([]string{"git", "curl"}).
		PipInstall([]string{"numpy", "pandas"}).
		Workdir("/app").
		Env("PYTHONUNBUFFERED", "1").
		Run("echo 'hello'").
		Entrypoint([]string{"python"}).
		Cmd([]string{"app.py"})

	dockerfile := img.Dockerfile()
	lines := strings.Split(dockerfile, "\n")
	assert.Equal(t, "FROM python:3.11", lines[0])
	assert.Len(t, lines, 8) // FROM + 7 instructions
}

// TestAddLocalFile tests the AddLocalFile method
func TestAddLocalFile(t *testing.T) {
	img := Base("python:3.11").AddLocalFile("./requirements.txt", "/app/requirements.txt")
	assert.Contains(t, img.Dockerfile(), "COPY ./requirements.txt /app/requirements.txt")

	contexts := img.Contexts()
	require.Len(t, contexts, 1)
	assert.Equal(t, "./requirements.txt", contexts[0].SourcePath)
	assert.Equal(t, "./requirements.txt", contexts[0].ArchivePath)
}

// TestAddLocalDir tests the AddLocalDir method
func TestAddLocalDir(t *testing.T) {
	img := Base("python:3.11").AddLocalDir("./src", "/app/src")
	assert.Contains(t, img.Dockerfile(), "COPY ./src /app/src")

	contexts := img.Contexts()
	require.Len(t, contexts, 1)
	assert.Equal(t, "./src", contexts[0].SourcePath)
}

// TestDockerfileOutput tests the Dockerfile method
func TestDockerfileOutput(t *testing.T) {
	img := Base("python:3.11").Run("echo hello")
	expected := "FROM python:3.11\nRUN echo hello"
	assert.Equal(t, expected, img.Dockerfile())
}

// TestContextsEmpty tests Contexts returns empty when no local files
func TestContextsEmpty(t *testing.T) {
	img := Base("python:3.11")
	assert.Empty(t, img.Contexts())
}

// TestMultipleContexts tests multiple AddLocalFile/Dir calls
func TestMultipleContexts(t *testing.T) {
	img := Base("python:3.11").
		AddLocalFile("./file1.txt", "/app/file1.txt").
		AddLocalDir("./src", "/app/src")

	contexts := img.Contexts()
	assert.Len(t, contexts, 2)
}

func strPtr(s string) *string {
	return &s
}

func TestImageAdditionalEdgeCases(t *testing.T) {
	t.Run("cmd and entrypoint handle empty slices", func(t *testing.T) {
		img := Base("alpine").Entrypoint([]string{}).Cmd([]string{})
		dockerfile := img.Dockerfile()
		assert.Contains(t, dockerfile, `ENTRYPOINT []`)
		assert.True(t, strings.Contains(dockerfile, `CMD [""]`) || strings.Contains(dockerfile, `CMD []`))
	})

	t.Run("volume preserves order", func(t *testing.T) {
		img := Base("alpine").Volume([]string{"/cache", "/data", "/logs"})
		assert.Contains(t, img.Dockerfile(), "VOLUME [/cache /data /logs]")
	})
}

func TestImageContextsAndDockerfileStability(t *testing.T) {
	img := Base("python:3.11").
		AddLocalFile("./requirements.txt", "/app/requirements.txt").
		AddLocalDir("./src", "/app/src").
		Run("python --version")

	contexts := img.Contexts()
	require.Len(t, contexts, 2)
	assert.Equal(t, "./requirements.txt", contexts[0].ArchivePath)
	assert.Equal(t, "./src", contexts[1].ArchivePath)
	assert.Equal(t, strings.Join([]string{
		"FROM python:3.11",
		"COPY ./requirements.txt /app/requirements.txt",
		"COPY ./src /app/src",
		"RUN python --version",
	}, "\n"), img.Dockerfile())
}
