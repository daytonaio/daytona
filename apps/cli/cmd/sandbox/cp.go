// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sandbox

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	apiclient_cli "github.com/daytonaio/daytona/cli/apiclient"
	"github.com/daytonaio/daytona/cli/cmd/common"
	"github.com/daytonaio/daytona/cli/internal/clierr"
	view_common "github.com/daytonaio/daytona/cli/views/common"
	apiclient "github.com/daytonaio/daytona/libs/api-client-go"
	"github.com/spf13/cobra"
)

// cpResult is the structured output of `daytona cp` in --format mode.
type cpResult struct {
	Source           string `json:"source" yaml:"source"`
	Destination      string `json:"destination" yaml:"destination"`
	Direction        string `json:"direction" yaml:"direction"`
	FilesTransferred int    `json:"filesTransferred" yaml:"filesTransferred"`
}

var CpCmd = &cobra.Command{
	Use:   "cp SOURCE DESTINATION",
	Short: "Copy files between the local machine and a sandbox",
	Long: `Copy files or directories between the local filesystem and a sandbox.

Exactly one of SOURCE or DESTINATION must reference a sandbox path using the
<sandbox>:<path> form, where <sandbox> is a sandbox ID or name. Directories
are copied recursively. Copying into an existing directory places the source
basename inside it, and missing parent directories are created.`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		req, err := parseCpArgs(args[0], args[1])
		if err != nil {
			return err
		}

		ctx := context.Background()

		apiClient, err := apiclient_cli.GetApiClient(nil, nil)
		if err != nil {
			return err
		}

		sandbox, res, err := apiClient.SandboxAPI.GetSandbox(ctx, req.sandboxRef).Execute()
		if err != nil {
			return apiclient_cli.HandleErrorResponse(res, err)
		}

		if err := common.RequireStartedState(sandbox); err != nil {
			return err
		}

		var transferred int
		direction := "download"
		if req.upload {
			direction = "upload"
			transferred, err = cpUpload(ctx, apiClient, sandbox.Id, req.sandboxRef, req.localPath, req.remotePath)
		} else {
			transferred, err = cpDownload(ctx, apiClient, sandbox.Id, req.sandboxRef, req.remotePath, req.localPath)
		}
		if err != nil {
			return err
		}

		if common.FormatFlag != "" {
			common.NewFormatter(cpResult{
				Source:           args[0],
				Destination:      args[1],
				Direction:        direction,
				FilesTransferred: transferred,
			}).Print()
			return nil
		}

		view_common.RenderInfoMessageBold(fmt.Sprintf("Copied %d file(s)", transferred))
		return nil
	},
}

// cpRequest describes a validated cp invocation: which sandbox is involved,
// the remote and local paths, and the transfer direction.
type cpRequest struct {
	sandboxRef string
	remotePath string
	localPath  string
	upload     bool
}

// parseCpEndpoint splits a cp argument into its sandbox and path parts. An
// argument is remote iff it contains a ":" whose prefix is longer than one
// character (so Windows drive letters like C:\foo stay local) and is not a
// relative-directory marker ("." or ".."). The path is everything after the
// first ":", so remote paths may themselves contain colons.
func parseCpEndpoint(s string) (sandbox, filePath string, remote bool) {
	idx := strings.Index(s, ":")
	if idx < 0 {
		return "", s, false
	}
	prefix := s[:idx]
	if len(prefix) <= 1 || prefix == ".." {
		return "", s, false
	}
	return prefix, s[idx+1:], true
}

// parseCpArgs validates that exactly one side of the copy references a
// sandbox and resolves the transfer direction. An empty remote path defaults
// to the sandbox working directory (".").
func parseCpArgs(source, destination string) (*cpRequest, error) {
	srcSandbox, srcPath, srcRemote := parseCpEndpoint(source)
	dstSandbox, dstPath, dstRemote := parseCpEndpoint(destination)

	var req *cpRequest
	switch {
	case srcRemote && dstRemote:
		return nil, clierr.New(clierr.CategoryUsage, "copying between two sandboxes is not supported").
			WithHint("exactly one of SOURCE and DESTINATION may use the <sandbox>:<path> form")
	case !srcRemote && !dstRemote:
		return nil, clierr.New(clierr.CategoryUsage, "neither SOURCE nor DESTINATION references a sandbox").
			WithHint("use <sandbox>:<path> for the remote side, e.g. 'daytona cp ./file.txt my-sandbox:/tmp/file.txt'")
	case dstRemote:
		req = &cpRequest{sandboxRef: dstSandbox, remotePath: dstPath, localPath: srcPath, upload: true}
	default:
		req = &cpRequest{sandboxRef: srcSandbox, remotePath: srcPath, localPath: dstPath, upload: false}
	}

	if req.remotePath == "" {
		req.remotePath = "."
	}
	return req, nil
}

// cpUpload copies a local file or directory tree into the sandbox and returns
// the number of files transferred.
func cpUpload(ctx context.Context, apiClient *apiclient.APIClient, sandboxId, sandboxRef, localSrc, remoteDst string) (int, error) {
	srcInfo, err := os.Stat(localSrc)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return 0, clierr.Newf(clierr.CategoryUsage, "local source %q does not exist", localSrc)
		}
		return 0, err
	}

	// Mirror cp semantics: copying into an existing remote directory places
	// the source basename inside it. The info call is best-effort — an error
	// just means the destination does not exist yet.
	if remoteInfo, _, infoErr := apiClient.ToolboxAPI.GetFileInfoDeprecated(ctx, sandboxId).Path(remoteDst).Execute(); infoErr == nil && remoteInfo != nil && remoteInfo.IsDir {
		remoteDst = path.Join(remoteDst, filepath.Base(localSrc))
	}

	if !srcInfo.IsDir() {
		if err := cpUploadFile(ctx, apiClient, sandboxId, localSrc, remoteDst); err != nil {
			return 0, err
		}
		cpReportTransfer("Uploaded", localSrc, sandboxRef+":"+remoteDst)
		return 1, nil
	}

	count := 0
	err = filepath.WalkDir(localSrc, func(p string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		rel, relErr := filepath.Rel(localSrc, p)
		if relErr != nil {
			return relErr
		}
		remotePath := remoteDst
		if rel != "." {
			remotePath = path.Join(remoteDst, filepath.ToSlash(rel))
		}

		if d.IsDir() {
			// Best-effort mirror of the directory itself so empty directories
			// are preserved; per-file uploads create missing parents anyway.
			_, _ = apiClient.ToolboxAPI.CreateFolderDeprecated(ctx, sandboxId).Path(remotePath).Mode("0755").Execute()
			return nil
		}

		if err := cpUploadFile(ctx, apiClient, sandboxId, p, remotePath); err != nil {
			return err
		}
		count++
		cpReportTransfer("Uploaded", p, sandboxRef+":"+remotePath)
		return nil
	})
	return count, err
}

// cpUploadFile uploads a single local file to remotePath in the sandbox,
// creating the remote parent directory best-effort first.
func cpUploadFile(ctx context.Context, apiClient *apiclient.APIClient, sandboxId, localPath, remotePath string) error {
	file, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	if parent := path.Dir(remotePath); parent != "" && parent != "." && parent != "/" {
		// Best-effort: if creation fails the upload reports the real error.
		_, _ = apiClient.ToolboxAPI.CreateFolderDeprecated(ctx, sandboxId).Path(parent).Mode("0755").Execute()
	}

	res, err := apiClient.ToolboxAPI.UploadFileDeprecated(ctx, sandboxId).Path(remotePath).File(file).Execute()
	if err != nil {
		return apiclient_cli.HandleErrorResponse(res, err)
	}
	return nil
}

// cpDownload copies a sandbox file or directory tree to the local filesystem
// and returns the number of files transferred.
func cpDownload(ctx context.Context, apiClient *apiclient.APIClient, sandboxId, sandboxRef, remoteSrc, localDst string) (int, error) {
	info, res, err := apiClient.ToolboxAPI.GetFileInfoDeprecated(ctx, sandboxId).Path(remoteSrc).Execute()
	if err != nil {
		return 0, apiclient_cli.HandleErrorResponse(res, err)
	}

	// Mirror cp semantics: copying into an existing local directory places
	// the source basename inside it.
	if dstInfo, statErr := os.Stat(localDst); statErr == nil && dstInfo.IsDir() {
		localDst = filepath.Join(localDst, path.Base(remoteSrc))
	}

	if info.IsDir {
		return cpDownloadDir(ctx, apiClient, sandboxId, sandboxRef, remoteSrc, localDst)
	}

	if err := cpDownloadFile(ctx, apiClient, sandboxId, remoteSrc, localDst); err != nil {
		return 0, err
	}
	cpReportTransfer("Downloaded", sandboxRef+":"+remoteSrc, localDst)
	return 1, nil
}

// cpDownloadDir recursively downloads a sandbox directory into localDir.
// FileInfo.Name returned by the toolbox is a basename, so child remote paths
// are built with the POSIX path package against the parent directory.
func cpDownloadDir(ctx context.Context, apiClient *apiclient.APIClient, sandboxId, sandboxRef, remoteDir, localDir string) (int, error) {
	if err := os.MkdirAll(localDir, 0o755); err != nil {
		return 0, err
	}

	entries, res, err := apiClient.ToolboxAPI.ListFilesDeprecated(ctx, sandboxId).Path(remoteDir).Execute()
	if err != nil {
		return 0, apiclient_cli.HandleErrorResponse(res, err)
	}

	count := 0
	for _, entry := range entries {
		remotePath := path.Join(remoteDir, entry.Name)
		localPath := filepath.Join(localDir, entry.Name)

		if entry.IsDir {
			n, err := cpDownloadDir(ctx, apiClient, sandboxId, sandboxRef, remotePath, localPath)
			count += n
			if err != nil {
				return count, err
			}
			continue
		}

		if err := cpDownloadFile(ctx, apiClient, sandboxId, remotePath, localPath); err != nil {
			return count, err
		}
		count++
		cpReportTransfer("Downloaded", sandboxRef+":"+remotePath, localPath)
	}
	return count, nil
}

// cpDownloadFile downloads a single sandbox file to localPath.
func cpDownloadFile(ctx context.Context, apiClient *apiclient.APIClient, sandboxId, remotePath, localPath string) error {
	tmpFile, res, err := apiClient.ToolboxAPI.DownloadFileDeprecated(ctx, sandboxId).Path(remotePath).Execute()
	if err != nil {
		return apiclient_cli.HandleErrorResponse(res, err)
	}
	return cpWriteLocalFile(tmpFile, localPath)
}

// cpWriteLocalFile copies the downloaded temp file to localPath, creating
// parent directories as needed, then closes and removes the temp file.
func cpWriteLocalFile(tmp *os.File, localPath string) (err error) {
	defer func() { _ = os.Remove(tmp.Name()) }()
	defer func() {
		if closeErr := tmp.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	if dir := filepath.Dir(localPath); dir != "" {
		if mkErr := os.MkdirAll(dir, 0o755); mkErr != nil {
			return mkErr
		}
	}

	dst, err := os.Create(localPath)
	if err != nil {
		return err
	}
	if _, copyErr := io.Copy(dst, tmp); copyErr != nil {
		_ = dst.Close()
		return copyErr
	}
	return dst.Close()
}

// cpReportTransfer prints the per-file human progress line; suppressed in
// --format mode where only the structured summary is emitted.
func cpReportTransfer(verb, source, destination string) {
	if common.FormatFlag != "" {
		return
	}
	view_common.RenderInfoMessage(fmt.Sprintf("%s %s -> %s", verb, source, destination))
}

func init() {
	common.RegisterFormatFlag(CpCmd)
}
