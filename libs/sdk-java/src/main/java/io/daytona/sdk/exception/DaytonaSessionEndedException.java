// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.exception;
/**
 * The shell session has ended.
 *
 * <p>Subclass of {@link DaytonaGoneException}.
 */
public class DaytonaSessionEndedException extends DaytonaGoneException {
    public DaytonaSessionEndedException(String message) {
        super(message);
    }

    public DaytonaSessionEndedException(String message, Throwable cause) {
        super(message, cause);
    }

    public DaytonaSessionEndedException(String message, String code, String source) {
        super(message, code, source);
    }

    public DaytonaSessionEndedException(String message, Throwable cause, String code, String source) {
        super(message, cause, code, source);
    }
}
