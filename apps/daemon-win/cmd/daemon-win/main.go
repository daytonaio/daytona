// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package main

import (
	"fmt"
	"os"
	"runtime"
	"time"
)

func main() {
	fmt.Println("===========================================")
	fmt.Println("  Daytona Windows Daemon - Hello World!")
	fmt.Println("===========================================")
	fmt.Printf("OS:        %s\n", runtime.GOOS)
	fmt.Printf("Arch:      %s\n", runtime.GOARCH)
	fmt.Printf("Go:        %s\n", runtime.Version())
	fmt.Printf("Hostname:  %s\n", getHostname())
	fmt.Printf("Time:      %s\n", time.Now().Format(time.RFC3339))
	fmt.Println("===========================================")
	fmt.Println("Daemon started successfully!")
}

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}
