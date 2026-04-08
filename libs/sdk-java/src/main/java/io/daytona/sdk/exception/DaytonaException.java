// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.exception;

import java.util.Collections;
import java.util.Map;

/**
 * Base exception for all Daytona SDK errors.
 *
 * <p>Subclasses map to specific HTTP status codes and allow callers to catch
 * precise failure conditions without string-parsing error messages:
 *
 * <pre>{@code
 * try {
 *     Sandbox sandbox = daytona.sandbox().get("nonexistent-id");
 * } catch (DaytonaNotFoundException e) {
 *     // sandbox does not exist
 * } catch (DaytonaAuthenticationException e) {
 *     // invalid API key
 * } catch (DaytonaException e) {
 *     // other SDK error
 * }
 * }</pre>
 */
public class DaytonaException extends RuntimeException {
    private final int statusCode;
    private final Map<String, String> headers;

    /**
     * Creates a generic Daytona exception.
     *
     * @param message error description
     */
    public DaytonaException(String message) {
        super(message);
        this.statusCode = 0;
        this.headers = Collections.emptyMap();
    }

    /**
     * Creates a generic Daytona exception with a cause.
     *
     * @param message error description
     * @param cause root cause
     */
    public DaytonaException(String message, Throwable cause) {
        super(message, cause);
        this.statusCode = 0;
        this.headers = Collections.emptyMap();
    }

    /**
     * Creates a Daytona exception with explicit HTTP status code.
     *
     * @param statusCode HTTP status code
     * @param message error description
     */
    public DaytonaException(int statusCode, String message) {
        super(message);
        this.statusCode = statusCode;
        this.headers = Collections.emptyMap();
    }

    /**
     * Creates a Daytona exception with HTTP status code and headers.
     *
     * @param statusCode HTTP status code
     * @param message error description
     * @param headers response headers
     */
    public DaytonaException(int statusCode, String message, Map<String, String> headers) {
        super(message);
        this.statusCode = statusCode;
        this.headers = headers != null ? Collections.unmodifiableMap(headers) : Collections.emptyMap();
    }

    /** Returns the HTTP status code, or 0 if not applicable. */
    public int getStatusCode() {
        return statusCode;
    }

    /** Returns the HTTP response headers, or an empty map if not available. */
    public Map<String, String> getHeaders() {
        return headers;
    }
}
