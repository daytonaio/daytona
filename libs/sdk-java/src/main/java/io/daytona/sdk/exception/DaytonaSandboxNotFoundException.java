// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.exception;
/**
 * The targeted sandbox does not exist.
 *
 * <p>Subclass of {@link DaytonaNotFoundException}.
 */
public class DaytonaSandboxNotFoundException extends DaytonaNotFoundException {
    public DaytonaSandboxNotFoundException(String message) {
        super(message);
    }

    public DaytonaSandboxNotFoundException(String message, Throwable cause) {
        super(message, cause);
    }

    public DaytonaSandboxNotFoundException(String message, String code, String source) {
        super(message, code, source);
    }

    public DaytonaSandboxNotFoundException(String message, Throwable cause, String code, String source) {
        super(message, cause, code, source);
    }
}
