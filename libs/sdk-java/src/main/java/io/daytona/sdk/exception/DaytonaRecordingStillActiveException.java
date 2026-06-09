// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.exception;
/**
 * The recording is still running; stop it first.
 *
 * <p>Subclass of {@link DaytonaConflictException}.
 */
public class DaytonaRecordingStillActiveException extends DaytonaConflictException {
    public DaytonaRecordingStillActiveException(String message) {
        super(message);
    }

    public DaytonaRecordingStillActiveException(String message, Throwable cause) {
        super(message, cause);
    }

    public DaytonaRecordingStillActiveException(String message, String code, String source) {
        super(message, code, source);
    }

    public DaytonaRecordingStillActiveException(String message, Throwable cause, String code, String source) {
        super(message, cause, code, source);
    }
}
