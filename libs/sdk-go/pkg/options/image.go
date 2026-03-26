// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package options

// PipInstall holds optional parameters for [daytona.Image.PipInstall].
type PipInstall struct {
	FindLinks      []string // URLs to search for packages
	IndexURL       string   // Base URL of the Python Package Index
	ExtraIndexURLs []string // Extra index URLs for package lookup
	Pre            bool     // Allow pre-release and development versions
	ExtraOptions   string   // Additional pip command-line options
}

// WithFindLinks adds find-links URLs for pip install.
//
// Find-links URLs are searched for packages before the package index.
// Useful for installing packages from local directories or custom URLs.
//
// Example:
//
//	image := daytona.Base("python:3.11").PipInstall(
//	    []string{"mypackage"},
//	    options.WithFindLinks("/path/to/wheels", "https://example.com/wheels/"),
//	)
func WithFindLinks(links ...string) func(*PipInstall) {
	return func(opts *PipInstall) {
		opts.FindLinks = append(opts.FindLinks, links...)
	}
}

// WithIndexURL sets the base URL of the Python Package Index.
//
// Replaces the default PyPI (https://pypi.org/simple) with a custom index.
//
// Example:
//
//	image := daytona.Base("python:3.11").PipInstall(
//	    []string{"mypackage"},
//	    options.WithIndexURL("https://my-pypi.example.com/simple/"),
//	)
func WithIndexURL(url string) func(*PipInstall) {
	return func(opts *PipInstall) {
		opts.IndexURL = url
	}
}

// WithExtraIndexURLs adds extra index URLs for pip install.
//
// Extra indexes are checked in addition to the main index URL.
// Useful for installing packages from both PyPI and a private index.
//
// Example:
//
//	image := daytona.Base("python:3.11").PipInstall(
//	    []string{"mypackage"},
//	    options.WithExtraIndexURLs("https://private.example.com/simple/"),
//	)
func WithExtraIndexURLs(urls ...string) func(*PipInstall) {
	return func(opts *PipInstall) {
		opts.ExtraIndexURLs = append(opts.ExtraIndexURLs, urls...)
	}
}

// WithPre enables installation of pre-release and development versions.
//
// Example:
//
//	image := daytona.Base("python:3.11").PipInstall(
//	    []string{"mypackage"},
//	    options.WithPre(),
//	)
func WithPre() func(*PipInstall) {
	return func(opts *PipInstall) {
		opts.Pre = true
	}
}

// WithExtraOptions adds extra command-line options for pip install.
//
// Use this for pip options not covered by other With* functions.
//
// Example:
//
//	image := daytona.Base("python:3.11").PipInstall(
//	    []string{"mypackage"},
//	    options.WithExtraOptions("--no-cache-dir --upgrade"),
//	)
func WithExtraOptions(options string) func(*PipInstall) {
	return func(opts *PipInstall) {
		opts.ExtraOptions = options
	}
}
