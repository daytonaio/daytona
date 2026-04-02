// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk;

import io.daytona.sdk.exception.DaytonaException;
import okhttp3.OkHttpClient;
import okhttp3.Request;
import okhttp3.Response;

import java.io.BufferedReader;
import java.io.InputStreamReader;
import java.nio.charset.StandardCharsets;
import java.util.concurrent.TimeUnit;
import java.util.function.Consumer;
import java.util.function.Supplier;

class BuildLogStreamer {

    private final OkHttpClient httpClient;
    private final String authToken;

    BuildLogStreamer(OkHttpClient httpClient, String authToken) {
        this.httpClient = httpClient.newBuilder()
                .readTimeout(5, TimeUnit.MINUTES)
                .build();
        this.authToken = authToken;
    }

    void streamLogs(String logsUrl, Consumer<String> onLog, Supplier<Boolean> isTerminal) {
        String url = logsUrl + (logsUrl.contains("?") ? "&" : "?") + "follow=true";
        Request request = new Request.Builder()
                .url(url)
                .addHeader("Authorization", "Bearer " + authToken)
                .build();

        try (Response response = httpClient.newCall(request).execute()) {
            if (response.body() == null) return;
            BufferedReader reader = new BufferedReader(
                    new InputStreamReader(response.body().byteStream(), StandardCharsets.UTF_8));
            String line;
            while ((line = reader.readLine()) != null) {
                onLog.accept(line);
                if (isTerminal.get()) break;
            }
        } catch (java.io.IOException e) {
            if (!isTerminal.get()) {
                throw new DaytonaException("Failed to stream build logs", e);
            }
        }
    }
}
