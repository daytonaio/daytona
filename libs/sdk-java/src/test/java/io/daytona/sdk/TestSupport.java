// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk;

import io.daytona.api.client.model.SandboxState;
import okhttp3.OkHttpClient;
import org.mockito.Mockito;

import java.lang.reflect.Field;
import java.math.BigDecimal;
import java.util.HashMap;
import java.util.Map;

final class TestSupport {
    private TestSupport() {
    }

    static io.daytona.api.client.model.Sandbox mainSandbox(String id, SandboxState state) {
        io.daytona.api.client.model.Sandbox sandbox = new io.daytona.api.client.model.Sandbox();
        sandbox.setId(id);
        sandbox.setOrganizationId("org-1");
        sandbox.setName("sandbox-" + id);
        sandbox.setUser("daytona");
        sandbox.setEnv(new HashMap<String, String>());
        sandbox.setLabels(new HashMap<String, String>());
        sandbox.setPublic(false);
        sandbox.setNetworkBlockAll(false);
        sandbox.setTarget("us");
        sandbox.setCpu(BigDecimal.ONE);
        sandbox.setGpu(BigDecimal.ZERO);
        sandbox.setMemory(BigDecimal.valueOf(2));
        sandbox.setDisk(BigDecimal.valueOf(3));
        sandbox.setState(state);
        sandbox.setToolboxProxyUrl("http://localhost:1/toolbox");
        return sandbox;
    }

    static io.daytona.api.client.model.Sandbox sandboxWithToolboxUrl(String id, SandboxState state, String toolboxUrl) {
        io.daytona.api.client.model.Sandbox sandbox = mainSandbox(id, state);
        sandbox.setToolboxProxyUrl(toolboxUrl);
        return sandbox;
    }

    static DaytonaConfig config() {
        return new DaytonaConfig.Builder()
                .apiKey("test-key")
                .apiUrl("https://example.com/api/")
                .target("eu")
                .build();
    }

    static void setField(Object target, String fieldName, Object value) {
        try {
            Field field = findField(target.getClass(), fieldName);
            field.setAccessible(true);
            field.set(target, value);
        } catch (Exception e) {
            throw new RuntimeException(e);
        }
    }

    @SuppressWarnings("unchecked")
    static <T> T getField(Object target, String fieldName, Class<T> type) {
        try {
            Field field = findField(target.getClass(), fieldName);
            field.setAccessible(true);
            return (T) field.get(target);
        } catch (Exception e) {
            throw new RuntimeException(e);
        }
    }

    static void withEnvironment(Map<String, String> updates, ThrowingRunnable runnable) throws Exception {
        Map<String, String> env = writableEnv();
        Map<String, String> original = new HashMap<String, String>(env);
        try {
            for (Map.Entry<String, String> entry : updates.entrySet()) {
                if (entry.getValue() == null) {
                    env.remove(entry.getKey());
                } else {
                    env.put(entry.getKey(), entry.getValue());
                }
            }
            runnable.run();
        } finally {
            env.clear();
            env.putAll(original);
        }
    }

    @SuppressWarnings("unchecked")
    private static Map<String, String> writableEnv() throws Exception {
        Map<String, String> env = System.getenv();
        Class<?> type = env.getClass();
        if (type.getName().equals("java.util.Collections$UnmodifiableMap")) {
            Field field = type.getDeclaredField("m");
            field.setAccessible(true);
            return (Map<String, String>) field.get(env);
        }
        return env;
    }

    static Sandbox mockSandbox(String basePath) {
        return mockSandbox(basePath, "python", "test-key", new OkHttpClient());
    }

    static Sandbox mockSandbox(String basePath, String language, String apiKey, OkHttpClient httpClient) {
        io.daytona.toolbox.client.ApiClient apiClient = new io.daytona.toolbox.client.ApiClient();
        apiClient.setBasePath(basePath);
        apiClient.setHttpClient(httpClient);
        Sandbox sandbox = Mockito.mock(Sandbox.class);
        Mockito.lenient().when(sandbox.getLanguage()).thenReturn(language);
        Mockito.lenient().when(sandbox.getApiKey()).thenReturn(apiKey);
        Mockito.lenient().when(sandbox.getToolboxApiClient()).thenReturn(apiClient);
        return sandbox;
    }

    private static Field findField(Class<?> type, String fieldName) throws NoSuchFieldException {
        Class<?> current = type;
        while (current != null) {
            try {
                return current.getDeclaredField(fieldName);
            } catch (NoSuchFieldException ignored) {
                current = current.getSuperclass();
            }
        }
        throw new NoSuchFieldException(fieldName);
    }

    @FunctionalInterface
    interface ThrowingRunnable {
        void run() throws Exception;
    }
}
