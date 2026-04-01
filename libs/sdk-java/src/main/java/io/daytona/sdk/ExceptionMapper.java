// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk;

import io.daytona.sdk.exception.DaytonaAuthenticationException;
import io.daytona.sdk.exception.DaytonaBadRequestException;
import io.daytona.sdk.exception.DaytonaConflictException;
import io.daytona.sdk.exception.DaytonaException;
import io.daytona.sdk.exception.DaytonaForbiddenException;
import io.daytona.sdk.exception.DaytonaNotFoundException;
import io.daytona.sdk.exception.DaytonaRateLimitException;
import io.daytona.sdk.exception.DaytonaServerException;
import io.daytona.sdk.exception.DaytonaValidationException;

import java.util.regex.Matcher;
import java.util.regex.Pattern;

final class ExceptionMapper {
    private ExceptionMapper() {
    }

    static <T> T callMain(MainSupplier<T> supplier) {
        try {
            return supplier.get();
        } catch (io.daytona.api.client.ApiException e) {
            throw map(e.getCode(), e.getResponseBody());
        }
    }

    static void runMain(MainRunnable runnable) {
        try {
            runnable.run();
        } catch (io.daytona.api.client.ApiException e) {
            throw map(e.getCode(), e.getResponseBody());
        }
    }

    static <T> T callToolbox(ToolboxSupplier<T> supplier) {
        try {
            return supplier.get();
        } catch (io.daytona.toolbox.client.ApiException e) {
            throw map(e.getCode(), e.getResponseBody());
        }
    }

    static void runToolbox(ToolboxRunnable runnable) {
        try {
            runnable.run();
        } catch (io.daytona.toolbox.client.ApiException e) {
            throw map(e.getCode(), e.getResponseBody());
        }
    }

    private static DaytonaException map(int statusCode, String responseBody) {
        String message = extractMessage(responseBody, statusCode);
        switch (statusCode) {
            case 400:
                return new DaytonaBadRequestException(message);
            case 401:
                return new DaytonaAuthenticationException(message);
            case 403:
                return new DaytonaForbiddenException(message);
            case 404:
                return new DaytonaNotFoundException(message);
            case 409:
                return new DaytonaConflictException(message);
            case 422:
                return new DaytonaValidationException(message);
            case 429:
                return new DaytonaRateLimitException(message);
            default:
                if (statusCode >= 500) {
                    return new DaytonaServerException(statusCode, message);
                }
                return new DaytonaException(statusCode, message);
        }
    }

    /**
     * Extracts a human-readable message from a raw JSON response body.
     * Looks for a "message" or "error" field; falls back to the raw body or a generic message.
     */
    private static String extractMessage(String responseBody, int statusCode) {
        if (responseBody == null || responseBody.isEmpty()) {
            return "Request failed with status " + statusCode;
        }
        // Try to extract "message" field from JSON
        Matcher messageMatcher = Pattern.compile("\"message\"\\s*:\\s*\"((?:[^\"\\\\]|\\\\.)*)\"")
                .matcher(responseBody);
        if (messageMatcher.find()) {
            return messageMatcher.group(1);
        }
        // Try to extract "error" field from JSON
        Matcher errorMatcher = Pattern.compile("\"error\"\\s*:\\s*\"((?:[^\"\\\\]|\\\\.)*)\"")
                .matcher(responseBody);
        if (errorMatcher.find()) {
            return errorMatcher.group(1);
        }
        return responseBody;
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