// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk;

import io.daytona.sdk.exception.DaytonaException;
import okhttp3.OkHttpClient;
import okhttp3.Request;
import okhttp3.WebSocket;
import okhttp3.WebSocketListener;
import okhttp3.mockwebserver.MockResponse;
import okhttp3.mockwebserver.MockWebServer;
import okio.ByteString;
import org.junit.jupiter.api.Test;

import java.util.ArrayList;
import java.util.List;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicInteger;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;

class PtyHandleTest {

    @Test
    void waitForConnectionAndExitParseControlMessages() throws Exception {
        try (MockWebServer server = new MockWebServer()) {
            List<byte[]> output = new ArrayList<byte[]>();
            server.enqueue(new MockResponse().withWebSocketUpgrade(new WebSocketListener() {
                @Override
                public void onOpen(WebSocket webSocket, okhttp3.Response response) {
                    webSocket.send("{\"type\":\"control\",\"status\":\"connected\"}");
                    webSocket.send("hello");
                    webSocket.close(1000, "{\"exitCode\":0,\"exitReason\":\"done\"}");
                }
            }));

            PtyHandle handle = new PtyHandle(new OkHttpClient(), new Request.Builder().url(server.url("/pty")).build(), "pty-1", (id, cols, rows) -> { }, id -> { }, output::add);

            handle.waitForConnection(1);
            PtyResult result = handle.waitForExit(1);

            assertThat(handle.isConnected()).isFalse();
            assertThat(result.getExitCode()).isEqualTo(0);
            assertThat(result.getError()).isEqualTo("done");
            assertThat(new String(output.get(0))).isEqualTo("hello");
        }
    }

    @Test
    void sendInputAndCallbacksWork() throws Exception {
        try (MockWebServer server = new MockWebServer()) {
            CountDownLatch latch = new CountDownLatch(2);
            List<String> messages = new ArrayList<String>();
            server.enqueue(new MockResponse().withWebSocketUpgrade(new WebSocketListener() {
                @Override
                public void onOpen(WebSocket webSocket, okhttp3.Response response) {
                    webSocket.send("{\"type\":\"control\",\"status\":\"connected\"}");
                }

                @Override
                public void onMessage(WebSocket webSocket, String text) {
                    messages.add(text);
                    latch.countDown();
                    if (latch.getCount() == 0) {
                        webSocket.close(1000, "{\"exitCode\":1,\"error\":\"finished\"}");
                    }
                }

                @Override
                public void onMessage(WebSocket webSocket, ByteString bytes) {
                    messages.add(bytes.utf8());
                    latch.countDown();
                    if (latch.getCount() == 0) {
                        webSocket.close(1000, "{\"exitCode\":1,\"error\":\"finished\"}");
                    }
                }
            }));

            AtomicInteger resizeCols = new AtomicInteger();
            AtomicInteger killed = new AtomicInteger();
            PtyHandle handle = new PtyHandle(new OkHttpClient(), new Request.Builder().url(server.url("/pty")).build(), "pty-2",
                    (id, cols, rows) -> resizeCols.set(cols + rows),
                    id -> killed.incrementAndGet(),
                    bytes -> { });

            handle.waitForConnection(1);
            handle.sendInput("ls");
            handle.sendInput("bin".getBytes());
            handle.resize(80, 24);
            handle.kill();
            assertThat(latch.await(1, TimeUnit.SECONDS)).isTrue();
            PtyResult result = handle.waitForExit(1);

            assertThat(messages).containsExactly("ls", "bin");
            assertThat(resizeCols.get()).isEqualTo(104);
            assertThat(killed.get()).isEqualTo(1);
            assertThat(result.getExitCode()).isEqualTo(1);
            assertThat(result.getError()).isEqualTo("finished");
        }
    }

    @Test
    void waitForConnectionFailsOnControlError() throws Exception {
        try (MockWebServer server = new MockWebServer()) {
            server.enqueue(new MockResponse().withWebSocketUpgrade(new WebSocketListener() {
                @Override
                public void onOpen(WebSocket webSocket, okhttp3.Response response) {
                    webSocket.send("{\"type\":\"control\",\"status\":\"error\",\"error\":\"denied\"}");
                }
            }));

            PtyHandle handle = new PtyHandle(new OkHttpClient(), new Request.Builder().url(server.url("/pty")).build(), "pty-3", (id, cols, rows) -> { }, id -> { }, bytes -> { });

            assertThatThrownBy(() -> handle.waitForConnection(1))
                    .isInstanceOf(DaytonaException.class)
                    .hasMessageContaining("PTY connection failed: denied");
        }
    }

    @Test
    void waitForConnectionTimesOutWhenNoControlMessageArrives() throws Exception {
        try (MockWebServer server = new MockWebServer()) {
            server.enqueue(new MockResponse().withWebSocketUpgrade(new WebSocketListener() {
                @Override
                public void onOpen(WebSocket webSocket, okhttp3.Response response) {
                }
            }));

            PtyHandle handle = new PtyHandle(new OkHttpClient(), new Request.Builder().url(server.url("/pty")).build(), "pty-timeout", (id, cols, rows) -> { }, id -> { }, bytes -> { });

            assertThatThrownBy(() -> handle.waitForConnection(0))
                    .isInstanceOf(DaytonaException.class)
                    .hasMessageContaining("Timed out waiting for PTY connection");
        }
    }

    @Test
    void waitForConnectionFailsWhenSocketClosesBeforeConnected() throws Exception {
        try (MockWebServer server = new MockWebServer()) {
            server.enqueue(new MockResponse().withWebSocketUpgrade(new WebSocketListener() {
                @Override
                public void onOpen(WebSocket webSocket, okhttp3.Response response) {
                    webSocket.close(1000, "closed");
                }
            }));

            PtyHandle handle = new PtyHandle(new OkHttpClient(), new Request.Builder().url(server.url("/pty")).build(), "pty-closed", (id, cols, rows) -> { }, id -> { }, bytes -> { });

            assertThatThrownBy(() -> handle.waitForConnection(1))
                    .isInstanceOf(DaytonaException.class)
                    .hasMessageContaining("PTY connection failed: closed");
        }
    }

    @Test
    void waitForExitUsesRawCloseReasonWhenJsonParsingFails() throws Exception {
        try (MockWebServer server = new MockWebServer()) {
            server.enqueue(new MockResponse().withWebSocketUpgrade(new WebSocketListener() {
                @Override
                public void onOpen(WebSocket webSocket, okhttp3.Response response) {
                    webSocket.send("{\"type\":\"control\",\"status\":\"connected\"}");
                    new Thread(() -> {
                        try { Thread.sleep(200); } catch (InterruptedException ignored) { }
                        webSocket.close(1000, "plain-close-reason");
                    }).start();
                }
            }));

            PtyHandle handle = new PtyHandle(new OkHttpClient(), new Request.Builder().url(server.url("/pty")).build(), "pty-raw", (id, cols, rows) -> { }, id -> { }, bytes -> { });
            handle.waitForConnection(2);

            PtyResult result = handle.waitForExit(2);

            assertThat(result.getExitCode()).isEqualTo(-1);
            assertThat(result.getError()).isEqualTo("plain-close-reason");
        }
    }

    @Test
    void waitForExitReturnsTimeoutResult() throws Exception {
        MockWebServer server = new MockWebServer();
        try {
            server.enqueue(new MockResponse().withWebSocketUpgrade(new WebSocketListener() {
                @Override
                public void onOpen(WebSocket webSocket, okhttp3.Response response) {
                    webSocket.send("{\"type\":\"control\",\"status\":\"connected\"}");
                }
            }));

            PtyHandle handle = new PtyHandle(new OkHttpClient(), new Request.Builder().url(server.url("/pty")).build(), "pty-4", (id, cols, rows) -> { }, id -> { }, bytes -> { });
            handle.waitForConnection(1);

            PtyResult result = handle.waitForExit(0);

            assertThat(result.getExitCode()).isEqualTo(-1);
            assertThat(result.getError()).isEqualTo("Timed out waiting for PTY exit");
            handle.disconnect();
            handle.waitForExit(1);
        } finally {
            try {
                server.shutdown();
            } catch (java.io.IOException ignored) {
            }
        }
    }
}
