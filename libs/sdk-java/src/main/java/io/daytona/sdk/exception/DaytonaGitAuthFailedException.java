// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.exception;
/**
 * Git authentication credentials were rejected by the remote.
 *
 * <p>Subclass of {@link DaytonaAuthenticationException}.
 */
public class DaytonaGitAuthFailedException extends DaytonaAuthenticationException {
    public DaytonaGitAuthFailedException(String message) {
        super(message);
    }

    public DaytonaGitAuthFailedException(String message, Throwable cause) {
        super(message, cause);
    }

    public DaytonaGitAuthFailedException(String message, String code, String source) {
        super(message, code, source);
    }

    public DaytonaGitAuthFailedException(String message, Throwable cause, String code, String source) {
        super(message, cause, code, source);
    }
}
