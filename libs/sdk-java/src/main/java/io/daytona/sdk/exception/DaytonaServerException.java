// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.exception;

/**
 * Raised for unexpected server-side failures (HTTP 5xx).
 *
 * <p>These are typically transient and safe to retry with exponential backoff.
 *
 * <pre>{@code
 * try {
 *     daytona.sandbox().create();
 * } catch (DaytonaServerException e) {
 *     System.err.println("Server error (status " + e.getStatusCode() + "), retry later");
 * }
 * }</pre>
 */
public class DaytonaServerException extends DaytonaException {
    public DaytonaServerException(int statusCode, String message) {
        super(statusCode, message);
    }
}
