// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.exception;
/**
 * The accessibility (AT-SPI) bus is not reachable.
 *
 * <p>Subclass of {@link DaytonaServiceUnavailableException}.
 */
public class DaytonaA11yUnavailableException extends DaytonaServiceUnavailableException {
    public DaytonaA11yUnavailableException(String message) {
        super(message);
    }

    public DaytonaA11yUnavailableException(String message, Throwable cause) {
        super(message, cause);
    }

    public DaytonaA11yUnavailableException(String message, String code, String source) {
        super(message, code, source);
    }

    public DaytonaA11yUnavailableException(String message, Throwable cause, String code, String source) {
        super(message, cause, code, source);
    }
}
