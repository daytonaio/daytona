// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.exception;

public class DaytonaTimeoutException extends DaytonaException {
    public DaytonaTimeoutException(String message, Throwable cause) {
        super(message, cause);
    }

    public DaytonaTimeoutException(String message) {
        super(message);
    }
}