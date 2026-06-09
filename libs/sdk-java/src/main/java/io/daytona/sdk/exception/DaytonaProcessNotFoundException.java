// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.exception;
/**
 * The requested process is not running.
 *
 * <p>Subclass of {@link DaytonaNotFoundException}.
 */
public class DaytonaProcessNotFoundException extends DaytonaNotFoundException {
    public DaytonaProcessNotFoundException(String message) {
        super(message);
    }

    public DaytonaProcessNotFoundException(String message, Throwable cause) {
        super(message, cause);
    }

    public DaytonaProcessNotFoundException(String message, String code, String source) {
        super(message, code, source);
    }

    public DaytonaProcessNotFoundException(String message, Throwable cause, String code, String source) {
        super(message, cause, code, source);
    }
}
