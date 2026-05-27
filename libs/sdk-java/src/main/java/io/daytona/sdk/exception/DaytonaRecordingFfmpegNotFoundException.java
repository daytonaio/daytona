// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.exception;
/**
 * ffmpeg binary is not installed; required for recording.
 *
 * <p>Subclass of {@link DaytonaServiceUnavailableException}.
 */
public class DaytonaRecordingFfmpegNotFoundException extends DaytonaServiceUnavailableException {
    public DaytonaRecordingFfmpegNotFoundException(String message) {
        super(message);
    }

    public DaytonaRecordingFfmpegNotFoundException(String message, Throwable cause) {
        super(message, cause);
    }

    public DaytonaRecordingFfmpegNotFoundException(String message, String code, String source) {
        super(message, code, source);
    }

    public DaytonaRecordingFfmpegNotFoundException(String message, Throwable cause, String code, String source) {
        super(message, cause, code, source);
    }
}
