// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package daytona

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/daytonaio/daytona/libs/sdk-go/pkg/errors"
	"github.com/daytonaio/daytona/libs/sdk-go/pkg/options"
	"github.com/daytonaio/daytona/libs/sdk-go/pkg/types"
	"github.com/daytonaio/daytona/libs/toolbox-api-client-go"
)

// FileSystemService provides file system operations for a sandbox.
//
// FileSystemService enables file and directory management including creating,
// reading, writing, moving, and deleting files. It also supports file searching
// and permission management. Access through [Sandbox.FileSystem].
//
// Example:
//
//	// List files in a directory
//	files, err := sandbox.FileSystem.ListFiles(ctx, "/home/user")
//
//	// Create a directory
//	err = sandbox.FileSystem.CreateFolder(ctx, "/home/user/mydir")
//
//	// Upload a file
//	err = sandbox.FileSystem.UploadFile(ctx, "/local/path/file.txt", "/home/user/file.txt")
//
//	// Download a file
//	data, err := sandbox.FileSystem.DownloadFile(ctx, "/home/user/file.txt", nil)
type FileSystemService struct {
	toolboxClient *toolbox.APIClient
	otel          *otelState
}

// NewFileSystemService creates a new FileSystemService with the provided toolbox client.
//
// This is typically called internally by the SDK when creating a [Sandbox].
// Users should access FileSystemService through [Sandbox.FileSystem] rather than
// creating it directly.
func NewFileSystemService(toolboxClient *toolbox.APIClient, otel *otelState) *FileSystemService {
	return &FileSystemService{
		toolboxClient: toolboxClient,
		otel:          otel,
	}
}

// CreateFolder creates a directory at the specified path.
//
// The path parameter specifies the absolute path for the new directory.
// Parent directories are created automatically if they don't exist.
//
// Optional parameters can be configured using functional options:
//   - [options.WithMode]: Set Unix file permissions (defaults to "0755")
//
// Example:
//
//	// Create with default permissions
//	err := sandbox.FileSystem.CreateFolder(ctx, "/home/user/mydir")
//
//	// Create with custom permissions
//	err := sandbox.FileSystem.CreateFolder(ctx, "/home/user/private",
//	    options.WithMode("0700"),
//	)
//
// Returns an error if the directory creation fails.
func (f *FileSystemService) CreateFolder(ctx context.Context, path string, opts ...func(*options.CreateFolder)) error {
	return withInstrumentationVoid(ctx, f.otel, "FileSystem", "CreateFolder", func(ctx context.Context) error {
		folderOpts := options.Apply(opts...)

		req := f.toolboxClient.FileSystemAPI.CreateFolder(ctx).Path(path)
		if folderOpts.Mode != nil {
			req = req.Mode(*folderOpts.Mode)
		} else {
			req = req.Mode("0755")
		}

		httpResp, err := req.Execute()
		if err != nil {
			return errors.ConvertToolboxError(err, httpResp)
		}

		return nil
	})
}

// ListFiles lists files and directories in the specified path.
//
// The path parameter specifies the directory to list.
//
// Returns a slice of [types.FileInfo] containing metadata for each file and directory,
// including name, size, permissions, modification time, and whether it's a directory.
//
// Example:
//
//	files, err := sandbox.FileSystem.ListFiles(ctx, "/home/user")
//	if err != nil {
//	    return err
//	}
//	for _, file := range files {
//	    if file.IsDirectory {
//	        fmt.Printf("[DIR]  %s\n", file.Name)
//	    } else {
//	        fmt.Printf("[FILE] %s (%d bytes)\n", file.Name, file.Size)
//	    }
//	}
//
// Returns an error if the path doesn't exist or isn't accessible.
func (f *FileSystemService) ListFiles(ctx context.Context, path string) ([]*types.FileInfo, error) {
	return withInstrumentation(ctx, f.otel, "FileSystem", "ListFiles", func(ctx context.Context) ([]*types.FileInfo, error) {
		files, httpResp, err := f.toolboxClient.FileSystemAPI.ListFiles(ctx).Path(path).Execute()
		if err != nil {
			return nil, errors.ConvertToolboxError(err, httpResp)
		}

		// Convert toolbox.FileInfo to types.FileInfo
		result := make([]*types.FileInfo, len(files))
		for i, file := range files {
			modTime, _ := time.Parse(time.RFC3339, file.GetModTime())
			result[i] = &types.FileInfo{
				Name:         file.GetName(),
				Size:         int64(file.GetSize()),
				Mode:         file.GetMode(),
				ModifiedTime: modTime,
				IsDirectory:  file.GetIsDir(),
			}
		}

		return result, nil
	})
}

// GetFileInfo retrieves metadata for a file or directory.
//
// The path parameter specifies the file or directory path.
//
// Returns [types.FileInfo] containing the file's name, size, permissions,
// modification time, and whether it's a directory.
//
// Example:
//
//	info, err := sandbox.FileSystem.GetFileInfo(ctx, "/home/user/file.txt")
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("Size: %d bytes, Modified: %s\n", info.Size, info.ModifiedTime)
//
// Returns an error if the path doesn't exist.
func (f *FileSystemService) GetFileInfo(ctx context.Context, path string) (*types.FileInfo, error) {
	return withInstrumentation(ctx, f.otel, "FileSystem", "GetFileInfo", func(ctx context.Context) (*types.FileInfo, error) {
		fileInfo, httpResp, err := f.toolboxClient.FileSystemAPI.GetFileInfo(ctx).Path(path).Execute()
		if err != nil {
			return nil, errors.ConvertToolboxError(err, httpResp)
		}

		modTime, _ := time.Parse(time.RFC3339, fileInfo.GetModTime())
		return &types.FileInfo{
			Name:         fileInfo.GetName(),
			Size:         int64(fileInfo.GetSize()),
			Mode:         fileInfo.GetMode(),
			ModifiedTime: modTime,
			IsDirectory:  fileInfo.GetIsDir(),
		}, nil
	})
}

// DeleteFile deletes a file or directory.
//
// Parameters:
//   - path: The file or directory path to delete
//   - recursive: If true, delete directories and their contents recursively
//
// Example:
//
//	// Delete a file
//	err := sandbox.FileSystem.DeleteFile(ctx, "/home/user/file.txt", false)
//
//	// Delete a directory recursively
//	err := sandbox.FileSystem.DeleteFile(ctx, "/home/user/mydir", true)
//
// Returns an error if the deletion fails (e.g., path doesn't exist, permission denied,
// or attempting to delete a non-empty directory without recursive=true).
func (f *FileSystemService) DeleteFile(ctx context.Context, path string, recursive bool) error {
	return withInstrumentationVoid(ctx, f.otel, "FileSystem", "DeleteFile", func(ctx context.Context) error {
		httpResp, err := f.toolboxClient.FileSystemAPI.DeleteFile(ctx).Path(path).Recursive(recursive).Execute()
		if err != nil {
			return errors.ConvertToolboxError(err, httpResp)
		}

		return nil
	})
}

// DownloadFile downloads a file from the sandbox.
//
// Parameters:
//   - remotePath: The path to the file in the sandbox
//   - localPath: Optional local path to save the file. If nil, only returns the data.
//
// Returns the file contents as a byte slice. If localPath is provided, also writes
// the contents to that local file.
//
// Example:
//
//	// Download and get contents
//	data, err := sandbox.FileSystem.DownloadFile(ctx, "/home/user/file.txt", nil)
//	fmt.Println(string(data))
//
//	// Download and save to local file
//	localPath := "/tmp/downloaded.txt"
//	data, err := sandbox.FileSystem.DownloadFile(ctx, "/home/user/file.txt", &localPath)
//
// Returns an error if the file doesn't exist or cannot be read.
func (f *FileSystemService) DownloadFile(ctx context.Context, remotePath string, localPath *string) ([]byte, error) {
	return withInstrumentation(ctx, f.otel, "FileSystem", "DownloadFile", func(ctx context.Context) ([]byte, error) {
		file, httpResp, err := f.toolboxClient.FileSystemAPI.DownloadFile(ctx).Path(remotePath).Execute()
		if err != nil {
			return nil, errors.ConvertToolboxError(err, httpResp)
		}
		defer file.Close()

		data, err := io.ReadAll(file)
		if err != nil {
			return nil, errors.NewDaytonaError("Failed to read file", 0, nil)
		}

		if localPath != nil {
			if err := os.WriteFile(*localPath, data, 0644); err != nil {
				return nil, errors.NewDaytonaError(fmt.Sprintf("Failed to write file: %v", err), 0, nil)
			}
		}

		return data, nil
	})
}

// UploadFile uploads a file to the sandbox.
//
// Parameters:
//   - source: Either a local file path (string) or file contents ([]byte)
//   - destination: The destination path in the sandbox
//
// Example:
//
//	// Upload from local file path
//	err := sandbox.FileSystem.UploadFile(ctx, "/local/path/file.txt", "/home/user/file.txt")
//
//	// Upload from byte slice
//	content := []byte("Hello, World!")
//	err := sandbox.FileSystem.UploadFile(ctx, content, "/home/user/hello.txt")
//
// Returns an error if the upload fails.
func (f *FileSystemService) UploadFile(ctx context.Context, source any, destination string) error {
	return withInstrumentationVoid(ctx, f.otel, "FileSystem", "UploadFile", func(ctx context.Context) error {
		var data []byte
		var err error

		switch src := source.(type) {
		case []byte:
			data = src
		case string:
			data, err = os.ReadFile(src)
			if err != nil {
				return errors.NewDaytonaError(fmt.Sprintf("Failed to read file: %v", err), 0, nil)
			}
		default:
			return errors.NewDaytonaError("Invalid source type", 0, nil)
		}

		// Create a temporary file for the toolbox API
		tmpFile, err := os.CreateTemp("", "daytona-upload-*")
		if err != nil {
			return errors.NewDaytonaError(fmt.Sprintf("Failed to create temp file: %v", err), 0, nil)
		}
		defer os.Remove(tmpFile.Name())
		defer tmpFile.Close()

		if _, err := tmpFile.Write(data); err != nil {
			return errors.NewDaytonaError(fmt.Sprintf("Failed to write temp file: %v", err), 0, nil)
		}

		// Seek to the beginning for reading
		if _, err := tmpFile.Seek(0, 0); err != nil {
			return errors.NewDaytonaError(fmt.Sprintf("Failed to seek temp file: %v", err), 0, nil)
		}

		_, httpResp, err := f.toolboxClient.FileSystemAPI.UploadFile(ctx).Path(destination).File(tmpFile).Execute()
		if err != nil {
			return errors.ConvertToolboxError(err, httpResp)
		}

		return nil
	})
}

// MoveFiles moves or renames a file or directory.
//
// Parameters:
//   - source: The current path of the file or directory
//   - destination: The new path for the file or directory
//
// This operation can be used for both moving and renaming:
//   - Same directory, different name = rename
//   - Different directory = move
//
// Example:
//
//	// Rename a file
//	err := sandbox.FileSystem.MoveFiles(ctx, "/home/user/old.txt", "/home/user/new.txt")
//
//	// Move a file to another directory
//	err := sandbox.FileSystem.MoveFiles(ctx, "/home/user/file.txt", "/home/user/backup/file.txt")
//
// Returns an error if the operation fails.
func (f *FileSystemService) MoveFiles(ctx context.Context, source, destination string) error {
	return withInstrumentationVoid(ctx, f.otel, "FileSystem", "MoveFiles", func(ctx context.Context) error {
		httpResp, err := f.toolboxClient.FileSystemAPI.MoveFile(ctx).Source(source).Destination(destination).Execute()
		if err != nil {
			return errors.ConvertToolboxError(err, httpResp)
		}

		return nil
	})
}

// SearchFiles searches for files matching a pattern in a directory.
//
// Parameters:
//   - path: The directory to search in
//   - pattern: The glob pattern to match file names (e.g., "*.txt", "test_*")
//
// Returns a map containing a "files" key with a list of matching file paths.
//
// Example:
//
//	result, err := sandbox.FileSystem.SearchFiles(ctx, "/home/user", "*.go")
//	if err != nil {
//	    return err
//	}
//	files := result.(map[string]any)["files"].([]string)
//	for _, file := range files {
//	    fmt.Println(file)
//	}
//
// Returns an error if the search fails.
func (f *FileSystemService) SearchFiles(ctx context.Context, path, pattern string) (any, error) {
	return withInstrumentation(ctx, f.otel, "FileSystem", "SearchFiles", func(ctx context.Context) (any, error) {
		resp, httpResp, err := f.toolboxClient.FileSystemAPI.SearchFiles(ctx).Path(path).Pattern(pattern).Execute()
		if err != nil {
			return nil, errors.ConvertToolboxError(err, httpResp)
		}

		// Convert to map for backward compatibility
		result := map[string]any{
			"files": resp.GetFiles(),
		}

		return result, nil
	})
}

// FindFiles searches for text content within files.
//
// Parameters:
//   - path: The directory to search in
//   - pattern: The text pattern to search for (supports regex)
//
// Returns a list of matches, each containing the file path, line number, and matching content.
//
// Example:
//
//	result, err := sandbox.FileSystem.FindFiles(ctx, "/home/user/project", "TODO:")
//	if err != nil {
//	    return err
//	}
//	matches := result.([]map[string]any)
//	for _, match := range matches {
//	    fmt.Printf("%s:%d: %s\n", match["file"], match["line"], match["content"])
//	}
//
// Returns an error if the search fails.
func (f *FileSystemService) FindFiles(ctx context.Context, path, pattern string) (any, error) {
	return withInstrumentation(ctx, f.otel, "FileSystem", "FindFiles", func(ctx context.Context) (any, error) {
		matches, httpResp, err := f.toolboxClient.FileSystemAPI.FindInFiles(ctx).Path(path).Pattern(pattern).Execute()
		if err != nil {
			return nil, errors.ConvertToolboxError(err, httpResp)
		}

		// Convert to generic format for backward compatibility
		result := make([]map[string]any, len(matches))
		for i, match := range matches {
			result[i] = map[string]any{
				"file":    match.GetFile(),
				"line":    match.GetLine(),
				"content": match.GetContent(),
			}
		}

		return result, nil
	})
}

// ReplaceInFiles replaces text in multiple files.
//
// Parameters:
//   - files: List of file paths to process
//   - pattern: The text pattern to search for (supports regex)
//   - newValue: The replacement text
//
// Returns a list of results for each file, indicating success or failure.
//
// Example:
//
//	files := []string{"/home/user/file1.txt", "/home/user/file2.txt"}
//	result, err := sandbox.FileSystem.ReplaceInFiles(ctx, files, "oldValue", "newValue")
//	if err != nil {
//	    return err
//	}
//	results := result.([]map[string]any)
//	for _, r := range results {
//	    if r["success"].(bool) {
//	        fmt.Printf("Updated: %s\n", r["file"])
//	    } else {
//	        fmt.Printf("Failed: %s - %s\n", r["file"], r["error"])
//	    }
//	}
//
// Returns an error if the operation fails entirely.
func (f *FileSystemService) ReplaceInFiles(ctx context.Context, files []string, pattern, newValue string) (any, error) {
	return withInstrumentation(ctx, f.otel, "FileSystem", "ReplaceInFiles", func(ctx context.Context) (any, error) {
		req := toolbox.NewReplaceRequest(files, newValue, pattern)
		results, httpResp, err := f.toolboxClient.FileSystemAPI.ReplaceInFiles(ctx).Request(*req).Execute()
		if err != nil {
			return nil, errors.ConvertToolboxError(err, httpResp)
		}

		// Convert to generic format for backward compatibility
		resultList := make([]map[string]any, len(results))
		for i, result := range results {
			entry := map[string]any{
				"file":    result.GetFile(),
				"success": result.GetSuccess(),
			}
			if result.Error != nil {
				entry["error"] = result.GetError()
			}
			resultList[i] = entry
		}

		return resultList, nil
	})
}

// SetFilePermissions sets file permissions, owner, and group.
//
// The path parameter specifies the file or directory to modify.
//
// Optional parameters can be configured using functional options:
//   - [options.WithPermissionMode]: Set Unix file permissions (e.g., "0644")
//   - [options.WithOwner]: Set file owner username
//   - [options.WithGroup]: Set file group name
//
// Example:
//
//	// Set permissions only
//	err := sandbox.FileSystem.SetFilePermissions(ctx, "/home/user/script.sh",
//	    options.WithPermissionMode("0755"),
//	)
//
//	// Set owner and group
//	err := sandbox.FileSystem.SetFilePermissions(ctx, "/home/user/file.txt",
//	    options.WithOwner("root"),
//	    options.WithGroup("users"),
//	)
//
//	// Set all at once
//	err := sandbox.FileSystem.SetFilePermissions(ctx, "/home/user/file.txt",
//	    options.WithPermissionMode("0640"),
//	    options.WithOwner("user"),
//	    options.WithGroup("staff"),
//	)
//
// Returns an error if the operation fails.
func (f *FileSystemService) SetFilePermissions(ctx context.Context, path string, opts ...func(*options.SetFilePermissions)) error {
	return withInstrumentationVoid(ctx, f.otel, "FileSystem", "SetFilePermissions", func(ctx context.Context) error {
		permOpts := options.Apply(opts...)

		req := f.toolboxClient.FileSystemAPI.SetFilePermissions(ctx).Path(path)
		if permOpts.Mode != nil {
			req = req.Mode(*permOpts.Mode)
		}
		if permOpts.Owner != nil {
			req = req.Owner(*permOpts.Owner)
		}
		if permOpts.Group != nil {
			req = req.Group(*permOpts.Group)
		}

		httpResp, err := req.Execute()
		if err != nil {
			return errors.ConvertToolboxError(err, httpResp)
		}

		return nil
	})
}
