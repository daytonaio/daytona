// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.exception;
/**
 * Raised for HTTP 502 — an upstream dependency rejected or dropped the request.
 */
public class DaytonaBadGatewayException extends DaytonaServerException {
    public static final int STATUS_CODE = 502;

    public DaytonaBadGatewayException(String message) {
        super(STATUS_CODE, message);
    }

    public DaytonaBadGatewayException(String message, Throwable cause) {
        super(STATUS_CODE, message, cause);
    }

    public DaytonaBadGatewayException(String message, String code, String source) {
        super(STATUS_CODE, message, code, source);
    }

    public DaytonaBadGatewayException(String message, Throwable cause, String code, String source) {
        super(STATUS_CODE, message, cause, code, source);
    }
}
