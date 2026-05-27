// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.exception;
/**
 * The API key used for the request has expired.
 *
 * <p>Subclass of {@link DaytonaAuthenticationException}.
 */
public class DaytonaApiKeyExpiredException extends DaytonaAuthenticationException {
    public DaytonaApiKeyExpiredException(String message) {
        super(message);
    }

    public DaytonaApiKeyExpiredException(String message, Throwable cause) {
        super(message, cause);
    }

    public DaytonaApiKeyExpiredException(String message, String code, String source) {
        super(message, code, source);
    }

    public DaytonaApiKeyExpiredException(String message, Throwable cause, String code, String source) {
        super(message, cause, code, source);
    }
}
