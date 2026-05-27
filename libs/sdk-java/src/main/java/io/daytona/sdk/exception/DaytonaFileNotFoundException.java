// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.exception;
/**
 * Filesystem entry was not found.
 *
 * <p>Subclass of {@link DaytonaNotFoundException}.
 */
public class DaytonaFileNotFoundException extends DaytonaNotFoundException {
    public DaytonaFileNotFoundException(String message) {
        super(message);
    }

    public DaytonaFileNotFoundException(String message, Throwable cause) {
        super(message, cause);
    }

    public DaytonaFileNotFoundException(String message, String code, String source) {
        super(message, code, source);
    }

    public DaytonaFileNotFoundException(String message, Throwable cause, String code, String source) {
        super(message, cause, code, source);
    }
}
