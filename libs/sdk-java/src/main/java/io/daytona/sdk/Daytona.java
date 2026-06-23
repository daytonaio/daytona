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
import io.daytona.sdk.model.ListSandboxesQuery;
import java.math.BigDecimal;
import java.time.OffsetDateTime;
import java.util.ArrayList;
import java.util.Iterator;
import java.util.List;
import java.util.Map;
import java.util.NoSuchElementException;
import java.util.Spliterator;
import java.util.Spliterators;
import java.util.stream.Stream;
import java.util.stream.StreamSupport;
import okhttp3.OkHttpClient;

import java.net.URLEncoder;
import java.nio.charset.StandardCharsets;
import java.util.Collections;
import java.util.HashMap;
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
            throw new DaytonaException(
                    "Authentication required: set DAYTONA_API_KEY environment variable or pass apiKey in DaytonaConfig");
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
                if (params.getResources().getGpuType() != null) body.setGpuType(params.getResources().getGpuType());
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
     * Iterates over all Sandboxes (no filter, default sort).
     *
     * <p>Returns a lazily-paged {@link Iterable}; see {@link #list(ListSandboxesQuery)} for details
     * on partial hydration and Stream usage.
     *
     * @return iterable over Sandboxes
     */
    public Iterable<Sandbox> list() {
        return list(null);
    }

    /**
     * Iterates over Sandboxes matching the given query.
     *
     * <p>The returned {@link Iterable} lazily fetches pages from the API as iteration proceeds.
     * Sandboxes are hydrated from the list endpoint, so fields marked "Not returned by
     * {@code Daytona.list}" on {@link Sandbox} (env, networkBlockAll, networkAllowList, volumes,
     * buildInfo, backupCreatedAt) remain {@code null} until {@link Sandbox#refreshData()} is called.
     * For a {@link Stream} variant see {@link #listStream(ListSandboxesQuery)}.
     *
     * <pre>{@code
     * ListSandboxesQuery query = new ListSandboxesQuery();
     * query.setLabels(Map.of("env", "dev"));
     * for (Sandbox sandbox : daytona.list(query)) {
     *     System.out.println(sandbox.getId());
     * }
     * }</pre>
     *
     * @param query optional filters, sorting, and per-page size
     * @return iterable over Sandboxes
     */
    public Iterable<Sandbox> list(ListSandboxesQuery query) {
        return () -> new SandboxIterator(this, query);
    }

    /**
     * Streams all Sandboxes (no filter, default sort).
     *
     * <p>The returned stream should be closed (use try-with-resources).
     *
     * @return stream of Sandboxes
     * @see #list()
     */
    public Stream<Sandbox> listStream() {
        return listStream(null);
    }

    /**
     * Streams Sandboxes matching the given query.
     *
     * <p>The returned stream should be closed (use try-with-resources).
     *
     * <pre>{@code
     * try (Stream<Sandbox> stream = daytona.listStream(query)) {
     *     stream.filter(sb -> "started".equals(sb.getState()))
     *           .limit(5)
     *           .forEach(sb -> System.out.println(sb.getId()));
     * }
     * }</pre>
     *
     * @param query optional filters, sorting, and per-page size
     * @return stream of Sandboxes
     */
    public Stream<Sandbox> listStream(ListSandboxesQuery query) {
        Iterator<Sandbox> iter = list(query).iterator();
        Spliterator<Sandbox> spliterator = Spliterators.spliteratorUnknownSize(
                iter, Spliterator.ORDERED | Spliterator.NONNULL);
        return StreamSupport.stream(spliterator, false);
    }

    /**
     * Fetches a single page of Sandboxes. Package-private so {@link SandboxIterator}
     * can call it directly. Each call results in one outbound API request.
     */
    PageResult fetchSandboxPage(ListSandboxesQuery query, String cursor) {
        String labelsJson = null;
        BigDecimal limitVal = null;
        String id = null;
        String name = null;
        List<io.daytona.api.client.model.SandboxState> states = null;
        List<String> snapshots = null;
        List<String> targets = null;
        BigDecimal minCpu = null;
        BigDecimal maxCpu = null;
        BigDecimal minMemoryGib = null;
        BigDecimal maxMemoryGib = null;
        BigDecimal minDiskGib = null;
        BigDecimal maxDiskGib = null;
        Boolean isPublic = null;
        Boolean isRecoverable = null;
        OffsetDateTime createdAtAfter = null;
        OffsetDateTime createdAtBefore = null;
        OffsetDateTime lastActivityAfter = null;
        OffsetDateTime lastActivityBefore = null;
        io.daytona.api.client.model.SandboxListSortField sort = null;
        io.daytona.api.client.model.SandboxListSortDirection order = null;

        if (query != null) {
            if (query.getLimit() != null) limitVal = BigDecimal.valueOf(query.getLimit());
            id = query.getId();
            name = query.getName();
            if (query.getLabels() != null && !query.getLabels().isEmpty()) {
                labelsJson = toJson(query.getLabels());
            }
            if (query.getStates() != null) {
                states = new ArrayList<>(query.getStates().size());
                for (io.daytona.sdk.model.SandboxState s : query.getStates()) {
                    states.add(s.toApiClient());
                }
            }
            snapshots = query.getSnapshots();
            targets = query.getTargets();
            if (query.getMinCpu() != null) minCpu = BigDecimal.valueOf(query.getMinCpu());
            if (query.getMaxCpu() != null) maxCpu = BigDecimal.valueOf(query.getMaxCpu());
            if (query.getMinMemoryGib() != null) minMemoryGib = BigDecimal.valueOf(query.getMinMemoryGib());
            if (query.getMaxMemoryGib() != null) maxMemoryGib = BigDecimal.valueOf(query.getMaxMemoryGib());
            if (query.getMinDiskGib() != null) minDiskGib = BigDecimal.valueOf(query.getMinDiskGib());
            if (query.getMaxDiskGib() != null) maxDiskGib = BigDecimal.valueOf(query.getMaxDiskGib());
            isPublic = query.getIsPublic();
            isRecoverable = query.getIsRecoverable();
            createdAtAfter = query.getCreatedAtAfter();
            createdAtBefore = query.getCreatedAtBefore();
            lastActivityAfter = query.getLastActivityAfter();
            lastActivityBefore = query.getLastActivityBefore();
            if (query.getSort() != null) sort = query.getSort().toApiClient();
            if (query.getOrder() != null) order = query.getOrder().toApiClient();
        }

        final String fLabelsJson = labelsJson;
        final String fCursor = cursor;
        final BigDecimal fLimitVal = limitVal;
        final String fId = id;
        final String fName = name;
        final List<io.daytona.api.client.model.SandboxState> fStates = states;
        final List<String> fSnapshots = snapshots;
        final List<String> fTargets = targets;
        final BigDecimal fMinCpu = minCpu;
        final BigDecimal fMaxCpu = maxCpu;
        final BigDecimal fMinMemoryGib = minMemoryGib;
        final BigDecimal fMaxMemoryGib = maxMemoryGib;
        final BigDecimal fMinDiskGib = minDiskGib;
        final BigDecimal fMaxDiskGib = maxDiskGib;
        final Boolean fIsPublic = isPublic;
        final Boolean fIsRecoverable = isRecoverable;
        final OffsetDateTime fCreatedAtAfter = createdAtAfter;
        final OffsetDateTime fCreatedAtBefore = createdAtBefore;
        final OffsetDateTime fLastActivityAfter = lastActivityAfter;
        final OffsetDateTime fLastActivityBefore = lastActivityBefore;
        final io.daytona.api.client.model.SandboxListSortField fSort = sort;
        final io.daytona.api.client.model.SandboxListSortDirection fOrder = order;

        io.daytona.api.client.model.ListSandboxesResponse result = ExceptionMapper.callMain(() -> sandboxApi.listSandboxes(
                null,
                fCursor,
                fLimitVal,
                fId,
                fName,
                fLabelsJson,
                null,
                fStates,
                fSnapshots,
                fTargets,
                null,
                fMinCpu,
                fMaxCpu,
                fMinMemoryGib,
                fMaxMemoryGib,
                fMinDiskGib,
                fMaxDiskGib,
                fIsPublic,
                fIsRecoverable,
                fCreatedAtAfter,
                fCreatedAtBefore,
                fLastActivityAfter,
                fLastActivityBefore,
                fSort,
                fOrder
        ));

        List<Sandbox> items = new ArrayList<>();
        if (result != null && result.getItems() != null) {
            for (io.daytona.api.client.model.SandboxListItem item : result.getItems()) {
                items.add(new Sandbox(sandboxApi, config, item));
            }
        }
        String nextCursor = result != null ? result.getNextCursor() : null;
        return new PageResult(items, nextCursor);
    }

    /**
     * Internal page payload used by {@link SandboxIterator}.
     */
    static final class PageResult {
        final List<Sandbox> items;
        final String nextCursor;

        PageResult(List<Sandbox> items, String nextCursor) {
            this.items = items;
            this.nextCursor = nextCursor;
        }
    }

    /**
     * Cursor-based iterator that lazily pulls pages from the Daytona API.
     *
     * <p>Single-consumer, not thread-safe. Stops fetching as soon as the API
     * signals no further cursor.
     */
    private static final class SandboxIterator implements Iterator<Sandbox> {
        private final Daytona daytona;
        private final ListSandboxesQuery query;

        private List<Sandbox> page = null;
        private int pageIndex = 0;
        private String cursor = null;
        private boolean firstPageFetched = false;
        private boolean exhausted = false;

        SandboxIterator(Daytona daytona, ListSandboxesQuery query) {
            this.daytona = daytona;
            this.query = query;
        }

        @Override
        public boolean hasNext() {
            advanceIfNeeded();
            return page != null && pageIndex < page.size();
        }

        @Override
        public Sandbox next() {
            advanceIfNeeded();
            if (page == null || pageIndex >= page.size()) {
                throw new NoSuchElementException();
            }
            return page.get(pageIndex++);
        }

        private void advanceIfNeeded() {
            while ((page == null || pageIndex >= page.size()) && !exhausted) {
                if (firstPageFetched && cursor == null) {
                    exhausted = true;
                    return;
                }
                PageResult result = daytona.fetchSandboxPage(query, cursor);
                firstPageFetched = true;
                page = result.items;
                pageIndex = 0;
                cursor = result.nextCursor;
                if (cursor == null) {
                    exhausted = true;
                    if (page == null || page.isEmpty()) {
                        return;
                    }
                }
            }
        }
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
        if (params.getDomainAllowList() != null) body.setDomainAllowList(params.getDomainAllowList());
        if (params.getLinkedSandbox() != null) body.setLinkedSandbox(params.getLinkedSandbox());
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
