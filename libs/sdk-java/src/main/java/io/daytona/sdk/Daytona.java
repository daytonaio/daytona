// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import io.daytona.api.client.api.SandboxApi;
import io.daytona.api.client.model.CreateBuildInfo;
import io.daytona.api.client.model.CreateSandbox;
import io.daytona.api.client.model.SandboxVolume;
import io.daytona.sdk.exception.DaytonaException;
import io.daytona.sdk.model.CreateSandboxFromImageParams;
import io.daytona.sdk.model.CreateSandboxFromSnapshotParams;
import io.daytona.sdk.model.PaginatedSandboxes;
import okhttp3.OkHttpClient;

import java.net.URLEncoder;
import java.nio.charset.StandardCharsets;
import java.math.BigDecimal;
import java.util.ArrayList;
import java.util.Collections;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.lang.reflect.Field;

/**
 * Main class for interacting with the Daytona API.
 *
 * <p>Provides methods to create, retrieve, and list Sandboxes, and exposes service accessors for
 * Snapshots and Volumes.
 *
 * <p>Implements {@link AutoCloseable} for deterministic HTTP resource cleanup.
 *
 * @see DaytonaConfig
 * @see Sandbox
 */
public class Daytona implements AutoCloseable {
    public static final String CODE_TOOLBOX_LANGUAGE_LABEL = "code-toolbox-language";

    private final DaytonaConfig config;
    private final io.daytona.api.client.ApiClient apiClient;
    private final SandboxApi sandboxApi;
    private final SnapshotService snapshot;
    private final VolumeService volume;
    private final ObjectMapper objectMapper = new ObjectMapper();

    /**
     * Creates a client using environment variables.
     *
     * <p>Reads {@code DAYTONA_API_KEY}, {@code DAYTONA_API_URL}, and {@code DAYTONA_TARGET}.
     *
     * @throws DaytonaException if required authentication is missing
     */
    public Daytona() {
        this(new DaytonaConfig.Builder()
                .apiKey(System.getenv("DAYTONA_API_KEY"))
                .apiUrl(envOrDefault("DAYTONA_API_URL", "https://app.daytona.io/api"))
                .target(System.getenv("DAYTONA_TARGET"))
                .build());
    }

    /**
     * Creates a client with explicit configuration.
     *
     * @param config SDK configuration containing API key and endpoint settings
     * @throws DaytonaException if configuration is invalid or missing credentials
     */
    public Daytona(DaytonaConfig config) {
        if (config == null || config.getApiKey() == null || config.getApiKey().isEmpty()) {
            throw new DaytonaException("DAYTONA_API_KEY is required");
        }
        this.config = config;
        this.apiClient = createMainApiClient(config);
        this.sandboxApi = new SandboxApi(apiClient);
        this.snapshot = new SnapshotService(new io.daytona.api.client.api.SnapshotsApi(apiClient), apiClient.getHttpClient(), config.getApiKey());
        this.volume = new VolumeService(new io.daytona.api.client.api.VolumesApi(apiClient));
    }

    /**
     * Creates a Sandbox with default parameters and timeout.
     *
     * @return created and started {@link Sandbox}
     * @throws DaytonaException if creation or startup fails
     */
    public Sandbox create() {
        return create(new CreateSandboxFromSnapshotParams(), 60);
    }

    /**
     * Creates a Sandbox from snapshot-oriented parameters using default timeout.
     *
     * @param params snapshot creation parameters
     * @return created and started {@link Sandbox}
     * @throws DaytonaException if creation or startup fails
     */
    public Sandbox create(CreateSandboxFromSnapshotParams params) {
        return create(params, 60);
    }

    /**
     * Creates a Sandbox from image-oriented parameters using default timeout.
     *
     * @param params image creation parameters
     * @return created and started {@link Sandbox}
     * @throws DaytonaException if creation or startup fails
     */
    public Sandbox create(CreateSandboxFromImageParams params) {
        return create(params, 60);
    }

    /**
     * Creates a Sandbox from snapshot parameters.
     *
     * @param params snapshot creation parameters including env vars, labels, and lifecycle options
     * @param timeoutSeconds maximum seconds to wait for the Sandbox to reach {@code started}
     * @return created and started {@link Sandbox}
     * @throws DaytonaException if creation fails or the Sandbox does not start in time
     */
    public Sandbox create(CreateSandboxFromSnapshotParams params, long timeoutSeconds) {
        CreateSandbox body = baseSandboxBody(params);
        if (params != null && params.getSnapshot() != null && !params.getSnapshot().isEmpty()) {
            body.setSnapshot(params.getSnapshot());
        }
        io.daytona.api.client.model.Sandbox response = ExceptionMapper.callMain(() -> sandboxApi.createSandbox(body, null));
        Sandbox sandbox = new Sandbox(sandboxApi, config, response);
        sandbox.waitUntilStarted(timeoutSeconds);
        return sandbox;
    }

    /**
     * Creates a Sandbox from image parameters.
     *
     * @param params image creation parameters including image source and optional resources
     * @param timeoutSeconds maximum seconds to wait for the Sandbox to reach {@code started}
     * @return created and started {@link Sandbox}
     * @throws DaytonaException if creation fails or the Sandbox does not start in time
     */
    public Sandbox create(CreateSandboxFromImageParams params, long timeoutSeconds) {
        return create(params, timeoutSeconds, null);
    }

    /**
     * Creates a new Sandbox from a declarative image with build log streaming.
     *
     * @param params creation parameters including the image definition
     * @param timeoutSeconds maximum seconds to wait for the Sandbox to reach {@code started}
     * @param onSnapshotCreateLogs callback for build log lines; {@code null} to skip streaming
     * @return created and started {@link Sandbox}
     * @throws DaytonaException if creation fails or the Sandbox does not start in time
     */
    public Sandbox create(CreateSandboxFromImageParams params, long timeoutSeconds, java.util.function.Consumer<String> onSnapshotCreateLogs) {
        CreateSandbox body = baseSandboxBody(params);
        if (params != null) {
            Object image = params.getImage();
            if (image instanceof Image) {
                body.setBuildInfo(new CreateBuildInfo().dockerfileContent(((Image) image).getDockerfile()));
            } else if (image instanceof String && !((String) image).isEmpty()) {
                body.setBuildInfo(new CreateBuildInfo().dockerfileContent("FROM " + image + "\n"));
            }

            if (params.getResources() != null) {
                if (params.getResources().getCpu() != null) body.setCpu(params.getResources().getCpu());
                if (params.getResources().getGpu() != null) body.setGpu(params.getResources().getGpu());
                if (params.getResources().getMemory() != null) body.setMemory(params.getResources().getMemory());
                if (params.getResources().getDisk() != null) body.setDisk(params.getResources().getDisk());
            }
        }

        long startTime = System.currentTimeMillis();
        io.daytona.api.client.model.Sandbox response = ExceptionMapper.callMain(() -> sandboxApi.createSandbox(body, null));

        String initialState = response.getState() != null ? response.getState().getValue() : "";
        if (onSnapshotCreateLogs != null && "pending_build".equals(initialState)) {
            waitForBuildState(response.getId(), timeoutSeconds, startTime);
            streamSandboxBuildLogs(response.getId(), onSnapshotCreateLogs, timeoutSeconds, startTime);
        }

        Sandbox sandbox = new Sandbox(sandboxApi, config,
                ExceptionMapper.callMain(() -> sandboxApi.getSandbox(response.getId(), null, null)));
        long elapsed = (System.currentTimeMillis() - startTime) / 1000;
        long remaining = timeoutSeconds > 0 ? Math.max(1, timeoutSeconds - elapsed) : timeoutSeconds;
        sandbox.waitUntilStarted(remaining);
        return sandbox;
    }

    /**
     * Retrieves a Sandbox by ID or name.
     *
     * @param sandboxIdOrName Sandbox identifier or name
     * @return resolved {@link Sandbox}
     * @throws DaytonaException if the Sandbox is not found or request fails
     */
    public Sandbox get(String sandboxIdOrName) {
        io.daytona.api.client.model.Sandbox response = ExceptionMapper.callMain(() -> sandboxApi.getSandbox(sandboxIdOrName, null, null));
        return new Sandbox(sandboxApi, config, response);
    }

    /**
     * Lists Sandboxes using default pagination.
     *
     * @return first page of Sandboxes with default page size
     * @throws DaytonaException if listing fails
     */
    public PaginatedSandboxes list() {
        return list(null, 1, 10);
    }

    /**
     * Lists Sandboxes with optional label filtering and pagination.
     *
     * @param labels label filter map; only Sandboxes with matching labels are returned
     * @param page page number starting from 1
     * @param limit maximum items per page
     * @return paginated Sandbox list
     * @throws DaytonaException if listing fails
     */
    public PaginatedSandboxes list(Map<String, String> labels, Integer page, Integer limit) {
        int p = page == null ? 1 : page;
        int l = limit == null ? 10 : limit;
        String path = "/sandbox/paginated?page=" + p + "&limit=" + l;
        if (labels != null && !labels.isEmpty()) {
            path = path + "&labels=" + urlEncodeQuery(toJson(labels));
        }

        io.daytona.api.client.model.PaginatedSandboxes result = ExceptionMapper.callMain(() -> sandboxApi.listSandboxesPaginated(
                null,
                BigDecimal.valueOf(p),
                BigDecimal.valueOf(l),
                null,
                null,
                labels == null || labels.isEmpty() ? null : toJson(labels),
                null,
                null,
                null,
                null,
                null,
                null,
                null,
                null,
                null,
                null,
                null,
                null,
                null,
                null
        ));

        PaginatedSandboxes paginated = new PaginatedSandboxes();
        List<Map<String, Object>> items = new ArrayList<Map<String, Object>>();
        if (result != null && result.getItems() != null) {
            for (io.daytona.api.client.model.Sandbox item : result.getItems()) {
                items.add(sandboxToMap(item));
            }
        }
        paginated.setItems(items);
        paginated.setTotal(result != null && result.getTotal() != null ? result.getTotal().intValue() : 0);
        paginated.setPage(result != null && result.getPage() != null ? result.getPage().intValue() : 0);
        paginated.setTotalPages(result != null && result.getTotalPages() != null ? result.getTotalPages().intValue() : 0);
        return paginated;
    }

    /**
     * Returns Snapshot management service.
     *
     * @return snapshot service instance
     */
    public SnapshotService snapshot() {
        return snapshot;
    }

    /**
     * Returns Volume management service.
     *
     * @return volume service instance
     */
    public VolumeService volume() {
        return volume;
    }

    /**
     * Closes this client and releases underlying HTTP resources.
     */
    @Override
    public void close() {
        shutdownHttpClient(apiClient.getHttpClient());
    }

    private CreateSandbox baseSandboxBody(io.daytona.sdk.model.CreateSandboxParams params) {
        CreateSandbox body = new CreateSandbox();
        if (params == null) {
            if (config.getTarget() != null && !config.getTarget().isEmpty()) {
                body.setTarget(config.getTarget());
            }
            return body;
        }

        if (params.getName() != null) body.setName(params.getName());
        if (params.getUser() != null) body.setUser(params.getUser());
        if (params.getEnvVars() != null) body.setEnv(params.getEnvVars());
        if (params.getPublic() != null) body.setPublic(params.getPublic());
        if (params.getAutoStopInterval() != null) body.setAutoStopInterval(params.getAutoStopInterval());
        if (params.getAutoArchiveInterval() != null) body.setAutoArchiveInterval(params.getAutoArchiveInterval());
        if (params.getAutoDeleteInterval() != null) body.setAutoDeleteInterval(params.getAutoDeleteInterval());
        if (params.getNetworkBlockAll() != null) body.setNetworkBlockAll(params.getNetworkBlockAll());
        if (params.getVolumes() != null) {
            List<SandboxVolume> volumes = new ArrayList<SandboxVolume>();
            for (io.daytona.sdk.model.VolumeMount mount : params.getVolumes()) {
                volumes.add(new SandboxVolume().volumeId(mount.getVolumeId()).mountPath(mount.getMountPath()));
            }
            body.setVolumes(volumes);
        }

        Map<String, String> labels = params.getLabels() == null
                ? new HashMap<String, String>()
                : new HashMap<String, String>(params.getLabels());
        String language = params.getLanguage();
        if (language == null || language.isEmpty()) {
            language = CodeLanguage.PYTHON.getValue();
        }
        language = CodeLanguage.fromValue(language).getValue();
        labels.put(CODE_TOOLBOX_LANGUAGE_LABEL, language);
        if (!labels.isEmpty()) {
            body.setLabels(labels);
        }

        if (config.getTarget() != null && !config.getTarget().isEmpty()) {
            body.setTarget(config.getTarget());
        }

        return body;
    }

    private io.daytona.api.client.ApiClient createMainApiClient(DaytonaConfig cfg) {
        io.daytona.api.client.ApiClient client = new io.daytona.api.client.ApiClient();
        client.setBasePath(trimTrailingSlash(cfg.getApiUrl()));
        client.setBearerToken(cfg.getApiKey());
        String sdkVersion = Daytona.class.getPackage().getImplementationVersion();
        if (sdkVersion == null) sdkVersion = "dev";
        client.addDefaultHeader("X-Daytona-Source", "sdk-java");
        client.addDefaultHeader("X-Daytona-SDK-Version", sdkVersion);
        client.setUserAgent("sdk-java/" + sdkVersion);
        ensureOauth2AuthEntry(client);
        return client;
    }

    @SuppressWarnings("unchecked")
    private void ensureOauth2AuthEntry(io.daytona.api.client.ApiClient client) {
        if (client.getAuthentications().containsKey("oauth2")) {
            return;
        }
        try {
            Field authField = io.daytona.api.client.ApiClient.class.getDeclaredField("authentications");
            authField.setAccessible(true);
            Map<String, io.daytona.api.client.auth.Authentication> auths =
                    new HashMap<String, io.daytona.api.client.auth.Authentication>(
                            (Map<String, io.daytona.api.client.auth.Authentication>) authField.get(client)
                    );
            auths.put("oauth2", (q, h, c, p, m, u) -> {
            });
            authField.set(client, Collections.unmodifiableMap(auths));
        } catch (ReflectiveOperationException e) {
            throw new DaytonaException("Failed to initialize authentication", e);
        }
    }

    static void shutdownHttpClient(OkHttpClient client) {
        if (client == null) {
            return;
        }
        client.dispatcher().executorService().shutdown();
        client.connectionPool().evictAll();
        if (client.cache() != null) {
            try {
                client.cache().close();
            } catch (Exception ignored) {
            }
        }
    }

    private static String trimTrailingSlash(String value) {
        if (value == null) {
            return "";
        }
        String output = value;
        while (output.endsWith("/")) {
            output = output.substring(0, output.length() - 1);
        }
        return output;
    }

    static Map<String, Object> sandboxToMap(io.daytona.api.client.model.Sandbox sandbox) {
        Map<String, Object> map = new HashMap<String, Object>();
        if (sandbox == null) {
            return map;
        }
        map.put("id", sandbox.getId());
        map.put("name", sandbox.getName());
        map.put("state", sandbox.getState() == null ? null : sandbox.getState().getValue());
        map.put("target", sandbox.getTarget());
        map.put("user", sandbox.getUser());
        map.put("toolboxProxyUrl", sandbox.getToolboxProxyUrl());
        map.put("cpu", sandbox.getCpu() == null ? 0 : sandbox.getCpu().intValue());
        map.put("gpu", sandbox.getGpu() == null ? 0 : sandbox.getGpu().intValue());
        map.put("memory", sandbox.getMemory() == null ? 0 : sandbox.getMemory().intValue());
        map.put("disk", sandbox.getDisk() == null ? 0 : sandbox.getDisk().intValue());
        map.put("env", sandbox.getEnv());
        map.put("labels", sandbox.getLabels());
        map.put("autoStopInterval", sandbox.getAutoStopInterval() == null ? null : sandbox.getAutoStopInterval().intValue());
        map.put("autoArchiveInterval", sandbox.getAutoArchiveInterval() == null ? null : sandbox.getAutoArchiveInterval().intValue());
        map.put("autoDeleteInterval", sandbox.getAutoDeleteInterval() == null ? null : sandbox.getAutoDeleteInterval().intValue());
        return map;
    }

    static String urlEncodePathSegment(String value) {
        return URLEncoder.encode(value, StandardCharsets.UTF_8).replace("+", "%20");
    }

    static String urlEncodeQuery(String value) {
        return URLEncoder.encode(value == null ? "" : value, StandardCharsets.UTF_8);
    }

    static Map<String, String> castStringMap(Map<?, ?> source) {
        Map<String, String> out = new HashMap<String, String>();
        if (source == null) {
            return out;
        }
        for (Map.Entry<?, ?> e : source.entrySet()) {
            out.put(String.valueOf(e.getKey()), e.getValue() == null ? "" : String.valueOf(e.getValue()));
        }
        return out;
    }

    private static String envOrDefault(String key, String fallback) {
        String value = System.getenv(key);
        return value == null || value.isEmpty() ? fallback : value;
    }

    private String toJson(Object value) {
        try {
            return objectMapper.writeValueAsString(value);
        } catch (JsonProcessingException e) {
            throw new DaytonaException("Failed to serialize JSON", e);
        }
    }

    private void waitForBuildState(String sandboxId, long timeoutSeconds, long startTime) {
        while (true) {
            if (timeoutSeconds > 0) {
                long elapsed = (System.currentTimeMillis() - startTime) / 1000;
                if (elapsed > timeoutSeconds) {
                    throw new DaytonaException("Sandbox build pending for more than " + timeoutSeconds + " seconds");
                }
            }
            io.daytona.api.client.model.Sandbox s = ExceptionMapper.callMain(() -> sandboxApi.getSandbox(sandboxId, null, null));
            String state = s.getState() != null ? s.getState().getValue() : "";
            if (!"pending_build".equals(state)) return;
            try { Thread.sleep(1000); } catch (InterruptedException e) { Thread.currentThread().interrupt(); return; }
        }
    }

    private void streamSandboxBuildLogs(String sandboxId, java.util.function.Consumer<String> onLog, long timeoutSeconds, long startTime) {
        io.daytona.api.client.model.Url logsUrl = ExceptionMapper.callMain(() -> sandboxApi.getBuildLogsUrl(sandboxId, null));
        BuildLogStreamer streamer = new BuildLogStreamer(apiClient.getHttpClient(), config.getApiKey());
        streamer.streamLogs(logsUrl.getUrl(), onLog, () -> {
            io.daytona.api.client.model.Sandbox s = ExceptionMapper.callMain(() -> sandboxApi.getSandbox(sandboxId, null, null));
            String state = s.getState() != null ? s.getState().getValue() : "";
            return "started".equals(state) || "starting".equals(state) || "error".equals(state) || "build_failed".equals(state);
        });
    }
}
