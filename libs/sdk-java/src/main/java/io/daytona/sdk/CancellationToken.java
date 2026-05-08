// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk;

import java.util.concurrent.atomic.AtomicReference;

/**
 * Cancellation token for streaming download/upload operations.
 *
 * <p>Calling {@link #cancel()} aborts an in-flight HTTP request and causes the
 * SDK to throw a {@code DaytonaException} with a "cancelled" message.
 *
 * <p>A token is single-shot and can only be cancelled once. Tokens are not reusable
 * across multiple requests.
 */
public final class CancellationToken {
    private final AtomicReference<Runnable> handler = new AtomicReference<Runnable>();
    private volatile boolean cancelled;

    public void cancel() {
        cancelled = true;
        Runnable h = handler.getAndSet(null);
        if (h != null) {
            h.run();
        }
    }

    public boolean isCancelled() {
        return cancelled;
    }

    /**
     * Internal: registers a handler that runs when cancel() fires.
     * Returns a Runnable that callers must invoke (typically in a finally block)
     * to deregister the handler after the operation completes successfully.
     *
     * <p>Uses double-checked compareAndSet to avoid losing a cancel signal that
     * arrives between the cancelled-check and the handler-set.
     */
    Runnable onCancel(Runnable h) {
        if (cancelled) {
            h.run();
            return () -> {};
        }
        handler.set(h);
        if (cancelled) {
            Runnable existing = handler.getAndSet(null);
            if (existing != null) {
                existing.run();
            }
            return () -> {};
        }
        return () -> handler.compareAndSet(h, null);
    }
}
