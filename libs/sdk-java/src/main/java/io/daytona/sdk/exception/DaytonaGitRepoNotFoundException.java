// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.exception;
/**
 * The requested git repository does not exist.
 *
 * <p>Subclass of {@link DaytonaNotFoundException}.
 */
public class DaytonaGitRepoNotFoundException extends DaytonaNotFoundException {
    public DaytonaGitRepoNotFoundException(String message) {
        super(message);
    }

    public DaytonaGitRepoNotFoundException(String message, Throwable cause) {
        super(message, cause);
    }

    public DaytonaGitRepoNotFoundException(String message, String code, String source) {
        super(message, code, source);
    }

    public DaytonaGitRepoNotFoundException(String message, Throwable cause, String code, String source) {
        super(message, cause, code, source);
    }
}
