// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk;

import com.fasterxml.jackson.databind.JsonNode;
import io.daytona.api.client.api.SandboxApi;
import io.daytona.api.client.model.SandboxLabels;
import io.daytona.api.client.model.ToolboxProxyUrl;
import io.daytona.sdk.exception.DaytonaException;
import io.daytona.sdk.exception.DaytonaNotFoundException;
import io.daytona.sdk.internal.EventSubscriber;

import java.math.BigDecimal;
import java.util.Arrays;
import java.util.Collections;
import java.util.HashMap;
import java.util.HashSet;
import java.util.Map;
import java.util.Set;
import java.util.concurrent.CopyOnWriteArrayList;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.Executors;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.ScheduledFuture;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicReference;

/**
 * Represents a Daytona Sandbox instance.
 *
 * <p>Exposes lifecycle controls and operation facades for process execution, file-system access,
 * and Git. State changes are detected instantly via WebSocket events with periodic polling as a
 * safety net.
 */
public class Sandbox {

    private static final long POLL_SAFETY_INTERVAL_SECONDS = 1;

    private static final Set<String> STARTED_STATES = Collections.singleton("started");
    private static final Set<String> STOPPED_STATES = new HashSet<>(Arrays.asList("stopped", "destroyed"));
    private static final Set<String> DESTROYED_STATES = Collections.singleton("destroyed");
    private static final Set<String> ERROR_STATES = new HashSet<>(Arrays.asList("error", "build_failed"));
    private static final Set<String> RESIZE_TARGET_STATES;
    static {
        Set<String> states = new HashSet<>();
        for (io.daytona.api.client.model.SandboxState s : io.daytona.api.client.model.SandboxState.values()) {
            states.add(s.getValue());
        }
        states.remove("resizing");
        states.removeAll(ERROR_STATES);
        RESIZE_TARGET_STATES = Collections.unmodifiableSet(states);
    }

    private static final ScheduledExecutorService POLL_SCHEDULER = Executors.newScheduledThreadPool(2, r -> {
        Thread t = new Thread(r, "daytona-sandbox-poller");
        t.setDaemon(true);
        return t;
    });

    private final SandboxApi sandboxApi;
    private final io.daytona.toolbox.client.ApiClient toolboxApiClient;
    private final io.daytona.toolbox.client.api.InfoApi infoApi;
    private final String apiKey;
    private final EventSubscriber eventSubscriber;
    private final CopyOnWriteArrayList<StateWaiter> stateWaiters = new CopyOnWriteArrayList<>();

    private String id;
    private String name;
    private volatile String state;
    private String target;
    private String user;
    private String toolboxProxyUrl;
    private int cpu;
    private int gpu;
    private int memory;
    private int disk;
    private Map<String, String> env;
    private Map<String, String> labels;
    private Integer autoStopInterval;
    private Integer autoArchiveInterval;
    private Integer autoDeleteInterval;

    /** Process execution interface for this Sandbox. */
    public final Process process;
    /** File-system operations interface for this Sandbox. */
    public final FileSystem fs;
    /** Git operations interface for this Sandbox. */
    public final Git git;
    /** Computer use (desktop automation) interface for this Sandbox. */
    public final ComputerUse computerUse;
    /** Stateful code interpreter for this Sandbox (Python). */
    public final CodeInterpreter codeInterpreter;

    Sandbox(SandboxApi sandboxApi, DaytonaConfig config, io.daytona.api.client.model.Sandbox data,
            EventSubscriber eventSubscriber) {
        this.sandboxApi = sandboxApi;
        this.apiKey = config.getApiKey();
        this.eventSubscriber = eventSubscriber;
        updateFromModel(data);

        String proxyBase = this.toolboxProxyUrl;
        if (proxyBase == null || proxyBase.isEmpty()) {
            ToolboxProxyUrl proxy = ExceptionMapper.callMain(() -> sandboxApi.getToolboxProxyUrl(this.id, null));
            proxyBase = proxy == null ? "" : proxy.getUrl();
        }

        String toolboxBase = trimTrailingSlash(proxyBase) + "/" + this.id;
        this.toolboxApiClient = new io.daytona.toolbox.client.ApiClient();
        this.toolboxApiClient.setBasePath(toolboxBase);
        String sdkVersion = Daytona.class.getPackage().getImplementationVersion();
        if (sdkVersion == null) sdkVersion = "dev";
        this.toolboxApiClient.addDefaultHeader("Authorization", "Bearer " + config.getApiKey());
        this.toolboxApiClient.addDefaultHeader("X-Daytona-Source", "sdk-java");
        this.toolboxApiClient.addDefaultHeader("X-Daytona-SDK-Version", sdkVersion);
        this.toolboxApiClient.setUserAgent("sdk-java/" + sdkVersion);

        this.infoApi = new io.daytona.toolbox.client.api.InfoApi(toolboxApiClient);
        this.process = new Process(new io.daytona.toolbox.client.api.ProcessApi(toolboxApiClient), this);
        this.fs = new FileSystem(new io.daytona.toolbox.client.api.FileSystemApi(toolboxApiClient));
        this.git = new Git(new io.daytona.toolbox.client.api.GitApi(toolboxApiClient));
        this.computerUse = new ComputerUse(new io.daytona.toolbox.client.api.ComputerUseApi(toolboxApiClient));
        this.codeInterpreter = new CodeInterpreter(new io.daytona.toolbox.client.api.InterpreterApi(toolboxApiClient), this);

        subscribeToEvents();
    }

    /**
     * Creates an LSP server instance for the specified language and project.
     *
     * @param languageId language server to start (e.g. "typescript", "python", "go")
     * @param pathToProject absolute path to the project root inside the sandbox
     * @return a new {@link LspServer} configured for the given language
     */
    public LspServer createLspServer(String languageId, String pathToProject) {
        return new LspServer(new io.daytona.toolbox.client.api.LspApi(toolboxApiClient));
    }

    io.daytona.sdk.codetoolbox.CodeToolbox getCodeToolbox() {
        String lang = "python";
        if (labels != null && labels.containsKey("code-toolbox-language")) {
            lang = labels.get("code-toolbox-language");
        }
        if ("typescript".equalsIgnoreCase(lang)) {
            return new io.daytona.sdk.codetoolbox.TypeScriptCodeToolbox();
        } else if ("javascript".equalsIgnoreCase(lang)) {
            return new io.daytona.sdk.codetoolbox.JavaScriptCodeToolbox();
        }
        return new io.daytona.sdk.codetoolbox.PythonCodeToolbox();
    }

    /**
     * Starts this Sandbox with default timeout.
     *
     * @throws DaytonaException if the Sandbox fails to start
     */
    public void start() {
        start(60);
    }

    /**
     * Starts this Sandbox and waits for readiness.
     *
     * @param timeoutSeconds maximum seconds to wait; {@code 0} disables timeout
     * @throws DaytonaException if start fails or times out
     */
    public void start(long timeoutSeconds) {
        io.daytona.api.client.model.Sandbox response = ExceptionMapper.callMain(() -> sandboxApi.startSandbox(id, null));
        if (response != null) {
            updateFromModel(response);
        }
        waitUntilStarted(timeoutSeconds);
    }

    /**
     * Stops this Sandbox with default timeout.
     *
     * @throws DaytonaException if the Sandbox fails to stop
     */
    public void stop() {
        stop(60);
    }

    /**
     * Stops this Sandbox and waits until fully stopped.
     *
     * @param timeoutSeconds maximum seconds to wait; {@code 0} disables timeout
     * @throws DaytonaException if stop fails or times out
     */
    public void stop(long timeoutSeconds) {
        ExceptionMapper.callMain(() -> sandboxApi.stopSandbox(id, null, null));
        refreshData();
        waitUntilStopped(timeoutSeconds);
    }

    /**
     * Waits until Sandbox reaches {@code stopped} (or {@code destroyed}) state.
     *
     * @param timeoutSeconds maximum seconds to wait; {@code 0} disables timeout
     * @throws DaytonaException if timeout is invalid, state becomes error, or timeout expires
     */
    public void waitUntilStopped(long timeoutSeconds) {
        if (timeoutSeconds < 0) {
            throw new DaytonaException("Timeout must be non-negative");
        }
        if (STOPPED_STATES.contains(state)) {
            return;
        }
        waitForState(STOPPED_STATES, ERROR_STATES, timeoutSeconds, false);
    }

    /**
     * Deletes this Sandbox with default timeout behavior.
     *
     * @throws DaytonaException if deletion fails
     */
    public void delete() {
        delete(60);
    }

    /**
     * Deletes this Sandbox and waits for it to reach the {@code destroyed} state.
     *
     * @param timeoutSeconds maximum seconds to wait; {@code 0} disables timeout
     * @throws DaytonaException if deletion fails or times out
     */
    public void delete(long timeoutSeconds) {
        long startTime = System.currentTimeMillis();
        ExceptionMapper.callMain(() -> sandboxApi.deleteSandbox(id, null));

        refreshDataSafe();
        if ("destroyed".equalsIgnoreCase(state)) {
            return;
        }

        long elapsed = (System.currentTimeMillis() - startTime) / 1000;
        long remaining = timeoutSeconds > 0 ? Math.max(1, timeoutSeconds - elapsed) : 0;
        waitForState(DESTROYED_STATES, ERROR_STATES, remaining, true);
    }

    /**
     * Replaces Sandbox labels.
     *
     * @param labels label map to apply
     * @return updated labels
     * @throws DaytonaException if label update fails
     */
    public Map<String, String> setLabels(Map<String, String> labels) {
        ExceptionMapper.callMain(() -> {
            okhttp3.Call call = sandboxApi.replaceLabelsCall(id, new SandboxLabels().labels(labels), null, null);
            sandboxApi.getApiClient().execute(call, null);
            return null;
        });
        refreshData();
        return this.labels;
    }

    /**
     * Sets Sandbox auto-stop interval.
     *
     * @param minutes idle minutes before automatic stop
     * @throws DaytonaException if the update fails
     */
    public void setAutostopInterval(int minutes) {
        io.daytona.api.client.model.Sandbox response = ExceptionMapper.callMain(() -> sandboxApi.setAutostopInterval(id, BigDecimal.valueOf(minutes), null));
        if (response != null) {
            updateFromModel(response);
        }
    }

    /**
     * Sets Sandbox auto-archive interval.
     *
     * @param minutes minutes in stopped state before automatic archive
     * @throws DaytonaException if the update fails
     */
    public void setAutoArchiveInterval(int minutes) {
        io.daytona.api.client.model.Sandbox response = ExceptionMapper.callMain(() -> sandboxApi.setAutoArchiveInterval(id, BigDecimal.valueOf(minutes), null));
        if (response != null) {
            updateFromModel(response);
        }
    }

    /**
     * Sets Sandbox auto-delete interval.
     *
     * @param minutes minutes before automatic deletion after stop
     * @throws DaytonaException if the update fails
     */
    public void setAutoDeleteInterval(int minutes) {
        io.daytona.api.client.model.Sandbox response = ExceptionMapper.callMain(() -> sandboxApi.setAutoDeleteInterval(id, BigDecimal.valueOf(minutes), null));
        if (response != null) {
            updateFromModel(response);
        }
    }

    /**
     * Returns home directory path for Sandbox user.
     *
     * @return absolute home directory path
     * @throws DaytonaException if the request fails
     */
    public String getUserHomeDir() {
        io.daytona.toolbox.client.model.UserHomeDirResponse value = ExceptionMapper.callToolbox(() -> infoApi.getUserHomeDir());
        return value == null ? "" : asString(value.getDir());
    }

    /**
     * Returns current working directory path.
     *
     * @return absolute working directory path
     * @throws DaytonaException if the request fails
     */
    public String getWorkDir() {
        io.daytona.toolbox.client.model.WorkDirResponse value = ExceptionMapper.callToolbox(() -> infoApi.getWorkDir());
        return value == null ? "" : asString(value.getDir());
    }

    /**
     * Waits until Sandbox reaches {@code started} state.
     *
     * @param timeoutSeconds maximum seconds to wait; {@code 0} disables timeout
     * @throws DaytonaException if timeout is invalid, state becomes failure, or timeout expires
     */
    public void waitUntilStarted(long timeoutSeconds) {
        if (timeoutSeconds < 0) {
            throw new DaytonaException("Timeout must be non-negative");
        }
        if ("started".equalsIgnoreCase(state)) {
            return;
        }
        waitForState(STARTED_STATES, ERROR_STATES, timeoutSeconds, false);
    }

    /**
     * Waits for a resize operation to complete.
     *
     * @param timeoutSeconds maximum seconds to wait; {@code 0} disables timeout
     * @throws DaytonaException if resize times out or fails
     */
    public void waitForResizeComplete(long timeoutSeconds) {
        if (timeoutSeconds < 0) {
            throw new DaytonaException("Timeout must be non-negative");
        }
        if (!"resizing".equalsIgnoreCase(state)) {
            return;
        }
        waitForState(RESIZE_TARGET_STATES, ERROR_STATES, timeoutSeconds, false);
    }

    /**
     * Refreshes local Sandbox fields from latest API state.
     *
     * @throws DaytonaException if refresh fails
     */
    public void refreshData() {
        io.daytona.api.client.model.Sandbox data = ExceptionMapper.callMain(() -> sandboxApi.getSandbox(id, null, null));
        if (data != null) {
            updateFromModel(data);
        }
    }

    private void applyState(String newState) {
        if (newState == null || newState.equals(this.state)) {
            return;
        }
        this.state = newState;
        for (StateWaiter waiter : stateWaiters) {
            waiter.onStateChanged(newState);
        }
    }

    private void subscribeToEvents() {
        if (eventSubscriber == null) {
            return;
        }
        eventSubscriber.ensureConnected();
        eventSubscriber.subscribe(id, (eventName, data) -> {
            if (data == null || !data.isObject()) {
                return;
            }
            if ("sandbox.created".equals(eventName)) {
                JsonNode sandboxNode = data.has("sandbox") ? data.get("sandbox") : data;
                if (sandboxNode != null) {
                    updateFromJsonEvent(sandboxNode);
                }
            } else {
                JsonNode sandboxNode = data.has("sandbox") ? data.get("sandbox") : data;
                if (sandboxNode != null && sandboxNode.has("state")) {
                    JsonNode stateNode = sandboxNode.get("state");
                    if (stateNode != null && stateNode.isTextual()) {
                        applyState(stateNode.asText());
                    }
                }
            }
        }, Arrays.asList("sandbox.state.updated", "sandbox.created"));
    }

    private void waitForState(Set<String> targetStates, Set<String> errorStates,
                              long timeoutSeconds, boolean safeRefresh) {
        StateWaiter waiter = new StateWaiter(targetStates, errorStates);
        stateWaiters.add(waiter);

        try {
            String current = state;
            if (current != null) {
                waiter.onStateChanged(current);
            }
            if (waiter.isResolved()) {
                waiter.throwIfError();
                return;
            }

            ScheduledFuture<?> pollFuture = POLL_SCHEDULER.scheduleAtFixedRate(() -> {
                if (waiter.isResolved()) {
                    return;
                }
                try {
                    if (safeRefresh) {
                        refreshDataSafe();
                    } else {
                        refreshData();
                    }
                } catch (Exception e) {
                    return;
                }
            }, POLL_SAFETY_INTERVAL_SECONDS, POLL_SAFETY_INTERVAL_SECONDS, TimeUnit.SECONDS);

            try {
                boolean completed;
                if (timeoutSeconds > 0) {
                    completed = waiter.latch.await(timeoutSeconds, TimeUnit.SECONDS);
                } else {
                    waiter.latch.await();
                    completed = true;
                }

                if (!completed) {
                    throw new DaytonaException("Sandbox " + id + " did not reach target state within " + timeoutSeconds + " seconds");
                }
                waiter.throwIfError();
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
                throw new DaytonaException("Interrupted while waiting for sandbox state change", e);
            } finally {
                pollFuture.cancel(false);
            }
        } finally {
            stateWaiters.remove(waiter);
        }
    }

    private void refreshDataSafe() {
        try {
            refreshData();
        } catch (DaytonaNotFoundException e) {
            applyState("destroyed");
        } catch (Exception e) {
            if (e.getMessage() != null && e.getMessage().contains("404")) {
                applyState("destroyed");
            }
        }
    }

    private void updateFromModel(io.daytona.api.client.model.Sandbox data) {
        if (data == null) {
            return;
        }
        this.id = asString(data.getId());
        this.name = asString(data.getName());
        this.target = asString(data.getTarget());
        this.user = asString(data.getUser());
        String newProxyUrl = asString(data.getToolboxProxyUrl());
        if (!newProxyUrl.isEmpty() && !newProxyUrl.equals(this.toolboxProxyUrl) && this.toolboxApiClient != null) {
            this.toolboxApiClient.setBasePath(trimTrailingSlash(newProxyUrl) + "/" + this.id);
        }
        this.toolboxProxyUrl = newProxyUrl;
        this.cpu = data.getCpu() == null ? 0 : data.getCpu().intValue();
        this.gpu = data.getGpu() == null ? 0 : data.getGpu().intValue();
        this.memory = data.getMemory() == null ? 0 : data.getMemory().intValue();
        this.disk = data.getDisk() == null ? 0 : data.getDisk().intValue();
        this.env = data.getEnv() == null ? new HashMap<String, String>() : new HashMap<String, String>(data.getEnv());
        this.labels = data.getLabels() == null ? new HashMap<String, String>() : new HashMap<String, String>(data.getLabels());
        this.autoStopInterval = data.getAutoStopInterval() == null ? null : data.getAutoStopInterval().intValue();
        this.autoArchiveInterval = data.getAutoArchiveInterval() == null ? null : data.getAutoArchiveInterval().intValue();
        this.autoDeleteInterval = data.getAutoDeleteInterval() == null ? null : data.getAutoDeleteInterval().intValue();
        applyState(data.getState() == null ? "" : data.getState().getValue());
    }

    private void updateFromJsonEvent(JsonNode node) {
        try {
            io.daytona.api.client.model.Sandbox model = new com.fasterxml.jackson.databind.ObjectMapper()
                    .treeToValue(node, io.daytona.api.client.model.Sandbox.class);
            if (model != null) {
                updateFromModel(model);
            }
        } catch (Exception e) {
            JsonNode stateNode = node.path("state");
            if (stateNode.isTextual()) {
                applyState(stateNode.asText());
            }
        }
    }

    private static class StateWaiter {
        final Set<String> targetStates;
        final Set<String> errorStates;
        final AtomicReference<String> resolvedState = new AtomicReference<>();
        final CountDownLatch latch = new CountDownLatch(1);

        StateWaiter(Set<String> targetStates, Set<String> errorStates) {
            this.targetStates = targetStates;
            this.errorStates = errorStates;
        }

        void onStateChanged(String newState) {
            if (targetStates.contains(newState) || errorStates.contains(newState)) {
                if (resolvedState.compareAndSet(null, newState)) {
                    latch.countDown();
                }
            }
        }

        boolean isResolved() {
            return resolvedState.get() != null;
        }

        void throwIfError() {
            String resolved = resolvedState.get();
            if (resolved != null && errorStates.contains(resolved)) {
                throw new DaytonaException("Sandbox entered error state: " + resolved);
            }
        }
    }

    private String asString(Object value) {
        return value == null ? "" : String.valueOf(value);
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

    /** @return Sandbox ID */
    public String getId() { return id; }
    /** @return Sandbox name */
    public String getName() { return name; }
    /** @return lifecycle state */
    public String getState() { return state; }
    /** @return target identifier */
    public String getTarget() { return target; }
    /** @return OS user */
    public String getUser() { return user; }
    /** @return proxy URL */
    public String getToolboxProxyUrl() { return toolboxProxyUrl; }
    /** @return CPU cores */
    public int getCpu() { return cpu; }
    /** @return GPU units */
    public int getGpu() { return gpu; }
    /** @return memory in GiB */
    public int getMemory() { return memory; }
    /** @return disk in GiB */
    public int getDisk() { return disk; }
    /** @return environment map */
    public Map<String, String> getEnv() { return env; }
    /** @return labels map */
    public Map<String, String> getLabels() { return labels; }
    /** @return auto-stop interval */
    public Integer getAutoStopInterval() { return autoStopInterval; }
    /** @return auto-archive interval */
    public Integer getAutoArchiveInterval() { return autoArchiveInterval; }
    /** @return auto-delete interval */
    public Integer getAutoDeleteInterval() { return autoDeleteInterval; }
    /** @return process interface */
    public Process getProcess() { return process; }
    /** @return file-system interface */
    public FileSystem getFs() { return fs; }
    /** @return Git interface */
    public Git getGit() { return git; }
    io.daytona.toolbox.client.ApiClient getToolboxApiClient() { return toolboxApiClient; }
    String getApiKey() { return apiKey; }
}
