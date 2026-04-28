// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.internal;

import com.fasterxml.jackson.databind.JsonNode;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.UUID;
import java.util.concurrent.RejectedExecutionException;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.ScheduledFuture;
import java.util.concurrent.ScheduledThreadPoolExecutor;
import java.util.concurrent.TimeUnit;
import java.util.function.BiConsumer;

public class EventSubscriptionManager {

    private static final long SUBSCRIPTION_TTL_SECONDS = 300;

    private final EventDispatcher dispatcher;
    private final Object lock = new Object();
    private final Map<String, Subscription> subscriptions = new HashMap<>();
    private final ScheduledExecutorService scheduler;
    private volatile boolean closed;

    public EventSubscriptionManager(EventDispatcher dispatcher) {
        this.dispatcher = dispatcher;
        ScheduledThreadPoolExecutor scheduler = new ScheduledThreadPoolExecutor(1, r -> {
            Thread t = new Thread(r, "EventSubscriptionManager-timer");
            t.setDaemon(true);
            return t;
        });
        scheduler.setRemoveOnCancelPolicy(true); // Eagerly remove canceled tasks to prevent queue buildup from refresh()
        this.scheduler = scheduler;
    }

    public String subscribe(String resourceId, BiConsumer<String, JsonNode> handler, List<String> events) {
        if (closed) {
            return null; // Reject after shutdown to prevent use-after-close
        }

        Runnable unsubscribe = dispatcher.subscribe(resourceId, handler, events);
        if (closed) {
            unsubscribe.run();
            return null;
        }

        String subId = UUID.randomUUID().toString();
        boolean rollback = false;
        boolean schedulingFailed = false;

        synchronized (lock) {
            if (closed) {
                rollback = true;
            } else {
                subscriptions.put(subId, new Subscription(unsubscribe));
                try {
                    scheduleTimerLocked(subId);
                } catch (RejectedExecutionException e) {
                    subscriptions.remove(subId);
                    rollback = true;
                    schedulingFailed = true;
                }
            }
        }

        if (rollback) {
            if (schedulingFailed) {
                unsubscribe.run(); // Rollback dispatcher subscription on scheduling failure
            } else {
                unsubscribe.run();
            }
            return null;
        }

        return subId;
    }

    public boolean refresh(String subId) {
        if (closed) {
            return false; // Reject after shutdown to prevent use-after-close
        }

        synchronized (lock) {
            if (closed || !subscriptions.containsKey(subId)) {
                return false;
            }
            scheduleTimerLocked(subId);
            return true;
        }
    }

    public void unsubscribe(String subId) {
        Subscription subscription;
        synchronized (lock) {
            subscription = subscriptions.remove(subId);
            if (subscription == null) {
                return;
            }
            if (subscription.timer != null) {
                subscription.timer.cancel(false);
                subscription.timer = null;
            }
        }

        subscription.unsubscribe.run();
    }

    public void shutdown() {
        closed = true;
        List<Subscription> toUnsubscribe;
        synchronized (lock) {
            toUnsubscribe = new ArrayList<>(subscriptions.values());
            subscriptions.clear();
        }

        for (Subscription subscription : toUnsubscribe) {
            if (subscription.timer != null) {
                subscription.timer.cancel(false);
                subscription.timer = null;
            }
        }

        for (Subscription subscription : toUnsubscribe) {
            subscription.unsubscribe.run();
        }

        scheduler.shutdown();
        try {
            if (!scheduler.awaitTermination(5, TimeUnit.SECONDS)) {
                scheduler.shutdownNow();
            }
        } catch (InterruptedException e) {
            scheduler.shutdownNow();
            Thread.currentThread().interrupt();
        }
    }

    private void scheduleTimerLocked(String subId) {
        Subscription subscription = subscriptions.get(subId);
        if (subscription == null) {
            return;
        }

        if (subscription.timer != null) {
            subscription.timer.cancel(false);
        }

        final ScheduledFuture<?>[] timerRef = new ScheduledFuture<?>[1];
        ScheduledFuture<?> timer = scheduler.schedule(() -> {
            Subscription expired = null;
            synchronized (lock) {
                Subscription current = subscriptions.get(subId);
                if (current == null || current.timer != timerRef[0]) {
                    return;
                }
                expired = subscriptions.remove(subId);
                if (expired != null) {
                    expired.timer = null;
                }
            }

            if (expired != null) {
                expired.unsubscribe.run();
            }
        }, SUBSCRIPTION_TTL_SECONDS, TimeUnit.SECONDS);

        timerRef[0] = timer;
        subscription.timer = timer;
    }

    private static final class Subscription {
        private final Runnable unsubscribe;
        private ScheduledFuture<?> timer;

        private Subscription(Runnable unsubscribe) {
            this.unsubscribe = unsubscribe;
        }
    }
}
