// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk;

public final class DaytonaConfig {
    private final String apiKey;
    private final String apiUrl;
    private final String target;

    private DaytonaConfig(Builder builder) {
        this.apiKey = builder.apiKey;
        this.apiUrl = builder.apiUrl == null || builder.apiUrl.isEmpty()
                ? "https://app.daytona.io/api"
                : builder.apiUrl;
        this.target = builder.target;
    }

    public String getApiKey() {
        return apiKey;
    }

    public String getApiUrl() {
        return apiUrl;
    }

    public String getTarget() {
        return target;
    }

    public static class Builder {
        private String apiKey;
        private String apiUrl;
        private String target;

        public Builder apiKey(String apiKey) {
            this.apiKey = apiKey;
            return this;
        }

        public Builder apiUrl(String apiUrl) {
            this.apiUrl = apiUrl;
            return this;
        }

        public Builder target(String target) {
            this.target = target;
            return this;
        }

        public DaytonaConfig build() {
            return new DaytonaConfig(this);
        }
    }
}