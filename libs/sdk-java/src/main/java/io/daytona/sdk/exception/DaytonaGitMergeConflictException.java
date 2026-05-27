// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.exception;
/**
 * Git merge has conflicts that need manual resolution.
 *
 * <p>Subclass of {@link DaytonaConflictException}.
 */
public class DaytonaGitMergeConflictException extends DaytonaConflictException {
    public DaytonaGitMergeConflictException(String message) {
        super(message);
    }

    public DaytonaGitMergeConflictException(String message, Throwable cause) {
        super(message, cause);
    }

    public DaytonaGitMergeConflictException(String message, String code, String source) {
        super(message, code, source);
    }

    public DaytonaGitMergeConflictException(String message, Throwable cause, String code, String source) {
        super(message, cause, code, source);
    }
}
