// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.exception;
/**
 * The requested git branch does not exist.
 *
 * <p>Subclass of {@link DaytonaNotFoundException}.
 */
public class DaytonaGitBranchNotFoundException extends DaytonaNotFoundException {
    public DaytonaGitBranchNotFoundException(String message) {
        super(message);
    }

    public DaytonaGitBranchNotFoundException(String message, Throwable cause) {
        super(message, cause);
    }

    public DaytonaGitBranchNotFoundException(String message, String code, String source) {
        super(message, code, source);
    }

    public DaytonaGitBranchNotFoundException(String message, Throwable cause, String code, String source) {
        super(message, cause, code, source);
    }
}
