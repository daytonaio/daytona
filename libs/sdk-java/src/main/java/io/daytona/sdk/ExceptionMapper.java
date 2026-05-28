// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk;

import io.daytona.sdk.exception.DaytonaA11yUnavailableException;
import io.daytona.sdk.exception.DaytonaAuthenticationException;
import io.daytona.sdk.exception.DaytonaBadGatewayException;
import io.daytona.sdk.exception.DaytonaBadRequestException;
import io.daytona.sdk.exception.DaytonaCommandAlreadyCompletedException;
import io.daytona.sdk.exception.DaytonaConflictException;
import io.daytona.sdk.exception.DaytonaConnectionException;
import io.daytona.sdk.exception.DaytonaConnectionTimeoutException;
import io.daytona.sdk.exception.DaytonaException;
import io.daytona.sdk.exception.DaytonaFileAccessDeniedException;
import io.daytona.sdk.exception.DaytonaFileNotFoundException;
import io.daytona.sdk.exception.DaytonaForbiddenException;
import io.daytona.sdk.exception.DaytonaGitAuthFailedException;
import io.daytona.sdk.exception.DaytonaGitBranchExistsException;
import io.daytona.sdk.exception.DaytonaGitBranchNotFoundException;
import io.daytona.sdk.exception.DaytonaGitDirtyWorktreeException;
import io.daytona.sdk.exception.DaytonaGitMergeConflictException;
import io.daytona.sdk.exception.DaytonaGitPushRejectedException;
import io.daytona.sdk.exception.DaytonaGitRepoNotFoundException;
import io.daytona.sdk.exception.DaytonaGoneException;
import io.daytona.sdk.exception.DaytonaInternalServerException;
import io.daytona.sdk.exception.DaytonaLspServerNotInitializedException;
import io.daytona.sdk.exception.DaytonaNotFoundException;
import io.daytona.sdk.exception.DaytonaProcessExecutionTimeoutException;
import io.daytona.sdk.exception.DaytonaProcessNotFoundException;
import io.daytona.sdk.exception.DaytonaRateLimitException;
import io.daytona.sdk.exception.DaytonaRecordingFfmpegNotFoundException;
import io.daytona.sdk.exception.DaytonaRecordingStillActiveException;
import io.daytona.sdk.exception.DaytonaServerException;
import io.daytona.sdk.exception.DaytonaServiceUnavailableException;
import io.daytona.sdk.exception.DaytonaSessionEndedException;
import io.daytona.sdk.exception.DaytonaTimeoutException;
import io.daytona.sdk.exception.DaytonaUnprocessableEntityException;

import java.net.SocketTimeoutException;
import java.util.Collections;
import java.util.HashMap;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

final class ExceptionMapper {
    /** Wire-format source identifiers (kept in sync with Python/TS SDKs). */
    private static final String SRC_API = "DAYTONA_API";
    private static final String SRC_DAEMON = "DAYTONA_DAEMON";
    private static final String SRC_PROXY = "DAYTONA_PROXY";

    @FunctionalInterface
    private interface ErrorFactory {
        DaytonaException create(String message, Throwable cause, String code, String source);
    }

    /**
     * (source, code) → factory. Lookup order: exact (source, code) match → HTTP
     * status class → base DaytonaException. Codes are intentionally inlined
     * as string literals (not enum imports) to mirror the TypeScript SDK and
     * keep the mapper self-contained.
     */
    private static final Map<String, ErrorFactory> CODE_TO_FACTORY = buildCodeMap();

    private ExceptionMapper() {
    }

    static <T> T callMain(MainSupplier<T> supplier) {
        try {
            return supplier.get();
        } catch (io.daytona.api.client.ApiException e) {
            throw map(e.getCode(), e.getResponseBody(), flattenHeaders(e.getResponseHeaders()), e);
        }
    }

    static void runMain(MainRunnable runnable) {
        try {
            runnable.run();
        } catch (io.daytona.api.client.ApiException e) {
            throw map(e.getCode(), e.getResponseBody(), flattenHeaders(e.getResponseHeaders()), e);
        }
    }

    static <T> T callToolbox(ToolboxSupplier<T> supplier) {
        try {
            return supplier.get();
        } catch (io.daytona.toolbox.client.ApiException e) {
            throw map(e.getCode(), e.getResponseBody(), flattenHeaders(e.getResponseHeaders()), e);
        }
    }

    static void runToolbox(ToolboxRunnable runnable) {
        try {
            runnable.run();
        } catch (io.daytona.toolbox.client.ApiException e) {
            throw map(e.getCode(), e.getResponseBody(), flattenHeaders(e.getResponseHeaders()), e);
        }
    }

    static DaytonaException map(int statusCode, String responseBody, Throwable cause) {
        return map(statusCode, responseBody, Collections.emptyMap(), cause);
    }

    static DaytonaException map(int statusCode, String responseBody, Map<String, String> headers, Throwable cause) {
        // Only treat status==0 as a transport failure when the ApiException
        // wraps an underlying Throwable; client-side ApiExceptions thrown for
        // parameter validation also have status==0 but no wrapped cause.
        if (statusCode == 0 && (responseBody == null || responseBody.isEmpty())
                && cause != null && cause.getCause() != null) {
            return mapTransportFailure(cause);
        }
        ErrorDetails errorDetails = extractErrorDetails(responseBody, statusCode);
        String message = errorDetails.message();
        if (statusCode == 0 && (responseBody == null || responseBody.isEmpty())
                && cause != null && cause.getMessage() != null && !cause.getMessage().isEmpty()) {
            message = cause.getMessage();
        }

        // (source, code) exact match takes precedence over the HTTP status.
        if (errorDetails.source() != null && errorDetails.code() != null) {
            ErrorFactory factory = CODE_TO_FACTORY.get(errorDetails.source() + "|" + errorDetails.code());
            if (factory != null) {
                return factory.create(message, cause, errorDetails.code(), errorDetails.source());
            }
        }

        switch (statusCode) {
            case 400:
                return new DaytonaBadRequestException(message, cause, errorDetails.code(), errorDetails.source());
            case 401:
                return new DaytonaAuthenticationException(message, cause, errorDetails.code(), errorDetails.source());
            case 403:
                return new DaytonaForbiddenException(message, cause, errorDetails.code(), errorDetails.source());
            case 404:
                return new DaytonaNotFoundException(message, cause, errorDetails.code(), errorDetails.source());
            case 408:
                return new DaytonaTimeoutException(message, cause, errorDetails.code(), errorDetails.source());
            case 409:
                return new DaytonaConflictException(message, cause, errorDetails.code(), errorDetails.source());
            case 410:
                return new DaytonaGoneException(message, cause, errorDetails.code(), errorDetails.source());
            case 422:
                return new DaytonaUnprocessableEntityException(message, cause, errorDetails.code(), errorDetails.source());
            case 429:
                return new DaytonaRateLimitException(message, cause, errorDetails.code(), errorDetails.source());
            case 500:
                return new DaytonaInternalServerException(message, cause, errorDetails.code(), errorDetails.source());
            case 502:
                return new DaytonaBadGatewayException(message, cause, errorDetails.code(), errorDetails.source());
            case 503:
                return new DaytonaServiceUnavailableException(message, cause, errorDetails.code(), errorDetails.source());
            case 504:
                return new DaytonaTimeoutException(message, cause, errorDetails.code(), errorDetails.source());
            default:
                if (statusCode >= 500) {
                    return new DaytonaServerException(statusCode, message, cause, errorDetails.code(), errorDetails.source());
                }
                return new DaytonaException(statusCode, message, headers, cause, errorDetails.code(), errorDetails.source());
        }
    }

    private static DaytonaException mapTransportFailure(Throwable cause) {
        Throwable root = rootCause(cause);
        String message = rootMessage(root);
        if (root instanceof SocketTimeoutException) {
            return new DaytonaConnectionTimeoutException("Request timed out: " + message, cause);
        }
        return new DaytonaConnectionException("Connection failed: " + message, cause);
    }

    private static Map<String, ErrorFactory> buildCodeMap() {
        Map<String, ErrorFactory> map = new HashMap<>();

        // Daemon: git
        map.put(SRC_DAEMON + "|GIT_AUTH_FAILED", DaytonaGitAuthFailedException::new);
        map.put(SRC_DAEMON + "|GIT_REPO_NOT_FOUND", DaytonaGitRepoNotFoundException::new);
        map.put(SRC_DAEMON + "|GIT_BRANCH_NOT_FOUND", DaytonaGitBranchNotFoundException::new);
        map.put(SRC_DAEMON + "|GIT_BRANCH_EXISTS", DaytonaGitBranchExistsException::new);
        map.put(SRC_DAEMON + "|GIT_PUSH_REJECTED", DaytonaGitPushRejectedException::new);
        map.put(SRC_DAEMON + "|GIT_DIRTY_WORKTREE", DaytonaGitDirtyWorktreeException::new);
        map.put(SRC_DAEMON + "|GIT_MERGE_CONFLICT", DaytonaGitMergeConflictException::new);

        // Daemon: filesystem
        map.put(SRC_DAEMON + "|FILE_NOT_FOUND", DaytonaFileNotFoundException::new);
        map.put(SRC_DAEMON + "|FILE_ACCESS_DENIED", DaytonaFileAccessDeniedException::new);

        // Daemon: lsp
        map.put(SRC_DAEMON + "|LSP_SERVER_NOT_INITIALIZED", DaytonaLspServerNotInitializedException::new);

        // Daemon: process / session
        map.put(SRC_DAEMON + "|PROCESS_EXECUTION_TIMEOUT", DaytonaProcessExecutionTimeoutException::new);
        map.put(SRC_DAEMON + "|PROCESS_NOT_FOUND", DaytonaProcessNotFoundException::new);
        map.put(SRC_DAEMON + "|SESSION_ENDED", DaytonaSessionEndedException::new);
        map.put(SRC_DAEMON + "|COMMAND_ALREADY_COMPLETED", DaytonaCommandAlreadyCompletedException::new);

        // Daemon: computer-use
        map.put(SRC_DAEMON + "|A11Y_UNAVAILABLE", DaytonaA11yUnavailableException::new);
        map.put(SRC_DAEMON + "|RECORDING_STILL_ACTIVE", DaytonaRecordingStillActiveException::new);
        map.put(SRC_DAEMON + "|RECORDING_FFMPEG_NOT_FOUND", DaytonaRecordingFfmpegNotFoundException::new);

        return Collections.unmodifiableMap(map);
    }

    private static Throwable rootCause(Throwable t) {
        Throwable current = t;
        while (current.getCause() != null && current.getCause() != current) {
            current = current.getCause();
        }
        return current;
    }

    private static String rootMessage(Throwable t) {
        String msg = t.getMessage();
        if (msg != null && !msg.isEmpty()) {
            return msg;
        }
        return t.getClass().getSimpleName();
    }

    /**
     * Extracts a human-readable message from a raw JSON response body.
     * Looks for a "message" or "error" field; falls back to the raw body or a generic message.
     */
    private static ErrorDetails extractErrorDetails(String responseBody, int statusCode) {
        if (responseBody == null || responseBody.isEmpty()) {
            return new ErrorDetails("Request failed with status " + statusCode, null, null);
        }

        String message = extractJsonField(responseBody, "message");
        if (message == null) {
            message = extractJsonField(responseBody, "error");
        }
        if (message == null) {
            message = responseBody;
        }

        return new ErrorDetails(
                message,
                extractJsonField(responseBody, "code"),
                extractJsonField(responseBody, "source"));
    }

    private static String extractJsonField(String responseBody, String field) {
        Matcher matcher = Pattern.compile("\"" + Pattern.quote(field) + "\"\\s*:\\s*\"((?:[^\"\\\\]|\\\\.)*)\"")
                .matcher(responseBody);
        if (matcher.find()) {
            return matcher.group(1);
        }
        return null;
    }

    private static Map<String, String> flattenHeaders(Map<String, List<String>> responseHeaders) {
        if (responseHeaders == null || responseHeaders.isEmpty()) {
            return Collections.emptyMap();
        }

        Map<String, String> flattenedHeaders = new LinkedHashMap<>();
        for (Map.Entry<String, List<String>> entry : responseHeaders.entrySet()) {
            List<String> values = entry.getValue();
            flattenedHeaders.put(entry.getKey(), values == null ? "" : String.join(", ", values));
        }
        return flattenedHeaders;
    }

    private static final class ErrorDetails {
        private final String message;
        private final String code;
        private final String source;

        private ErrorDetails(String message, String code, String source) {
            this.message = message;
            this.code = code;
            this.source = source;
        }

        private String message() {
            return message;
        }

        private String code() {
            return code;
        }

        private String source() {
            return source;
        }
    }

    @FunctionalInterface
    interface MainSupplier<T> {
        T get() throws io.daytona.api.client.ApiException;
    }

    @FunctionalInterface
    interface MainRunnable {
        void run() throws io.daytona.api.client.ApiException;
    }

    @FunctionalInterface
    interface ToolboxSupplier<T> {
        T get() throws io.daytona.toolbox.client.ApiException;
    }

    @FunctionalInterface
    interface ToolboxRunnable {
        void run() throws io.daytona.toolbox.client.ApiException;
    }
}
