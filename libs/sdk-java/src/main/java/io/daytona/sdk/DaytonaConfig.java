// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk;

/**
 * Configuration used to initialize a {@link Daytona} client.
 *
 * <p>Contains API authentication settings, API endpoint URL, and the default target region used
 * when creating new Sandboxes.
 */
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

    /**
     * Returns the API key used to authenticate SDK requests.
     *
     * @return API key configured for the client
     */
    public String getApiKey() {
        return apiKey;
    }

    /**
     * Returns the Daytona API base URL.
     *
     * @return API URL used for main API requests
     */
    public String getApiUrl() {
        return apiUrl;
    }

    /**
     * Returns the default target location for newly created Sandboxes.
     *
     * @return target region identifier, or {@code null} if not configured
     */
    public String getTarget() {
        return target;
    }

    /**
     * Builder for creating immutable {@link DaytonaConfig} instances.
     */
    public static class Builder {
        private String apiKey;
        private String apiUrl;
        private String target;

        /**
         * Sets the API key used for authenticating SDK requests.
         *
         * @param apiKey Daytona API key
         * @return this builder instance
         */
        public Builder apiKey(String apiKey) {
            this.apiKey = apiKey;
            return this;
        }

        /**
         * Sets the Daytona API base URL.
         *
         * @param apiUrl API URL to use; defaults to {@code https://app.daytona.io/api} when omitted
         * @return this builder instance
         */
        public Builder apiUrl(String apiUrl) {
            this.apiUrl = apiUrl;
            return this;
        }

        /**
         * Sets the default target region for new Sandboxes.
         *
         * @param target target location identifier
         * @return this builder instance
         */
        public Builder target(String target) {
            this.target = target;
            return this;
        }

        /**
         * Builds a new immutable {@link DaytonaConfig}.
         *
         * @return configured {@link DaytonaConfig} instance
         */
        public DaytonaConfig build() {
            return new DaytonaConfig(this);
        }
    }
}
