// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk;

import io.daytona.sdk.exception.DaytonaException;
import io.daytona.sdk.exception.DaytonaNotFoundException;
import io.daytona.sdk.exception.DaytonaRateLimitException;

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
        String message = responseBody == null || responseBody.isEmpty()
                ? "Request failed with status " + statusCode
                : responseBody;
        if (statusCode == 404) {
            return new DaytonaNotFoundException(message);
        }
        if (statusCode == 429) {
            return new DaytonaRateLimitException(message);
        }
        return new DaytonaException(statusCode, message);
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