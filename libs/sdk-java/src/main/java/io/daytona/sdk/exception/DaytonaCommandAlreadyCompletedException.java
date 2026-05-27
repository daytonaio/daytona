// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.exception;
/**
 * The shell command already finished.
 *
 * <p>Subclass of {@link DaytonaGoneException}.
 */
public class DaytonaCommandAlreadyCompletedException extends DaytonaGoneException {
    public DaytonaCommandAlreadyCompletedException(String message) {
        super(message);
    }

    public DaytonaCommandAlreadyCompletedException(String message, Throwable cause) {
        super(message, cause);
    }

    public DaytonaCommandAlreadyCompletedException(String message, String code, String source) {
        super(message, code, source);
    }

    public DaytonaCommandAlreadyCompletedException(String message, Throwable cause, String code, String source) {
        super(message, cause, code, source);
    }
}
