// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.exception;
/**
 * Worktree has uncommitted changes.
 *
 * <p>Subclass of {@link DaytonaConflictException}.
 */
public class DaytonaGitDirtyWorktreeException extends DaytonaConflictException {
    public DaytonaGitDirtyWorktreeException(String message) {
        super(message);
    }

    public DaytonaGitDirtyWorktreeException(String message, Throwable cause) {
        super(message, cause);
    }

    public DaytonaGitDirtyWorktreeException(String message, String code, String source) {
        super(message, code, source);
    }

    public DaytonaGitDirtyWorktreeException(String message, Throwable cause, String code, String source) {
        super(message, cause, code, source);
    }
}
