//go:build linux

// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package recording

import (
	"fmt"
	"os"
	"os/exec"
)

// newCaptureCmd builds the ffmpeg command for Linux screen capture.
// -f x11grab: X11 screen capture
// -framerate 30: 30 FPS
// -i <display>: capture from $DISPLAY (screen 0), defaulting to :0
// -c:v libx264: H.264 codec
// -preset ultrafast: fast encoding for real-time capture
// -pix_fmt yuv420p: standard pixel format for compatibility
//
// The returned cleanup is a no-op on Linux; it exists for the Windows
// console-user token lifetime (see start_windows.go).
func newCaptureCmd(ffmpegPath, filePath string) (*exec.Cmd, func(), error) {
	// DISPLAY is required for X11 capture; default to :0 if not set
	display := os.Getenv("DISPLAY")
	if display == "" {
		display = ":0"
	}

	cmd := exec.Command(ffmpegPath,
		"-f", "x11grab",
		"-framerate", "30",
		"-i", display,
		"-c:v", "libx264",
		"-preset", "ultrafast",
		"-pix_fmt", "yuv420p",
		"-y", // Overwrite output file if exists
		filePath,
	)

	// Set environment to ensure DISPLAY is available
	cmd.Env = append(os.Environ(), fmt.Sprintf("DISPLAY=%s", display))

	return cmd, func() {}, nil
}
