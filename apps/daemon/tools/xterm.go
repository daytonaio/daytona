// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

// download_xterm.go
package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

const (
	XTERM_VERSION     = "5.3.0"
	XTERM_FIT_VERSION = "0.8.0"
)

func main() {
	// Get project root path
	_, filename, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(filename), "..")
	// Create static directory structure
	staticDir := filepath.Join(projectRoot, "pkg", "terminal", "static")
	err := os.MkdirAll(staticDir, 0755)
	if err != nil {
		fmt.Printf("Error creating directory %s: %v\n", staticDir, err)
		os.Exit(1)
	}

	// Files to download from cdnjs
	files := map[string]string{
		filepath.Join(staticDir, "xterm.js"):           fmt.Sprintf("https://cdn.jsdelivr.net/npm/xterm@%s/lib/xterm.js", XTERM_VERSION),
		filepath.Join(staticDir, "xterm.css"):          fmt.Sprintf("https://cdn.jsdelivr.net/npm/xterm@%s/css/xterm.css", XTERM_VERSION),
		filepath.Join(staticDir, "xterm-addon-fit.js"): fmt.Sprintf("https://cdn.jsdelivr.net/npm/xterm-addon-fit@%s/lib/xterm-addon-fit.js", XTERM_FIT_VERSION),
	}

	// Download each file
	for filePath, url := range files {
		fmt.Printf("Downloading %s...\n", filePath)

		resp, err := http.Get(url)
		if err != nil {
			fmt.Printf("Error downloading %s: %v\n", url, err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		file, err := os.Create(filePath)
		if err != nil {
			fmt.Printf("Error creating file %s: %v\n", filePath, err)
			os.Exit(1)
		}

		_, err = io.Copy(file, resp.Body)
		file.Close()
		if err != nil {
			fmt.Printf("Error writing to file %s: %v\n", filePath, err)
			os.Exit(1)
		}
	}

	fmt.Println("xterm.js files downloaded successfully")
}
