// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package daytona

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/daytonaio/daytona/libs/sdk-go/pkg/options"
)

// DockerImage provides a fluent interface for building Docker images declaratively.
//
// DockerImage allows you to define Docker images using Go code instead of Dockerfiles.
// Methods can be chained to build up the image definition, which is then converted
// to a Dockerfile when used with [SnapshotService.Create].
//
// Example:
//
//	// Create a Python image with dependencies
//	image := daytona.Base("python:3.11-slim").
//	    AptGet([]string{"git", "curl"}).
//	    PipInstall([]string{"numpy", "pandas"}).
//	    Workdir("/app").
//	    Env("PYTHONUNBUFFERED", "1")
//
//	// Use with snapshot creation
//	snapshot, logChan, err := client.Snapshots.Create(ctx, &types.CreateSnapshotParams{
//	    Name:  "my-python-env",
//	    DockerImage: image,
//	})
type DockerImage struct {
	instructions []string
	contexts     []DockerImageContext
}

// DockerImageContext represents a local file or directory to include in the build context.
//
// When using [DockerImage.AddLocalFile] or [DockerImage.AddLocalDir], the file/directory is
// uploaded to object storage and included in the Docker build context.
type DockerImageContext struct {
	SourcePath  string // Local path to the file or directory
	ArchivePath string // Path within the build context archive
}

// Base creates a new Image from a base Docker image.
//
// This is typically the starting point for building an image definition.
// The baseImage parameter is any valid Docker image reference.
//
// Example:
//
//	image := daytona.Base("ubuntu:22.04")
//	image := daytona.Base("python:3.11-slim")
//	image := daytona.Base("node:18-alpine")
func Base(baseImage string) *DockerImage {
	return &DockerImage{
		instructions: []string{fmt.Sprintf("FROM %s", baseImage)},
	}
}

// DebianSlim creates a Python image based on Debian slim.
//
// This is a convenience function for creating Python environments.
// If pythonVersion is nil, defaults to Python 3.12.
//
// Example:
//
//	// Use default Python 3.12
//	image := daytona.DebianSlim(nil)
//
//	// Use specific version
//	version := "3.10"
//	image := daytona.DebianSlim(&version)
func DebianSlim(pythonVersion *string) *DockerImage {
	version := "3.12"
	if pythonVersion != nil {
		version = *pythonVersion
	}
	return Base(fmt.Sprintf("python:%s-slim-bookworm", version))
}

// FromDockerfile creates an Image from an existing Dockerfile string.
//
// Use this when you have an existing Dockerfile you want to use.
//
// Example:
//
//	dockerfile := `FROM python:3.11
//	RUN pip install numpy
//	WORKDIR /app`
//	image := daytona.FromDockerfile(dockerfile)
func FromDockerfile(dockerfile string) *DockerImage {
	return &DockerImage{
		instructions: strings.Split(dockerfile, "\n"),
	}
}

// PipInstall adds a pip install instruction for Python packages.
//
// Optional parameters can be configured using functional options:
//   - [options.WithFindLinks]: Add find-links URLs
//   - [options.WithIndexURL]: Set custom PyPI index
//   - [options.WithExtraIndexURLs]: Add extra index URLs
//   - [options.WithPre]: Allow pre-release versions
//   - [options.WithExtraOptions]: Add additional pip options
//
// Example:
//
//	// Basic installation
//	image := daytona.Base("python:3.11").PipInstall([]string{"numpy", "pandas"})
//
//	// With options
//	image := daytona.Base("python:3.11").PipInstall(
//	    []string{"torch"},
//	    options.WithIndexURL("https://download.pytorch.org/whl/cpu"),
//	    options.WithExtraOptions("--no-cache-dir"),
//	)
func (img *DockerImage) PipInstall(packages []string, opts ...func(*options.PipInstall)) *DockerImage {
	if len(packages) == 0 {
		return img
	}

	pipOpts := options.Apply(opts...)

	cmd := []string{"pip", "install"}

	if len(pipOpts.FindLinks) > 0 {
		for _, link := range pipOpts.FindLinks {
			cmd = append(cmd, "--find-links", link)
		}
	}
	if pipOpts.IndexURL != "" {
		cmd = append(cmd, "--index-url", pipOpts.IndexURL)
	}
	if len(pipOpts.ExtraIndexURLs) > 0 {
		for _, url := range pipOpts.ExtraIndexURLs {
			cmd = append(cmd, "--extra-index-url", url)
		}
	}
	if pipOpts.Pre {
		cmd = append(cmd, "--pre")
	}
	if pipOpts.ExtraOptions != "" {
		cmd = append(cmd, pipOpts.ExtraOptions)
	}

	cmd = append(cmd, packages...)

	img.instructions = append(img.instructions, fmt.Sprintf("RUN %s", strings.Join(cmd, " ")))
	return img
}

// AptGet adds an apt-get install instruction for system packages.
//
// This automatically handles updating the package list and cleaning up
// afterward to minimize image size.
//
// Example:
//
//	image := daytona.Base("ubuntu:22.04").AptGet([]string{"git", "curl", "build-essential"})
func (img *DockerImage) AptGet(packages []string) *DockerImage {
	if len(packages) == 0 {
		return img
	}

	cmd := fmt.Sprintf("apt-get update && apt-get install -y %s && rm -rf /var/lib/apt/lists/*", strings.Join(packages, " "))
	img.instructions = append(img.instructions, fmt.Sprintf("RUN %s", cmd))
	return img
}

// Run adds a RUN instruction to execute a shell command.
//
// Example:
//
//	image := daytona.Base("ubuntu:22.04").
//	    Run("mkdir -p /app/data").
//	    Run("chmod 755 /app")
func (img *DockerImage) Run(command string) *DockerImage {
	img.instructions = append(img.instructions, fmt.Sprintf("RUN %s", command))
	return img
}

// Env sets an environment variable in the image.
//
// Example:
//
//	image := daytona.Base("python:3.11").
//	    Env("PYTHONUNBUFFERED", "1").
//	    Env("APP_ENV", "production")
func (img *DockerImage) Env(key, value string) *DockerImage {
	img.instructions = append(img.instructions, fmt.Sprintf("ENV %s=%s", key, value))
	return img
}

// Workdir sets the working directory for subsequent instructions.
//
// Example:
//
//	image := daytona.Base("python:3.11").
//	    Workdir("/app").
//	    Run("pip install -r requirements.txt")
func (img *DockerImage) Workdir(path string) *DockerImage {
	img.instructions = append(img.instructions, fmt.Sprintf("WORKDIR %s", path))
	return img
}

// Entrypoint sets the entrypoint for the image.
//
// The cmd parameter is the command and arguments as a slice.
//
// Example:
//
//	image := daytona.Base("python:3.11").
//	    Entrypoint([]string{"python", "-m", "myapp"})
func (img *DockerImage) Entrypoint(cmd []string) *DockerImage {
	jsonCmd, _ := json.Marshal(cmd)
	img.instructions = append(img.instructions, fmt.Sprintf("ENTRYPOINT %s", jsonCmd))
	return img
}

// Cmd sets the default command for the image.
//
// If an entrypoint is set, the cmd provides default arguments to it.
//
// Example:
//
//	image := daytona.Base("python:3.11").
//	    Cmd([]string{"python", "app.py"})
func (img *DockerImage) Cmd(cmd []string) *DockerImage {
	cmdStr := strings.Join(cmd, "\", \"")
	img.instructions = append(img.instructions, fmt.Sprintf("CMD [\"%s\"]", cmdStr))
	return img
}

// User sets the user for subsequent instructions and container runtime.
//
// Example:
//
//	image := daytona.Base("python:3.11").
//	    Run("useradd -m appuser").
//	    User("appuser").
//	    Workdir("/home/appuser")
func (img *DockerImage) User(username string) *DockerImage {
	img.instructions = append(img.instructions, fmt.Sprintf("USER %s", username))
	return img
}

// Copy adds a COPY instruction to copy files into the image.
//
// For local files, use [DockerImage.AddLocalFile] instead, which handles uploading
// to the build context.
//
// Example:
//
//	image := daytona.Base("python:3.11").
//	    Copy("requirements.txt", "/app/requirements.txt")
func (img *DockerImage) Copy(source, destination string) *DockerImage {
	img.instructions = append(img.instructions, fmt.Sprintf("COPY %s %s", source, destination))
	return img
}

// Add adds an ADD instruction to the image.
//
// ADD supports URLs and automatic tar extraction. For simple file copying,
// prefer [DockerImage.Copy].
//
// Example:
//
//	image := daytona.Base("ubuntu:22.04").
//	    Add("https://example.com/app.tar.gz", "/app/")
func (img *DockerImage) Add(source, destination string) *DockerImage {
	img.instructions = append(img.instructions, fmt.Sprintf("ADD %s %s", source, destination))
	return img
}

// Expose declares ports that the container listens on.
//
// This is documentation for users and tools; it doesn't actually publish ports.
//
// Example:
//
//	image := daytona.Base("python:3.11").
//	    Expose([]int{8080, 8443})
func (img *DockerImage) Expose(ports []int) *DockerImage {
	portStrs := make([]string, len(ports))
	for i, port := range ports {
		portStrs[i] = fmt.Sprintf("%d", port)
	}
	img.instructions = append(img.instructions, fmt.Sprintf("EXPOSE %s", strings.Join(portStrs, " ")))
	return img
}

// Label adds metadata to the image.
//
// Example:
//
//	image := daytona.Base("python:3.11").
//	    Label("maintainer", "team@example.com").
//	    Label("version", "1.0.0")
func (img *DockerImage) Label(key, value string) *DockerImage {
	img.instructions = append(img.instructions, fmt.Sprintf("LABEL %s=\"%s\"", key, value))
	return img
}

// Volume declares mount points for the container.
//
// Example:
//
//	image := daytona.Base("python:3.11").
//	    Volume([]string{"/data", "/logs"})
func (img *DockerImage) Volume(paths []string) *DockerImage {
	pathsStr := strings.Join(paths, " ")
	img.instructions = append(img.instructions, fmt.Sprintf("VOLUME [%s]", pathsStr))
	return img
}

// AddLocalFile adds a local file to the build context and copies it to the image.
//
// The file is uploaded to object storage and included in the Docker build context.
//
// Example:
//
//	image := daytona.Base("python:3.11").
//	    AddLocalFile("./requirements.txt", "/app/requirements.txt").
//	    Run("pip install -r /app/requirements.txt")
func (img *DockerImage) AddLocalFile(localPath, remotePath string) *DockerImage {
	img.contexts = append(img.contexts, DockerImageContext{
		SourcePath:  localPath,
		ArchivePath: localPath,
	})
	img.instructions = append(img.instructions, fmt.Sprintf("COPY %s %s", localPath, remotePath))
	return img
}

// AddLocalDir adds a local directory to the build context and copies it to the image.
//
// The directory is uploaded to object storage and included in the Docker build context.
//
// Example:
//
//	image := daytona.Base("python:3.11").
//	    AddLocalDir("./src", "/app/src")
func (img *DockerImage) AddLocalDir(localPath, remotePath string) *DockerImage {
	img.contexts = append(img.contexts, DockerImageContext{
		SourcePath:  localPath,
		ArchivePath: localPath,
	})
	img.instructions = append(img.instructions, fmt.Sprintf("COPY %s %s", localPath, remotePath))
	return img
}

// Dockerfile returns the generated Dockerfile content.
//
// This is called internally when creating snapshots.
//
// Example:
//
//	image := daytona.Base("python:3.11").PipInstall([]string{"numpy"})
//	fmt.Println(image.Dockerfile())
//	// Output:
//	// FROM python:3.11
//	// RUN pip install numpy
func (img *DockerImage) Dockerfile() string {
	return strings.Join(img.instructions, "\n")
}

// Contexts returns the build contexts for local files/directories.
//
// This is called internally when creating snapshots to upload local files.
func (img *DockerImage) Contexts() []DockerImageContext {
	return img.contexts
}
