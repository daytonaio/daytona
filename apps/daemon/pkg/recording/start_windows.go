//go:build windows

// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package recording

import (
	"fmt"
	"os/exec"
	"syscall"
	"time"

	"github.com/daytonaio/daemon/pkg/winsession"
	"golang.org/x/sys/windows"
)

// consoleSessionTimeout is short: the daemon runs as a service in session 0,
// and gdigrab can only capture the interactive desktop, so by the time a
// recording is requested AutoLogon must already have produced the console
// session. If it hasn't, fail fast instead of stalling the API call.
const consoleSessionTimeout = 5 * time.Second

// newCaptureCmd builds the ffmpeg command for Windows screen capture.
// -f gdigrab: GDI screen capture
// -framerate 30: 30 FPS
// -i desktop: capture the entire virtual screen of the calling session
// -c:v libx264: H.264 codec
// -preset ultrafast: fast encoding for real-time capture
// -pix_fmt yuv420p: standard pixel format for compatibility
//
// The daemon runs as SYSTEM in session 0, whose desktop has nothing to
// capture; spawn ffmpeg with the interactive console user's token so gdigrab
// sees the real desktop (same mechanism as the computer-use plugin spawn).
//
// CreateProcessAsUser references the token into the child during Start(); our
// duplicated handle stays ours and must outlive cmd.Start(), so the returned
// cleanup (which closes it) is deferred to StartRecording exit on every path.
func newCaptureCmd(ffmpegPath, filePath string) (*exec.Cmd, func(), error) {
	cmd := exec.Command(ffmpegPath,
		"-f", "gdigrab",
		"-framerate", "30",
		"-i", "desktop",
		"-c:v", "libx264",
		"-preset", "ultrafast",
		"-pix_fmt", "yuv420p",
		"-y", // Overwrite output file if exists
		filePath,
	)

	token, err := winsession.ActiveConsoleUserToken(consoleSessionTimeout)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to resolve console session for recording: %w", err)
	}

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Token:         syscall.Token(token),
		HideWindow:    true,
		CreationFlags: windows.CREATE_NO_WINDOW,
	}

	return cmd, func() { token.Close() }, nil
}
