// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk;

import io.daytona.sdk.exception.DaytonaException;
import org.junit.jupiter.api.Test;

import java.util.HashMap;
import java.util.Map;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;

class DaytonaConfigTest {

    @Test
    void builderStoresExplicitValues() {
        DaytonaConfig config = new DaytonaConfig.Builder()
                .apiKey("key")
                .apiUrl("https://custom/api")
                .target("us")
                .build();

        assertThat(config.getApiKey()).isEqualTo("key");
        assertThat(config.getApiUrl()).isEqualTo("https://custom/api");
        assertThat(config.getTarget()).isEqualTo("us");
    }

    @Test
    void builderUsesDefaultApiUrlWhenNull() {
        DaytonaConfig config = new DaytonaConfig.Builder()
                .apiKey("key")
                .apiUrl(null)
                .build();

        assertThat(config.getApiUrl()).isEqualTo("https://app.daytona.io/api");
    }

    @Test
    void builderUsesDefaultApiUrlWhenEmpty() {
        DaytonaConfig config = new DaytonaConfig.Builder()
                .apiKey("key")
                .apiUrl("")
                .build();

        assertThat(config.getApiUrl()).isEqualTo("https://app.daytona.io/api");
    }

    @Test
    void builderAllowsNullTargetAndApiKey() {
        DaytonaConfig config = new DaytonaConfig.Builder().build();

        assertThat(config.getApiKey()).isNull();
        assertThat(config.getTarget()).isNull();
        assertThat(config.getApiUrl()).isEqualTo("https://app.daytona.io/api");
    }

    @Test
    void defaultDaytonaConstructorReadsEnvironmentVariables() throws Exception {
        Map<String, String> env = new HashMap<String, String>();
        env.put("DAYTONA_API_KEY", "env-key");
        env.put("DAYTONA_API_URL", "https://env.example/api/");
        env.put("DAYTONA_TARGET", "eu");

        TestSupport.withEnvironment(env, () -> {
            try (Daytona daytona = new Daytona()) {
                DaytonaConfig config = TestSupport.getField(daytona, "config", DaytonaConfig.class);
                assertThat(config.getApiKey()).isEqualTo("env-key");
                assertThat(config.getApiUrl()).isEqualTo("https://env.example/api/");
                assertThat(config.getTarget()).isEqualTo("eu");
            }
        });
    }

    @Test
    void defaultDaytonaConstructorFallsBackToDefaultApiUrl() throws Exception {
        Map<String, String> env = new HashMap<String, String>();
        env.put("DAYTONA_API_KEY", "env-key");
        env.put("DAYTONA_API_URL", null);
        env.put("DAYTONA_TARGET", null);

        TestSupport.withEnvironment(env, () -> {
            try (Daytona daytona = new Daytona()) {
                DaytonaConfig config = TestSupport.getField(daytona, "config", DaytonaConfig.class);
                assertThat(config.getApiUrl()).isEqualTo("https://app.daytona.io/api");
                assertThat(config.getTarget()).isNull();
            }
        });
    }

    @Test
    void defaultDaytonaConstructorUsesFallbackWhenApiUrlEnvIsEmpty() throws Exception {
        Map<String, String> env = new HashMap<String, String>();
        env.put("DAYTONA_API_KEY", "env-key");
        env.put("DAYTONA_API_URL", "");

        TestSupport.withEnvironment(env, () -> {
            try (Daytona daytona = new Daytona()) {
                DaytonaConfig config = TestSupport.getField(daytona, "config", DaytonaConfig.class);
                assertThat(config.getApiUrl()).isEqualTo("https://app.daytona.io/api");
            }
        });
    }

    @Test
    void defaultDaytonaConstructorRequiresApiKey() throws Exception {
        Map<String, String> env = new HashMap<String, String>();
        env.put("DAYTONA_API_KEY", null);
        env.put("DAYTONA_API_URL", null);
        env.put("DAYTONA_TARGET", null);

        TestSupport.withEnvironment(env, () -> assertThatThrownBy(Daytona::new)
                .isInstanceOf(DaytonaException.class)
                .hasMessage("DAYTONA_API_KEY is required"));
    }
}
