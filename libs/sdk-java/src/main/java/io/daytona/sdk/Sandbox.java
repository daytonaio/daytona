// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk;

import io.daytona.api.client.api.SandboxApi;
import io.daytona.api.client.model.SandboxLabels;
import io.daytona.api.client.model.ToolboxProxyUrl;
import io.daytona.sdk.exception.DaytonaException;

import java.math.BigDecimal;
import java.util.HashMap;
import java.util.Map;

public class Sandbox {
    private final SandboxApi sandboxApi;
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

    public final SandboxProcess process;
    public final SandboxFileSystem fs;
    public final SandboxGit git;

    Sandbox(SandboxApi sandboxApi, DaytonaConfig config, io.daytona.api.client.model.Sandbox data) {
        this.sandboxApi = sandboxApi;
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
        this.toolboxApiClient.addDefaultHeader("Authorization", "Bearer " + config.getApiKey());
        this.toolboxApiClient.addDefaultHeader("X-Daytona-Source", "sdk-java");
        this.toolboxApiClient.setUserAgent("sdk-java/0.1.0");

        this.infoApi = new io.daytona.toolbox.client.api.InfoApi(toolboxApiClient);
        this.process = new SandboxProcess(new io.daytona.toolbox.client.api.ProcessApi(toolboxApiClient), this);
        this.fs = new SandboxFileSystem(new io.daytona.toolbox.client.api.FileSystemApi(toolboxApiClient));
        this.git = new SandboxGit(new io.daytona.toolbox.client.api.GitApi(toolboxApiClient));
    }

    public void start() {
        start(60);
    }

    public void start(long timeoutSeconds) {
        io.daytona.api.client.model.Sandbox response = ExceptionMapper.callMain(() -> sandboxApi.startSandbox(id, null));
        if (response != null) {
            updateFromModel(response);
        }
        waitUntilStarted(timeoutSeconds);
    }

    public void stop() {
        stop(60);
    }

    public void stop(long timeoutSeconds) {
        ExceptionMapper.callMain(() -> sandboxApi.stopSandbox(id, null, null));
        refreshData();
        waitUntilStopped(timeoutSeconds);
    }

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

    public void delete() {
        delete(60);
    }

    public void delete(long timeoutSeconds) {
        ExceptionMapper.callMain(() -> sandboxApi.deleteSandbox(id, null));
    }

    public Map<String, String> setLabels(Map<String, String> labels) {
        ExceptionMapper.callMain(() -> {
            okhttp3.Call call = sandboxApi.replaceLabelsCall(id, new SandboxLabels().labels(labels), null, null);
            sandboxApi.getApiClient().execute(call, null);
            return null;
        });
        refreshData();
        return this.labels;
    }

    public void setAutostopInterval(int minutes) {
        io.daytona.api.client.model.Sandbox response = ExceptionMapper.callMain(() -> sandboxApi.setAutostopInterval(id, BigDecimal.valueOf(minutes), null));
        if (response != null) {
            updateFromModel(response);
        }
    }

    public void setAutoArchiveInterval(int minutes) {
        io.daytona.api.client.model.Sandbox response = ExceptionMapper.callMain(() -> sandboxApi.setAutoArchiveInterval(id, BigDecimal.valueOf(minutes), null));
        if (response != null) {
            updateFromModel(response);
        }
    }

    public void setAutoDeleteInterval(int minutes) {
        io.daytona.api.client.model.Sandbox response = ExceptionMapper.callMain(() -> sandboxApi.setAutoDeleteInterval(id, BigDecimal.valueOf(minutes), null));
        if (response != null) {
            updateFromModel(response);
        }
    }

    public String getUserHomeDir() {
        io.daytona.toolbox.client.model.UserHomeDirResponse value = ExceptionMapper.callToolbox(() -> infoApi.getUserHomeDir());
        return value == null ? "" : asString(value.getDir());
    }

    public String getWorkDir() {
        io.daytona.toolbox.client.model.WorkDirResponse value = ExceptionMapper.callToolbox(() -> infoApi.getWorkDir());
        return value == null ? "" : asString(value.getDir());
    }

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

    public String getId() { return id; }
    public String getName() { return name; }
    public String getState() { return state; }
    public String getTarget() { return target; }
    public String getUser() { return user; }
    public String getToolboxProxyUrl() { return toolboxProxyUrl; }
    public int getCpu() { return cpu; }
    public int getGpu() { return gpu; }
    public int getMemory() { return memory; }
    public int getDisk() { return disk; }
    public Map<String, String> getEnv() { return env; }
    public Map<String, String> getLabels() { return labels; }
    public Integer getAutoStopInterval() { return autoStopInterval; }
    public Integer getAutoArchiveInterval() { return autoArchiveInterval; }
    public Integer getAutoDeleteInterval() { return autoDeleteInterval; }

    public SandboxProcess getProcess() { return process; }
    public SandboxFileSystem getFs() { return fs; }
    public SandboxGit getGit() { return git; }
    io.daytona.toolbox.client.ApiClient getToolboxApiClient() { return toolboxApiClient; }
    String getApiKey() { return apiKey; }
}
