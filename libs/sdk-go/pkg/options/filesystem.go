// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package options

// CreateFolder holds optional parameters for [daytona.FileSystemService.CreateFolder].
type CreateFolder struct {
	Mode *string // Unix file permissions (e.g., "0755")
}

// WithMode sets the Unix file permissions for the created folder.
//
// The mode should be specified as an octal string (e.g., "0755", "0700").
// If not specified, defaults to "0755".
//
// Example:
//
//	err := sandbox.FileSystem.CreateFolder(ctx, "/home/user/mydir",
//	    options.WithMode("0700"),
//	)
func WithMode(mode string) func(*CreateFolder) {
	return func(opts *CreateFolder) {
		opts.Mode = &mode
	}
}

// SetFilePermissions holds optional parameters for [daytona.FileSystemService.SetFilePermissions].
type SetFilePermissions struct {
	Mode  *string // Unix file permissions (e.g., "0644")
	Owner *string // File owner username
	Group *string // File group name
}

// WithPermissionMode sets the Unix file permissions.
//
// The mode should be specified as an octal string (e.g., "0644", "0755").
//
// Example:
//
//	err := sandbox.FileSystem.SetFilePermissions(ctx, "/home/user/file.txt",
//	    options.WithPermissionMode("0644"),
//	)
func WithPermissionMode(mode string) func(*SetFilePermissions) {
	return func(opts *SetFilePermissions) {
		opts.Mode = &mode
	}
}

// WithOwner sets the file owner.
//
// The owner should be a valid username on the sandbox system.
//
// Example:
//
//	err := sandbox.FileSystem.SetFilePermissions(ctx, "/home/user/file.txt",
//	    options.WithOwner("root"),
//	)
func WithOwner(owner string) func(*SetFilePermissions) {
	return func(opts *SetFilePermissions) {
		opts.Owner = &owner
	}
}

// WithGroup sets the file group.
//
// The group should be a valid group name on the sandbox system.
//
// Example:
//
//	err := sandbox.FileSystem.SetFilePermissions(ctx, "/home/user/file.txt",
//	    options.WithGroup("users"),
//	)
func WithGroup(group string) func(*SetFilePermissions) {
	return func(opts *SetFilePermissions) {
		opts.Group = &group
	}
}
