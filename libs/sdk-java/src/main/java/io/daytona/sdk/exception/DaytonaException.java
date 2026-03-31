// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.exception;

public class DaytonaException extends RuntimeException {
    private final int statusCode;

    public DaytonaException(String message) {
        super(message);
        this.statusCode = 0;
    }

    public DaytonaException(String message, Throwable cause) {
        super(message, cause);
        this.statusCode = 0;
    }

    public DaytonaException(int statusCode, String message) {
        super(message);
        this.statusCode = statusCode;
    }

    public int getStatusCode() {
        return statusCode;
    }
}