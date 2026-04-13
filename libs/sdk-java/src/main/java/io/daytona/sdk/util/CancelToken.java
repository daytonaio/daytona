// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.util;

import java.util.concurrent.atomic.AtomicBoolean;

/**
 * Thread-safe cancellation token for aborting in-flight SDK operations.
 *
 * <p>When {@link #cancel()} is called, any SDK method that was passed this token
 * will throw {@link io.daytona.sdk.exception.DaytonaException} at the earliest opportunity,
 * and the underlying HTTP connection will be abandoned, causing the server to
 * terminate the running process.
 */
public class CancelToken {
    private final AtomicBoolean cancelled = new AtomicBoolean(false);

    public void cancel() {
        cancelled.set(true);
    }

    public boolean isCancelled() {
        return cancelled.get();
    }
}
