// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.exception;
/**
 * Git push was rejected (non-fast-forward / stale ref).
 *
 * <p>Subclass of {@link DaytonaConflictException}.
 */
public class DaytonaGitPushRejectedException extends DaytonaConflictException {
    public DaytonaGitPushRejectedException(String message) {
        super(message);
    }

    public DaytonaGitPushRejectedException(String message, Throwable cause) {
        super(message, cause);
    }

    public DaytonaGitPushRejectedException(String message, String code, String source) {
        super(message, code, source);
    }

    public DaytonaGitPushRejectedException(String message, Throwable cause, String code, String source) {
        super(message, cause, code, source);
    }
}
