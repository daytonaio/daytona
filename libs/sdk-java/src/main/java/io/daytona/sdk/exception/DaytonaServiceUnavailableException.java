// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.exception;
/**
 * Raised for HTTP 503 — the service is temporarily refusing traffic.
 */
public class DaytonaServiceUnavailableException extends DaytonaServerException {
    public static final int STATUS_CODE = 503;

    public DaytonaServiceUnavailableException(String message) {
        super(STATUS_CODE, message);
    }

    public DaytonaServiceUnavailableException(String message, Throwable cause) {
        super(STATUS_CODE, message, cause);
    }

    public DaytonaServiceUnavailableException(String message, String code, String source) {
        super(STATUS_CODE, message, code, source);
    }

    public DaytonaServiceUnavailableException(String message, Throwable cause, String code, String source) {
        super(STATUS_CODE, message, cause, code, source);
    }
}
