// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk;

import io.daytona.api.client.api.SandboxApi;
import io.daytona.api.client.model.BuildInfo;
import io.daytona.api.client.model.CreateSandboxSnapshot;
import io.daytona.api.client.model.ForkSandbox;
import io.daytona.api.client.model.SandboxLabels;
import io.daytona.api.client.model.SandboxListItem;
import io.daytona.api.client.model.SandboxVolume;
import io.daytona.api.client.model.ToolboxProxyUrl;
import io.daytona.api.client.model.UpdateSandboxNetworkSettings;
import io.daytona.sdk.exception.DaytonaException;

import java.math.BigDecimal;
import java.util.Collections;
import java.util.HashMap;
import java.util.List;
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

    // Fields shared by both io.daytona.api.client.model.Sandbox and SandboxListItem.
    private String id;
    private String name;
    private String organizationId;
    private String snapshot;
    private String user;
    private Map<String, String> labels;
    private Boolean isPublic;
    private String target;
    private int cpu;
    private int gpu;
    private int memory;
    private int disk;
    private String state;
    private String errorReason;
    private Boolean recoverable;
    private String backupState;
    private Integer autoStopInterval;
    private Integer autoArchiveInterval;
    private Integer autoDeleteInterval;
    private String createdAt;
    private String updatedAt;
    private String lastActivityAt;
    private String toolboxProxyUrl;

    // Fields only present on the full Sandbox DTO; not populated by Daytona.list() —
    // call refreshData() on each item to populate.
    private Map<String, String> env;
    private Boolean networkBlockAll;
    private String networkAllowList;
    private String domainAllowList;
    private List<SandboxVolume> volumes;
    private BuildInfo buildInfo;
    private String backupCreatedAt;

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
        populateFromDTO(data);
        this.toolboxApiClient = buildToolboxApiClient(sandboxApi, config);
        this.infoApi = new io.daytona.toolbox.client.api.InfoApi(toolboxApiClient);
        this.process = new Process(new io.daytona.toolbox.client.api.ProcessApi(toolboxApiClient), this);
        this.fs = new FileSystem(new io.daytona.toolbox.client.api.FileSystemApi(toolboxApiClient));
        this.git = new Git(new io.daytona.toolbox.client.api.GitApi(toolboxApiClient));
        this.computerUse = new ComputerUse(new io.daytona.toolbox.client.api.ComputerUseApi(toolboxApiClient));
        this.codeInterpreter = new CodeInterpreter(new io.daytona.toolbox.client.api.InterpreterApi(toolboxApiClient), this);
    }

    Sandbox(SandboxApi sandboxApi, DaytonaConfig config, SandboxListItem data) {
        this.sandboxApi = sandboxApi;
        this.config = config;
        this.apiKey = config.getApiKey();
        populateFromDTO(data);
        this.toolboxApiClient = buildToolboxApiClient(sandboxApi, config);
        this.infoApi = new io.daytona.toolbox.client.api.InfoApi(toolboxApiClient);
        this.process = new Process(new io.daytona.toolbox.client.api.ProcessApi(toolboxApiClient), this);
        this.fs = new FileSystem(new io.daytona.toolbox.client.api.FileSystemApi(toolboxApiClient));
        this.git = new Git(new io.daytona.toolbox.client.api.GitApi(toolboxApiClient));
        this.computerUse = new ComputerUse(new io.daytona.toolbox.client.api.ComputerUseApi(toolboxApiClient));
        this.codeInterpreter = new CodeInterpreter(new io.daytona.toolbox.client.api.InterpreterApi(toolboxApiClient), this);
    }

    /**
     * Builds the toolbox HTTP client, resolving the proxy URL if missing and attaching auth + SDK headers.
     * Requires {@code this.id} and {@code this.toolboxProxyUrl} to be populated.
     */
    private io.daytona.toolbox.client.ApiClient buildToolboxApiClient(SandboxApi sandboxApi, DaytonaConfig config) {
        String proxyBase = this.toolboxProxyUrl;
        if (proxyBase == null || proxyBase.isEmpty()) {
            ToolboxProxyUrl proxy = ExceptionMapper.callMain(() -> sandboxApi.getToolboxProxyUrl(this.id, null));
            proxyBase = proxy == null ? "" : proxy.getUrl();
        }

        String toolboxBase = trimTrailingSlash(proxyBase) + "/" + this.id;
        io.daytona.toolbox.client.ApiClient client = new io.daytona.toolbox.client.ApiClient();
        client.setBasePath(toolboxBase);
        String sdkVersion = Daytona.class.getPackage().getImplementationVersion();
        if (sdkVersion == null) sdkVersion = "dev";
        client.addDefaultHeader("Authorization", "Bearer " + config.getApiKey());
        client.addDefaultHeader("X-Daytona-Source", "sdk-java");
        client.addDefaultHeader("X-Daytona-SDK-Version", sdkVersion);
        client.setUserAgent("sdk-java/" + sdkVersion);
        return client;
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
            populateFromDTO(response);
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
            populateFromDTO(response);
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
            populateFromDTO(response);
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
            populateFromDTO(response);
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
            populateFromDTO(response);
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
     * Refreshes local Sandbox fields from latest API state. After refresh, all fields
     * — including those not returned by {@link Daytona#list} — are populated.
     *
     * @throws DaytonaException if refresh fails
     */
    public void refreshData() {
        io.daytona.api.client.model.Sandbox data = ExceptionMapper.callMain(() -> sandboxApi.getSandbox(id, null, null));
        if (data != null) {
            populateFromDTO(data);
        }
    }

    /**
     * Copies fields from the full {@link io.daytona.api.client.model.Sandbox} DTO onto this instance.
     *
     * <p>Populates every field, including those not returned by the list endpoint (env,
     * networkBlockAll, networkAllowList, volumes, buildInfo, backupCreatedAt).
     */
    private void populateFromDTO(io.daytona.api.client.model.Sandbox d) {
        if (d == null) {
            return;
        }
        populateCommonFields(
                d.getId(), d.getName(), d.getOrganizationId(), d.getSnapshot(), d.getUser(),
                d.getLabels(), d.getPublic(), d.getTarget(),
                d.getCpu(), d.getGpu(), d.getMemory(), d.getDisk(),
                d.getState() == null ? null : d.getState().getValue(),
                d.getErrorReason(), d.getRecoverable(),
                d.getBackupState() == null ? null : d.getBackupState().getValue(),
                d.getAutoStopInterval(), d.getAutoArchiveInterval(), d.getAutoDeleteInterval(),
                d.getCreatedAt(), d.getUpdatedAt(), d.getLastActivityAt(),
                d.getToolboxProxyUrl()
        );

        // Fields only present on the full Sandbox DTO.
        this.env = d.getEnv() == null ? new HashMap<String, String>() : new HashMap<String, String>(d.getEnv());
        this.networkBlockAll = d.getNetworkBlockAll();
        this.networkAllowList = d.getNetworkAllowList();
        this.domainAllowList = d.getDomainAllowList();
        this.volumes = d.getVolumes() == null ? null : Collections.unmodifiableList(d.getVolumes());
        this.buildInfo = d.getBuildInfo();
        this.backupCreatedAt = d.getBackupCreatedAt();
    }

    /**
     * Copies fields from a {@link SandboxListItem} DTO onto this instance.
     *
     * <p>The list endpoint omits env, networkBlockAll, networkAllowList, volumes, buildInfo, and
     * backupCreatedAt; those fields remain {@code null} until {@link #refreshData()} is called.
     */
    private void populateFromDTO(SandboxListItem d) {
        if (d == null) {
            return;
        }
        populateCommonFields(
                d.getId(), d.getName(), d.getOrganizationId(), d.getSnapshot(), d.getUser(),
                d.getLabels(), d.getPublic(), d.getTarget(),
                d.getCpu(), d.getGpu(), d.getMemory(), d.getDisk(),
                d.getState() == null ? null : d.getState().getValue(),
                d.getErrorReason(), d.getRecoverable(),
                d.getBackupState() == null ? null : d.getBackupState().getValue(),
                d.getAutoStopInterval(), d.getAutoArchiveInterval(), d.getAutoDeleteInterval(),
                d.getCreatedAt(), d.getUpdatedAt(), d.getLastActivityAt(),
                d.getToolboxProxyUrl()
        );
    }

    // Shared population logic for fields present on both Sandbox and SandboxListItem DTOs.
    // Takes already-extracted values (rather than the DTO itself) so the two type-safe overloads
    // above can each call it without referencing the other DTO's enum types.
    private void populateCommonFields(
            String id, String name, String organizationId, String snapshot, String user,
            Map<String, String> labels, Boolean isPublic, String target,
            BigDecimal cpu, BigDecimal gpu, BigDecimal memory, BigDecimal disk,
            String state, String errorReason, Boolean recoverable, String backupState,
            BigDecimal autoStopInterval, BigDecimal autoArchiveInterval, BigDecimal autoDeleteInterval,
            String createdAt, String updatedAt, String lastActivityAt,
            String toolboxProxyUrl) {
        this.id = asString(id);
        this.name = asString(name);
        this.organizationId = asString(organizationId);
        this.snapshot = snapshot;
        this.user = asString(user);
        this.labels = labels == null ? new HashMap<String, String>() : new HashMap<String, String>(labels);
        this.isPublic = isPublic;
        this.target = asString(target);
        this.cpu = cpu == null ? 0 : cpu.intValue();
        this.gpu = gpu == null ? 0 : gpu.intValue();
        this.memory = memory == null ? 0 : memory.intValue();
        this.disk = disk == null ? 0 : disk.intValue();
        this.state = state == null ? "" : state;
        this.errorReason = errorReason;
        this.recoverable = recoverable;
        this.backupState = backupState;
        this.autoStopInterval = autoStopInterval == null ? null : autoStopInterval.intValue();
        this.autoArchiveInterval = autoArchiveInterval == null ? null : autoArchiveInterval.intValue();
        this.autoDeleteInterval = autoDeleteInterval == null ? null : autoDeleteInterval.intValue();
        this.createdAt = createdAt;
        this.updatedAt = updatedAt;
        this.lastActivityAt = lastActivityAt;
        this.toolboxProxyUrl = asString(toolboxProxyUrl);
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
     * Pauses the Sandbox, freezing all running processes.
     * Uses default timeout of 60 seconds.
     *
     * @throws DaytonaException if the pause operation fails
     */
    public void pause() throws DaytonaException {
        pause(60);
    }

    /**
     * Pauses the Sandbox, freezing all running processes.
     * The Sandbox will enter a 'pausing' state and transition to 'paused' when complete.
     *
     * @param timeoutSeconds maximum time to wait in seconds (0 = no timeout)
     * @throws DaytonaException if timeout is negative or the operation fails/times out
     */
    public void pause(long timeoutSeconds) throws DaytonaException {
        if (timeoutSeconds < 0) {
            throw new DaytonaException("Timeout must be a non-negative number");
        }

        ExceptionMapper.callMain(() -> sandboxApi.pauseSandbox(id, null));
        refreshData();
        waitForPauseComplete(timeoutSeconds);
    }

    private void waitForPauseComplete(long timeoutSeconds) {
        long startedAt = System.currentTimeMillis();
        while ("pausing".equalsIgnoreCase(state)) {
            refreshData();
            if ("error".equalsIgnoreCase(state) || "build_failed".equalsIgnoreCase(state)) {
                throw new DaytonaException("Sandbox pause failed with state: " + state);
            }
            if (!"pausing".equalsIgnoreCase(state)) {
                return;
            }
            if (timeoutSeconds > 0 && (System.currentTimeMillis() - startedAt) > timeoutSeconds * 1000L) {
                throw new DaytonaException("Sandbox pause did not complete before timeout");
            }
            try {
                Thread.sleep(250);
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
                throw new DaytonaException("Interrupted while waiting for pause complete", e);
            }
        }
    }

    /** @return Sandbox ID. */
    public String getId() { return id; }
    /** @return Sandbox name. */
    public String getName() { return name; }
    /** @return organization ID that owns this Sandbox. */
    public String getOrganizationId() { return organizationId; }
    /** @return Daytona snapshot used to create this Sandbox, or {@code null} if none. */
    public String getSnapshot() { return snapshot; }
    /** @return OS user running in the Sandbox. */
    public String getUser() { return user; }
    /** @return custom labels attached to the Sandbox. */
    public Map<String, String> getLabels() { return labels; }
    /** @return whether the Sandbox HTTP preview is publicly accessible. */
    public Boolean getPublic() { return isPublic; }
    /** @return target region/environment where the Sandbox runs. */
    public String getTarget() { return target; }
    /** @return allocated CPU cores. */
    public int getCpu() { return cpu; }
    /** @return allocated GPU units. */
    public int getGpu() { return gpu; }
    /** @return allocated memory in GiB. */
    public int getMemory() { return memory; }
    /** @return allocated disk in GiB. */
    public int getDisk() { return disk; }
    /** @return current lifecycle state (e.g. "started", "stopped"). */
    public String getState() { return state; }
    /** @return error message if the Sandbox is in an error state, or {@code null}. */
    public String getErrorReason() { return errorReason; }
    /** @return whether the Sandbox error is recoverable, or {@code null} if unknown. */
    public Boolean getRecoverable() { return recoverable; }
    /** @return current state of the Sandbox backup as a string, or {@code null}. */
    public String getBackupState() { return backupState; }
    /** @return auto-stop interval in minutes (0 means disabled). */
    public Integer getAutoStopInterval() { return autoStopInterval; }
    /** @return auto-archive interval in minutes. */
    public Integer getAutoArchiveInterval() { return autoArchiveInterval; }
    /** @return auto-delete interval in minutes (negative means disabled). */
    public Integer getAutoDeleteInterval() { return autoDeleteInterval; }
    /** @return when the Sandbox was created, or {@code null}. */
    public String getCreatedAt() { return createdAt; }
    /** @return when the Sandbox was last updated, or {@code null}. */
    public String getUpdatedAt() { return updatedAt; }
    /** @return when the Sandbox last had activity, or {@code null}. */
    public String getLastActivityAt() { return lastActivityAt; }
    /** @return toolbox proxy URL. */
    public String getToolboxProxyUrl() { return toolboxProxyUrl; }

    /**
     * Returns Sandbox environment variables.
     *
     * <p>Not returned by {@link Daytona#list}; call {@link #refreshData()} on each item to populate.
     *
     * @return environment map, or {@code null} if not yet populated
     */
    public Map<String, String> getEnv() { return env; }
    /**
     * Returns whether all network access is blocked for this Sandbox.
     *
     * <p>Not returned by {@link Daytona#list}; call {@link #refreshData()} on each item to populate.
     *
     * @return block-all flag, or {@code null} if not yet populated
     */
    public Boolean getNetworkBlockAll() { return networkBlockAll; }
    /**
     * Returns the comma-separated CIDR allow list, if any.
     *
     * <p>Not returned by {@link Daytona#list}; call {@link #refreshData()} on each item to populate.
     *
     * @return allow list, or {@code null}
     */
    public String getNetworkAllowList() { return networkAllowList; }
    /**
     * Returns the comma-separated list of allowed domains, if any.
     *
     * <p>Not returned by {@link Daytona#list}; call {@link #refreshData()} on each item to populate.
     *
     * @return allowed domains, or {@code null}
     */
    public String getDomainAllowList() { return domainAllowList; }
    /**
     * Returns volumes attached to the Sandbox.
     *
     * <p>Not returned by {@link Daytona#list}; call {@link #refreshData()} on each item to populate.
     *
     * @return immutable list of attached volumes, or {@code null} if not yet populated
     */
    public List<SandboxVolume> getVolumes() { return volumes; }
    /**
     * Returns build information if the Sandbox was created from a dynamic build.
     *
     * <p>Not returned by {@link Daytona#list}; call {@link #refreshData()} on each item to populate.
     *
     * @return build info, or {@code null}
     */
    public BuildInfo getBuildInfo() { return buildInfo; }
    /**
     * Returns the creation timestamp of the last backup.
     *
     * <p>Not returned by {@link Daytona#list}; call {@link #refreshData()} on each item to populate.
     *
     * @return backup timestamp, or {@code null}
     */
    public String getBackupCreatedAt() { return backupCreatedAt; }

    /** @return process operations facade. */
    public Process getProcess() { return process; }
    /** @return file-system operations facade. */
    public FileSystem getFs() { return fs; }
    /** @return Git operations facade. */
    public Git getGit() { return git; }
    io.daytona.toolbox.client.ApiClient getToolboxApiClient() { return toolboxApiClient; }
    String getApiKey() { return apiKey; }
}
