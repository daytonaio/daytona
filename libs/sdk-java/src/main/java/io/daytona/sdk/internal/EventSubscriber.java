// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.internal;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import io.socket.client.IO;
import io.socket.client.Socket;
import io.socket.client.SocketOptionBuilder;
import io.socket.engineio.client.transports.WebSocket;

import java.net.URI;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import java.util.Map;
import java.util.Set;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.Executors;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.ScheduledFuture;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicInteger;
import java.util.function.BiConsumer;
import java.util.logging.Level;
import java.util.logging.Logger;

/**
 * Manages a Socket.IO connection to the Daytona notification gateway and
 * dispatches sandbox events to per-resource handlers.
 *
 * <p>Events are dynamically registered and dispatched based on the event names
 * passed to {@link #subscribe}. When no listeners remain, the connection is
 * closed after a 30-second grace period and re-established when a new
 * subscription is added.
 */
public class EventSubscriber {

    private static final Logger LOG = Logger.getLogger(EventSubscriber.class.getName());
    private static final long DISCONNECT_DELAY_SECONDS = 30;
    private static final ObjectMapper MAPPER = new ObjectMapper();
    private static final String[] NESTED_ENTITY_KEYS = {"sandbox", "volume", "snapshot", "runner"};

    private final String apiUrl;
    private final String token;
    private final String organizationId;

    private volatile Socket socket;
    private final Object lock = new Object();
    private final Map<String, Map<Integer, BiConsumer<String, JsonNode>>> listeners = new ConcurrentHashMap<>();
    private final AtomicInteger nextSubId = new AtomicInteger(0);
    private final Set<String> registeredEvents = ConcurrentHashMap.newKeySet();
    private final ScheduledExecutorService scheduler;
    private ScheduledFuture<?> disconnectTimer;

    private volatile boolean connected;
    private volatile boolean failed;
    private volatile String failError = "";

    public EventSubscriber(String apiUrl, String token) {
        this(apiUrl, token, null);
    }

    public EventSubscriber(String apiUrl, String token, String organizationId) {
        this.apiUrl = apiUrl;
        this.token = token;
        this.organizationId = organizationId;
        this.scheduler = Executors.newSingleThreadScheduledExecutor(r -> {
            Thread t = new Thread(r, "daytona-event-subscriber");
            t.setDaemon(true);
            return t;
        });
    }

    /**
     * Ensures a connection attempt is in progress or already established.
     * Non-blocking; starts a background thread to connect if needed.
     */
    public void ensureConnected() {
        if (connected) {
            return;
        }
        scheduler.execute(() -> {
            try {
                connect();
            } catch (Exception e) {
                LOG.log(Level.FINE, "Background connect failed", e);
            }
        });
    }

    /**
     * Establishes the Socket.IO connection. Safe to call from any thread.
     * Connection errors are captured in the {@code failed} flag — they never propagate.
     */
    public void connect() {
        try {
            synchronized (lock) {
                if (connected || socket != null) {
                    return;
                }
            }

            URI apiUri = URI.create(apiUrl);
            String scheme = apiUri.getScheme() != null ? apiUri.getScheme() : "https";
            URI baseUri = new URI(scheme, null, apiUri.getHost(), apiUri.getPort(), null, null, null);

            SocketOptionBuilder builder = SocketOptionBuilder.builder()
                    .setPath("/api/socket.io/")
                    .setTransports(new String[]{WebSocket.NAME})
                    .setAuth(Collections.singletonMap("token", token))
                    .setReconnection(true)
                    .setReconnectionDelay(1000)
                    .setReconnectionDelayMax(30000)
                    .setReconnectionAttempts(Integer.MAX_VALUE);

            if (organizationId != null && !organizationId.isEmpty()) {
                builder.setQuery("organizationId=" + organizationId);
            }

            Socket sock = IO.socket(baseUri, builder.build());

            sock.on(Socket.EVENT_CONNECT, args -> {
                synchronized (lock) {
                    connected = true;
                    failed = false;
                    failError = "";
                }
                reregisterAllEvents(sock);
            });

            sock.on(Socket.EVENT_CONNECT_ERROR, args -> {
                synchronized (lock) {
                    connected = false;
                }
            });

            sock.on(Socket.EVENT_DISCONNECT, args -> {
                synchronized (lock) {
                    connected = false;
                }
            });

            reregisterAllEvents(sock);

            synchronized (lock) {
                this.socket = sock;
            }

            sock.connect();
        } catch (Exception e) {
            synchronized (lock) {
                failed = true;
                failError = "WebSocket connection failed: " + e.getMessage();
            }
        }
    }

    /**
     * Subscribes to events for a specific resource.
     *
     * @param resourceId entity identifier to listen for
     * @param handler    called with {@code (eventName, jsonData)} for matching events
     * @param events     Socket.IO event names to register
     * @return unsubscribe callback
     */
    public Runnable subscribe(String resourceId, BiConsumer<String, JsonNode> handler, List<String> events) {
        synchronized (lock) {
            if (disconnectTimer != null) {
                disconnectTimer.cancel(false);
                disconnectTimer = null;
            }

            for (String event : events) {
                if (registeredEvents.add(event) && socket != null) {
                    registerSocketEvent(socket, event);
                }
            }
        }

        int subId = nextSubId.getAndIncrement();
        listeners.computeIfAbsent(resourceId, k -> new ConcurrentHashMap<>()).put(subId, handler);

        return () -> {
            Map<Integer, BiConsumer<String, JsonNode>> subs = listeners.get(resourceId);
            if (subs != null) {
                subs.remove(subId);
                if (subs.isEmpty()) {
                    listeners.remove(resourceId);
                }
            }
            if (listeners.isEmpty()) {
                synchronized (lock) {
                    disconnectTimer = scheduler.schedule(() -> {
                        if (listeners.isEmpty()) {
                            disconnect();
                        }
                    }, DISCONNECT_DELAY_SECONDS, TimeUnit.SECONDS);
                }
            }
        };
    }

    boolean isConnected() {
        return connected;
    }

    boolean isFailed() {
        return failed;
    }

    String getFailError() {
        return failError;
    }

    /**
     * Disconnects and removes all listeners and event registrations.
     */
    public void disconnect() {
        synchronized (lock) {
            if (disconnectTimer != null) {
                disconnectTimer.cancel(false);
                disconnectTimer = null;
            }
            if (socket != null) {
                socket.off();
                socket.disconnect();
                socket.close();
                socket = null;
            }
            connected = false;
        }
        listeners.clear();
        registeredEvents.clear();
    }

    /**
     * Shuts down the subscriber and its background scheduler.
     */
    public void shutdown() {
        disconnect();
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

    private void reregisterAllEvents(Socket sock) {
        for (String event : registeredEvents) {
            registerSocketEvent(sock, event);
        }
    }

    private void registerSocketEvent(Socket sock, String eventName) {
        sock.on(eventName, args -> handleEventArgs(eventName, args));
    }

    private void handleEventArgs(String eventName, Object[] args) {
        if (args == null || args.length == 0) {
            return;
        }
        try {
            JsonNode data = MAPPER.readTree(args[0].toString());
            handleEvent(eventName, data);
        } catch (Exception e) {
            LOG.log(Level.FINE, "Failed to parse event data for " + eventName, e);
        }
    }

    private void handleEvent(String eventName, JsonNode data) {
        if (!registeredEvents.contains(eventName)) {
            return;
        }

        String entityId = extractEntityId(data);
        if (entityId == null || entityId.isEmpty()) {
            return;
        }

        dispatch(entityId, eventName, data);
    }

    /**
     * Extracts entity ID from event data using nested key lookup.
     * Tries "sandbox", "volume", "snapshot", "runner" nested objects first,
     * then falls back to a top-level "id" field.
     */
    static String extractEntityId(JsonNode data) {
        if (data == null || !data.isObject()) {
            return null;
        }

        for (String key : NESTED_ENTITY_KEYS) {
            JsonNode nested = data.get(key);
            if (nested != null && nested.isObject()) {
                JsonNode id = nested.get("id");
                if (id != null && id.isTextual() && !id.asText().isEmpty()) {
                    return id.asText();
                }
            }
        }

        JsonNode id = data.get("id");
        return (id != null && id.isTextual()) ? id.asText() : null;
    }

    private void dispatch(String entityId, String eventName, JsonNode data) {
        Map<Integer, BiConsumer<String, JsonNode>> subs = listeners.get(entityId);
        if (subs == null || subs.isEmpty()) {
            return;
        }

        List<BiConsumer<String, JsonNode>> snapshot = new ArrayList<>(subs.values());
        for (BiConsumer<String, JsonNode> handler : snapshot) {
            try {
                handler.accept(eventName, data);
            } catch (Exception e) {
                LOG.log(Level.FINE, "Event handler threw exception", e);
            }
        }
    }
}
