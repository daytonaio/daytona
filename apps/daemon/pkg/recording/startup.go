// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package recording

import (
	"errors"
	"fmt"
	"time"
)

// startupConfirmation is how long StartRecording waits to confirm ffmpeg
// survived startup before reporting success. Capture-source failures (no
// X display, no gdigrab desktop, bad device) kill ffmpeg within
// milliseconds, so a ~1s window catches them without making the API call
// noticeably slower.
const startupConfirmation = 1 * time.Second

// confirmStartup blocks until either grace elapses (ffmpeg survived startup;
// returns nil) or the wait goroutine signals done (ffmpeg exited during
// startup; returns an error). On failure the entry is removed from
// activeRecordings: the recording never really started, so callers get an
// error instead of a Recording, and list/stop/get report not-found rather
// than exposing a phantom "failed" entry for a session that was never
// acknowledged.
//
// Consuming done here cannot deadlock a later StopRecording: once the entry
// is popped, StopRecording returns ErrRecordingNotFound before it would ever
// wait on the channel. On the success path nothing is received, so the done
// value remains for StopRecording.
func (s *RecordingService) confirmStartup(id string, active *activeRecording, grace time.Duration) error {
	timer := time.NewTimer(grace)
	defer timer.Stop()

	select {
	case waitErr := <-active.done:
		// The wait goroutine has already run: it either marked the entry
		// failed (unexpected exit) or popped it (clean exit). Pop is
		// idempotent, so clear the map unconditionally and close our write
		// end of the stdin pipe (markFailed closes it too; double-close of
		// an *os.File is harmless and ignored).
		s.activeRecordings.Pop(id)
		if active.stdinPipe != nil {
			_ = active.stdinPipe.Close()
		}
		if waitErr != nil {
			return fmt.Errorf("ffmpeg exited during startup: %w", waitErr)
		}
		// childreap.Reap reports non-zero exits as (exitCode, nil), and a
		// clean instant exit is equally useless for a recording, so the nil
		// error still means startup failure here.
		if active.cmd != nil && active.cmd.ProcessState != nil {
			return fmt.Errorf("ffmpeg exited during startup: %s", active.cmd.ProcessState)
		}
		return errors.New("ffmpeg exited during startup")
	case <-timer.C:
		return nil
	}
}
