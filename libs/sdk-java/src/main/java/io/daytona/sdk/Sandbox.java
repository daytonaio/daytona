// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk;

import io.daytona.api.client.api.SandboxApi;
import io.daytona.api.client.model.CreateSandboxSnapshot;
import io.daytona.api.client.model.ForkSandbox;
import io.daytona.api.client.model.SandboxLabels;
import io.daytona.api.client.model.ToolboxProxyUrl;
import io.daytona.api.client.model.UpdateSandboxNetworkSettings;
import io.daytona.sdk.exception.DaytonaException;

import java.math.BigDecimal;
import java.util.HashMap;
import java.util.Map;

/**
 * Represents a Daytona Sandbox instance.
 *
 * <p>Exposes lifecycle controls and operation facades for process execution, file-system access,
 * and Git.
 */
public class Sandbox {
    private final SandboxApi sandboxApi;
    private final DaytonaConfig config;
    private final io.daytona.toolbox.client.ApiClient toolboxApiClient;
    private final io.daytona.toolbox.client.api.InfoApi infoApi;
    private final String apiKey;

    private String id;
    private String name;
    private String state;
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
    private Boolean networkBlockAll;
    private String networkAllowList;

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

    Sandbox(SandboxApi sandboxApi, DaytonaConfig config, io.daytona.api.client.model.Sandbox data) {
        this.sandboxApi = sandboxApi;
        this.config = config;
        this.apiKey = config.getApiKey();
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

    String getLanguage() {
        String lang = "python";
        if (labels != null && labels.containsKey(Daytona.CODE_TOOLBOX_LANGUAGE_LABEL)) {
            lang = labels.get(Daytona.CODE_TOOLBOX_LANGUAGE_LABEL);
        }
        return lang;
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
        long startedAt = System.currentTimeMillis();
        while (!"stopped".equalsIgnoreCase(state) && !"destroyed".equalsIgnoreCase(state)) {
            refreshData();
            if ("stopped".equalsIgnoreCase(state) || "destroyed".equalsIgnoreCase(state)) {
                return;
            }
            if ("error".equalsIgnoreCase(state)) {
                throw new DaytonaException("Sandbox entered error state while stopping");
            }
            if (timeoutSeconds > 0 && (System.currentTimeMillis() - startedAt) > timeoutSeconds * 1000L) {
                throw new DaytonaException("Sandbox failed to stop before timeout");
            }
            try {
                Thread.sleep(250);
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
                throw new DaytonaException("Interrupted while waiting for sandbox stop", e);
            }
        }
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
     * Deletes this Sandbox.
     *
     * @param timeoutSeconds reserved timeout parameter for parity with other SDKs
     * @throws DaytonaException if deletion fails
     */
    public void delete(long timeoutSeconds) {
        ExceptionMapper.callMain(() -> sandboxApi.deleteSandbox(id, null));
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
     * Updates outbound network policy on the runner (block all, restore access, or CIDR allow list).
     *
     * @param settings request body; at least one of networkBlockAll or networkAllowList must be set
     * @throws DaytonaException if the update fails
     */
    public void updateNetworkSettings(UpdateSandboxNetworkSettings settings) {
        io.daytona.api.client.model.Sandbox response = ExceptionMapper.callMain(() -> sandboxApi.updateNetworkSettings(id, settings, null));
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

        long startedAt = System.currentTimeMillis();
        while (!"started".equalsIgnoreCase(state)) {
            refreshData();

            if ("error".equalsIgnoreCase(state) || "build_failed".equalsIgnoreCase(state)) {
                throw new DaytonaException("Sandbox entered failure state: " + state);
            }

            if (timeoutSeconds > 0 && (System.currentTimeMillis() - startedAt) > timeoutSeconds * 1000L) {
                throw new DaytonaException("Sandbox failed to become started before timeout");
            }

            try {
                Thread.sleep(250);
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
                throw new DaytonaException("Interrupted while waiting for sandbox start", e);
            }
        }
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

    private void updateFromModel(io.daytona.api.client.model.Sandbox data) {
        if (data == null) {
            return;
        }
        this.id = asString(data.getId());
        this.name = asString(data.getName());
        this.state = data.getState() == null ? "" : data.getState().getValue();
        this.target = asString(data.getTarget());
        this.user = asString(data.getUser());
        this.toolboxProxyUrl = asString(data.getToolboxProxyUrl());
        this.cpu = data.getCpu() == null ? 0 : data.getCpu().intValue();
        this.gpu = data.getGpu() == null ? 0 : data.getGpu().intValue();
        this.memory = data.getMemory() == null ? 0 : data.getMemory().intValue();
        this.disk = data.getDisk() == null ? 0 : data.getDisk().intValue();
        this.env = data.getEnv() == null ? new HashMap<String, String>() : new HashMap<String, String>(data.getEnv());
        this.labels = data.getLabels() == null ? new HashMap<String, String>() : new HashMap<String, String>(data.getLabels());
        this.autoStopInterval = data.getAutoStopInterval() == null ? null : data.getAutoStopInterval().intValue();
        this.autoArchiveInterval = data.getAutoArchiveInterval() == null ? null : data.getAutoArchiveInterval().intValue();
        this.autoDeleteInterval = data.getAutoDeleteInterval() == null ? null : data.getAutoDeleteInterval().intValue();
        this.networkBlockAll = data.getNetworkBlockAll();
        this.networkAllowList = data.getNetworkAllowList();
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

    /**
     * Forks this Sandbox, creating a new Sandbox with an identical filesystem.
     * Uses default timeout of 60 seconds.
     *
     * @return the forked {@link Sandbox} in started state
     * @throws DaytonaException if the fork operation fails or times out
     */
    public Sandbox experimentalFork() {
        return experimentalFork(null, 60);
    }

    /**
     * Forks this Sandbox, creating a new Sandbox with an identical filesystem.
     * The forked Sandbox is a copy-on-write clone of the original.
     *
     * @param name optional name for the forked Sandbox; {@code null} for auto-generated
     * @param timeoutSeconds maximum seconds to wait for the forked Sandbox to start; {@code 0} disables timeout
     * @return the forked {@link Sandbox} in started state
     * @throws DaytonaException if the fork operation fails or times out
     */
    public Sandbox experimentalFork(String name, long timeoutSeconds) {
        ForkSandbox forkReq = new ForkSandbox();
        if (name != null) {
            forkReq.setName(name);
        }
        io.daytona.api.client.model.Sandbox response = ExceptionMapper.callMain(
            () -> sandboxApi.forkSandbox(id, forkReq, null)
        );
        Sandbox forked = new Sandbox(sandboxApi, config, response);
        forked.waitUntilStarted(timeoutSeconds);
        return forked;
    }

    /**
     * Creates a snapshot from the current state of this Sandbox.
     * Uses default timeout of 60 seconds.
     *
     * @param name name for the new snapshot
     * @throws DaytonaException if the snapshot operation fails
     */
    public void experimentalCreateSnapshot(String name) {
        experimentalCreateSnapshot(name, 60);
    }

    /**
     * Creates a snapshot from the current state of this Sandbox.
     * The Sandbox will temporarily enter a 'snapshotting' state and return to its previous state when complete.
     *
     * @param name name for the new snapshot
     * @param timeoutSeconds reserved timeout parameter for parity with other SDKs
     * @throws DaytonaException if the snapshot operation fails
     */
    public void experimentalCreateSnapshot(String name, long timeoutSeconds) {
        CreateSandboxSnapshot req = new CreateSandboxSnapshot();
        req.setName(name);
        ExceptionMapper.callMain(() -> sandboxApi.createSandboxSnapshot(id, req, null));
        refreshData();
        waitForSnapshotComplete(timeoutSeconds);
    }

    private void waitForSnapshotComplete(long timeoutSeconds) {
        long startedAt = System.currentTimeMillis();
        while ("snapshotting".equalsIgnoreCase(state)) {
            refreshData();
            if ("error".equalsIgnoreCase(state) || "build_failed".equalsIgnoreCase(state)) {
                throw new DaytonaException("Sandbox snapshot failed with state: " + state);
            }
            if (!"snapshotting".equalsIgnoreCase(state)) {
                return;
            }
            if (timeoutSeconds > 0 && (System.currentTimeMillis() - startedAt) > timeoutSeconds * 1000L) {
                throw new DaytonaException("Sandbox snapshot did not complete before timeout");
            }
            try {
                Thread.sleep(250);
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
                throw new DaytonaException("Interrupted while waiting for snapshot complete", e);
            }
        }
    }

    /**
     * Returns Sandbox ID.
     *
     * @return Sandbox ID
     */
    public String getId() { return id; }
    /**
     * Returns Sandbox name.
     *
     * @return Sandbox name
     */
    public String getName() { return name; }
    /**
     * Returns Sandbox state.
     *
     * @return lifecycle state
     */
    public String getState() { return state; }
    /**
     * Returns target region.
     *
     * @return target identifier
     */
    public String getTarget() { return target; }
    /**
     * Returns Sandbox OS user.
     *
     * @return OS user
     */
    public String getUser() { return user; }
    /**
     * Returns toolbox proxy URL.
     *
     * @return proxy URL
     */
    public String getToolboxProxyUrl() { return toolboxProxyUrl; }
    /**
     * Returns allocated CPU cores.
     *
     * @return CPU cores
     */
    public int getCpu() { return cpu; }
    /**
     * Returns allocated GPU units.
     *
     * @return GPU units
     */
    public int getGpu() { return gpu; }
    /**
     * Returns allocated memory in GiB.
     *
     * @return memory in GiB
     */
    public int getMemory() { return memory; }
    /**
     * Returns allocated disk in GiB.
     *
     * @return disk in GiB
     */
    public int getDisk() { return disk; }
    /**
     * Returns Sandbox environment variables.
     *
     * @return environment map
     */
    public Map<String, String> getEnv() { return env; }
    /**
     * Returns Sandbox labels.
     *
     * @return labels map
     */
    public Map<String, String> getLabels() { return labels; }
    /**
     * Returns auto-stop interval in minutes.
     *
     * @return auto-stop interval
     */
    public Integer getAutoStopInterval() { return autoStopInterval; }
    /**
     * Returns auto-archive interval in minutes.
     *
     * @return auto-archive interval
     */
    public Integer getAutoArchiveInterval() { return autoArchiveInterval; }
    /**
     * Returns auto-delete interval in minutes.
     *
     * @return auto-delete interval
     */
    public Integer getAutoDeleteInterval() { return autoDeleteInterval; }
    /**
     * Returns whether all network access is blocked for this Sandbox.
     *
     * @return block-all flag, or null if unknown
     */
    public Boolean getNetworkBlockAll() { return networkBlockAll; }
    /**
     * Returns the comma-separated CIDR allow list, if any.
     *
     * @return allow list or null
     */
    public String getNetworkAllowList() { return networkAllowList; }

    /**
     * Returns process operations facade.
     *
     * @return process interface
     */
    public Process getProcess() { return process; }
    /**
     * Returns file-system operations facade.
     *
     * @return file-system interface
     */
    public FileSystem getFs() { return fs; }
    /**
     * Returns Git operations facade.
     *
     * @return Git interface
     */
    public Git getGit() { return git; }
    io.daytona.toolbox.client.ApiClient getToolboxApiClient() { return toolboxApiClient; }
    String getApiKey() { return apiKey; }
}
