// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import io.daytona.sdk.exception.DaytonaException;
import okhttp3.OkHttpClient;
import okhttp3.Request;
import okhttp3.Response;
import okhttp3.WebSocket;
import okhttp3.WebSocketListener;
import okio.ByteString;

import java.nio.charset.StandardCharsets;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.TimeUnit;
import java.util.function.Consumer;

/**
 * Handle for interacting with an active PTY session.
 *
 * <p>Supports bidirectional I/O, resize, kill, and waiting for connection/exit events.
 */
public class PtyHandle {
    private static final ObjectMapper OBJECT_MAPPER = new ObjectMapper();

    private final WebSocket ws;
    private final String sessionId;
    private volatile Integer exitCode;
    private volatile String error;
    private volatile boolean connected = false;
    private volatile boolean connectionEstablished = false;
    private final CountDownLatch connectionLatch = new CountDownLatch(1);
    private final CountDownLatch exitLatch = new CountDownLatch(1);

    private final PtyResizeCallback resizeCallback;
    private final PtyKillCallback killCallback;
    private final Consumer<byte[]> onData;

    PtyHandle(
            OkHttpClient client,
            Request request,
            String sessionId,
            PtyResizeCallback resizeCallback,
            PtyKillCallback killCallback,
            Consumer<byte[]> onData
    ) {
        this.sessionId = sessionId;
        this.resizeCallback = resizeCallback;
        this.killCallback = killCallback;
        this.onData = onData;
        this.ws = client.newWebSocket(request, new PtyWebSocketListener());
    }

    /**
     * Waits for PTY websocket connection to be fully established.
     *
     * @param timeoutSeconds maximum seconds to wait
     * @throws DaytonaException if connection fails or times out
     */
    public void waitForConnection(long timeoutSeconds) {
        if (connectionEstablished) {
            return;
        }
        try {
            boolean ready = connectionLatch.await(timeoutSeconds, TimeUnit.SECONDS);
            if (!ready) {
                throw new DaytonaException("Timed out waiting for PTY connection");
            }
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
            throw new DaytonaException("Interrupted while waiting for PTY connection", e);
        }

        if (error != null && !error.isEmpty()) {
            throw new DaytonaException("PTY connection failed: " + error);
        }
        if (!connectionEstablished) {
            throw new DaytonaException("PTY connection was not established");
        }
    }

    /**
     * Sends text input to PTY.
     *
     * @param data UTF-8 text to send
     * @throws DaytonaException if sending fails
     */
    public void sendInput(String data) {
        if (!ws.send(data)) {
            throw new DaytonaException("Failed to send PTY input");
        }
    }

    /**
     * Sends binary input to PTY.
     *
     * @param data binary payload
     * @throws DaytonaException if sending fails
     */
    public void sendInput(byte[] data) {
        if (!ws.send(ByteString.of(data))) {
            throw new DaytonaException("Failed to send PTY binary input");
        }
    }

    /**
     * Waits until the PTY session exits.
     *
     * @return final PTY result
     * @throws DaytonaException if interrupted while waiting
     */
    public PtyResult waitForExit() {
        try {
            exitLatch.await();
            return new PtyResult(exitCode == null ? -1 : exitCode, error);
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
            throw new DaytonaException("Interrupted while waiting for PTY exit", e);
        }
    }

    /**
     * Waits for PTY exit with timeout.
     *
     * @param timeoutSeconds maximum seconds to wait
     * @return final PTY result, or timeout result when exit does not occur in time
     * @throws DaytonaException if interrupted while waiting
     */
    public PtyResult waitForExit(long timeoutSeconds) {
        try {
            boolean finished = exitLatch.await(timeoutSeconds, TimeUnit.SECONDS);
            if (!finished) {
                return new PtyResult(-1, "Timed out waiting for PTY exit");
            }
            return new PtyResult(exitCode == null ? -1 : exitCode, error);
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
            throw new DaytonaException("Interrupted while waiting for PTY exit", e);
        }
    }

    /**
     * Resizes terminal dimensions.
     *
     * @param cols terminal width in columns
     * @param rows terminal height in rows
     */
    public void resize(int cols, int rows) {
        resizeCallback.resize(sessionId, cols, rows);
    }

    /**
     * Terminates PTY session.
     */
    public void kill() {
        killCallback.kill(sessionId);
    }

    /**
     * Disconnects the PTY websocket.
     */
    public void disconnect() {
        ws.close(1000, "client disconnect");
    }

    /**
     * Returns PTY session identifier.
     *
     * @return session ID
     */
    public String getSessionId() {
        return sessionId;
    }

    /**
     * Returns PTY exit code when available.
     *
     * @return exit code, or {@code null} if not known yet
     */
    public Integer getExitCode() {
        return exitCode;
    }

    /**
     * Returns PTY error or exit reason.
     *
     * @return error message, or {@code null} when none
     */
    public String getError() {
        return error;
    }

    /**
     * Returns websocket connectivity status.
     *
     * @return {@code true} when socket is currently connected
     */
    public boolean isConnected() {
        return connected;
    }

    private void emitData(byte[] data) {
        if (onData != null) {
            onData.accept(data);
        }
    }

    private class PtyWebSocketListener extends WebSocketListener {
        @Override
        public void onOpen(WebSocket ws, Response response) {
            connected = true;
        }

        @Override
        public void onMessage(WebSocket ws, String text) {
            try {
                JsonNode node = OBJECT_MAPPER.readTree(text);
                if (node != null && "control".equals(node.path("type").asText())) {
                    String status = node.path("status").asText();
                    if ("connected".equals(status)) {
                        connectionEstablished = true;
                        connectionLatch.countDown();
                        return;
                    }
                    if ("error".equals(status)) {
                        error = node.path("error").asText("PTY control error");
                        connectionLatch.countDown();
                        return;
                    }
                }
            } catch (Exception ignored) {
            }
            emitData(text.getBytes(StandardCharsets.UTF_8));
        }

        @Override
        public void onMessage(WebSocket ws, ByteString bytes) {
            emitData(bytes.toByteArray());
        }

        @Override
        public void onClosing(WebSocket ws, int code, String reason) {
            connected = false;
            parseCloseReason(reason);
            connectionLatch.countDown();
            exitLatch.countDown();
            ws.close(code, reason);
        }

        @Override
        public void onClosed(WebSocket ws, int code, String reason) {
            connected = false;
            parseCloseReason(reason);
            connectionLatch.countDown();
            exitLatch.countDown();
        }

        @Override
        public void onFailure(WebSocket ws, Throwable t, Response response) {
            error = t == null ? "PTY websocket failure" : t.getMessage();
            connected = false;
            connectionLatch.countDown();
            exitLatch.countDown();
        }
    }

    private void parseCloseReason(String reason) {
        if (reason == null || reason.isEmpty()) {
            return;
        }
        try {
            JsonNode node = OBJECT_MAPPER.readTree(reason);
            if (node.has("exitCode") && !node.get("exitCode").isNull()) {
                exitCode = node.get("exitCode").asInt();
            }
            if (node.has("exitReason") && !node.get("exitReason").isNull()) {
                error = node.get("exitReason").asText();
            }
            if (node.has("error") && !node.get("error").isNull()) {
                error = node.get("error").asText();
            }
        } catch (Exception ignored) {
            error = reason;
        }
    }

    @FunctionalInterface
    interface PtyResizeCallback {
        void resize(String sessionId, int cols, int rows);
    }

    @FunctionalInterface
    interface PtyKillCallback {
        void kill(String sessionId);
    }
}
