// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.exception;
/**
 * A git branch with this name already exists.
 *
 * <p>Subclass of {@link DaytonaConflictException}.
 */
public class DaytonaGitBranchExistsException extends DaytonaConflictException {
    public DaytonaGitBranchExistsException(String message) {
        super(message);
    }

    public DaytonaGitBranchExistsException(String message, Throwable cause) {
        super(message, cause);
    }

    public DaytonaGitBranchExistsException(String message, String code, String source) {
        super(message, code, source);
    }

    public DaytonaGitBranchExistsException(String message, Throwable cause, String code, String source) {
        super(message, cause, code, source);
    }
}
