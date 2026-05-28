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
    /**
     * Wire-format {@code source} values set by the translation layer when a
     * Daytona service stamps them on the wire envelope. A {@code null}
     * {@code source} means the response did not carry a structured envelope
     * (treat as opaque).
     */
    public static final String SOURCE_API = "DAYTONA_API";
    public static final String SOURCE_DAEMON = "DAYTONA_DAEMON";
    public static final String SOURCE_PROXY = "DAYTONA_PROXY";

    private final int statusCode;
    private final Map<String, String> headers;
    private final String code;
    private final String source;

    private DaytonaException(int statusCode, String message, Map<String, String> headers, String code, String source,
            Throwable cause) {
        super(message, cause);
        this.statusCode = statusCode;
        this.headers = headers != null ? Collections.unmodifiableMap(headers) : Collections.emptyMap();
        this.code = code;
        this.source = source;
    }

    /**
     * Creates a generic Daytona exception.
     *
     * @param message error description
     */
    public DaytonaException(String message) {
        this(0, message, Collections.emptyMap(), (String) null, (String) null, (Throwable) null);
    }

    /**
     * Creates a generic Daytona exception with a cause.
     *
     * @param message error description
     * @param cause root cause
     */
    public DaytonaException(String message, Throwable cause) {
        this(0, message, Collections.emptyMap(), null, null, cause);
    }

    /**
     * Creates a Daytona exception with explicit HTTP status code.
     *
     * @param statusCode HTTP status code
     * @param message error description
     */
    public DaytonaException(int statusCode, String message) {
        this(statusCode, message, Collections.emptyMap(), (String) null, (String) null, (Throwable) null);
    }

    /**
     * Creates a Daytona exception with explicit HTTP status code and a cause.
     *
     * @param statusCode HTTP status code
     * @param message error description
     * @param cause root cause
     */
    public DaytonaException(int statusCode, String message, Throwable cause) {
        this(statusCode, message, Collections.emptyMap(), null, null, cause);
    }

    /**
     * Creates a Daytona exception with HTTP status code and headers.
     *
     * @param statusCode HTTP status code
     * @param message error description
     * @param headers response headers
     */
    public DaytonaException(int statusCode, String message, Map<String, String> headers) {
        this(statusCode, message, headers, (String) null, (String) null, (Throwable) null);
    }

    /**
     * Creates a Daytona exception with HTTP status code, headers, error code, and source.
     *
     * @param statusCode HTTP status code
     * @param message error description
     * @param headers response headers
     * @param code machine-readable error code
     * @param source component that originated the error
     */
    public DaytonaException(int statusCode, String message, Map<String, String> headers, String code, String source) {
        this(statusCode, message, headers, code, source, null);
    }

    /**
     * Creates a Daytona exception with HTTP status code, error code, and source.
     *
     * @param statusCode HTTP status code
     * @param message error description
     * @param code machine-readable error code
     * @param source component that originated the error
     */
    public DaytonaException(int statusCode, String message, String code, String source) {
        this(statusCode, message, Collections.emptyMap(), code, source, null);
    }

    /**
     * Creates a Daytona exception with HTTP status code, cause, error code, and source.
     *
     * @param statusCode HTTP status code
     * @param message error description
     * @param cause root cause
     * @param code machine-readable error code
     * @param source component that originated the error
     */
    public DaytonaException(int statusCode, String message, Throwable cause, String code, String source) {
        this(statusCode, message, Collections.emptyMap(), code, source, cause);
    }

    /**
     * Creates a Daytona exception with HTTP status code, headers, cause, error code, and source.
     *
     * @param statusCode HTTP status code
     * @param message error description
     * @param headers response headers
     * @param cause root cause
     * @param code machine-readable error code
     * @param source component that originated the error
     */
    public DaytonaException(int statusCode, String message, Map<String, String> headers, Throwable cause,
            String code, String source) {
        this(statusCode, message, headers, code, source, cause);
    }

    /** Returns the HTTP status code, or 0 if not applicable. */
    public int getStatusCode() {
        return statusCode;
    }

    /** Returns the HTTP response headers, or an empty map if not available. */
    public Map<String, String> getHeaders() {
        return headers;
    }

    /** Returns the machine-readable error code, or null if not available. */
    public String getCode() {
        return code;
    }

    /**
     * Returns the originating service from the wire envelope. {@code null}
     * for SDK-side errors and for responses that don't carry the envelope.
     * Otherwise one of {@link #SOURCE_API}, {@link #SOURCE_DAEMON} or
     * {@link #SOURCE_PROXY}.
     */
    public String getSource() {
        return source;
    }
}
