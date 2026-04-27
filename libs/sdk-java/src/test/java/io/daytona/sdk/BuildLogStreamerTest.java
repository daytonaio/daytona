// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk;

import io.daytona.sdk.exception.DaytonaException;
import okhttp3.mockwebserver.MockResponse;
import okhttp3.mockwebserver.MockWebServer;
import org.junit.jupiter.api.Test;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import java.util.concurrent.atomic.AtomicInteger;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;

class BuildLogStreamerTest {

    @Test
    void streamLogsForwardsLinesAndAddsFollowParameter() throws Exception {
        try (MockWebServer server = new MockWebServer()) {
            server.enqueue(new MockResponse().setBody("line-1\nline-2\nline-3\n"));
            List<String> lines = new ArrayList<String>();
            AtomicInteger counter = new AtomicInteger();

            new BuildLogStreamer(new okhttp3.OkHttpClient(), "secret")
                    .streamLogs(server.url("/logs?name=snapshot").toString(), lines::add, () -> counter.incrementAndGet() >= 2);

            okhttp3.mockwebserver.RecordedRequest request = server.takeRequest();
            assertThat(lines).containsExactly("line-1", "line-2");
            assertThat(request.getPath()).isEqualTo("/logs?name=snapshot&follow=true");
            assertThat(request.getHeader("Authorization")).isEqualTo("Bearer secret");
        }
    }

    @Test
    void streamLogsThrowsWhenRequestFailsBeforeTerminalState() {
        assertThatThrownBy(() -> new BuildLogStreamer(new okhttp3.OkHttpClient(), "secret")
                .streamLogs("http://127.0.0.1:1/logs", line -> { }, () -> false))
                .isInstanceOf(DaytonaException.class)
                .hasMessageContaining("Failed to stream build logs");
    }

    @Test
    void streamLogsAddsFollowParameterWithoutExistingQuery() throws Exception {
        try (MockWebServer server = new MockWebServer()) {
            server.enqueue(new MockResponse().setBody("line-1\n"));

            new BuildLogStreamer(new okhttp3.OkHttpClient(), "secret")
                    .streamLogs(server.url("/logs").toString(), line -> { }, () -> false);

            assertThat(server.takeRequest().getPath()).isEqualTo("/logs?follow=true");
        }
    }

    @Test
    void streamLogsReturnsWhenResponseBodyMissing() throws Exception {
        try (MockWebServer server = new MockWebServer()) {
            server.enqueue(new MockResponse().setResponseCode(204));
            List<String> lines = new ArrayList<String>();

            new BuildLogStreamer(new okhttp3.OkHttpClient(), "secret")
                    .streamLogs(server.url("/logs").toString(), lines::add, () -> false);

            assertThat(lines).isEmpty();
        }
    }

    @Test
    void streamLogsSuppressesFailuresAfterTerminalState() {
        new BuildLogStreamer(new okhttp3.OkHttpClient(), "secret")
                .streamLogs("http://127.0.0.1:1/logs", line -> { }, () -> true);
    }
}
