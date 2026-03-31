// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

using Daytona.Sdk.Exceptions;
using Daytona.Sdk.Models;
using MainSandboxApi = Daytona.ApiClient.Api.SandboxApi;
using InfoApi = Daytona.ToolboxApiClient.Api.InfoApi;
using SandboxLabelsModel = Daytona.ApiClient.Model.SandboxLabels;
using SandboxModel = Daytona.ApiClient.Model.Sandbox;

namespace Daytona.Sdk;

public class Sandbox
{
    private readonly MainSandboxApi _mainApi;
    private readonly InfoApi _infoApi;
    private readonly string _apiKey;

    public string Id { get; private set; }
    public string Name { get; private set; }
    public string State { get; private set; }
    public string? User { get; private set; }
    public string? Target { get; private set; }
    public int? Cpu { get; private set; }
    public int? Gpu { get; private set; }
    public int? Memory { get; private set; }
    public int? Disk { get; private set; }
    public Dictionary<string, string> Env { get; private set; }
    public Dictionary<string, string> Labels { get; private set; }
    public bool Public { get; private set; }
    public string ToolboxProxyUrl { get; private set; }
    public int? AutoStopInterval { get; private set; }
    public int? AutoArchiveInterval { get; private set; }
    public int? AutoDeleteInterval { get; private set; }
    public bool? NetworkBlockAll { get; private set; }
    public string? NetworkAllowList { get; private set; }

    public SandboxProcess Process { get; }
    public SandboxFileSystem Fs { get; }
    public SandboxGit Git { get; }

    internal string ApiKey => _apiKey;
    internal string ToolboxBaseUrl { get; private set; }

    internal Sandbox(string apiKey, MainSandboxApi mainApi, SandboxModel sandbox)
    {
        _apiKey = apiKey;
        _mainApi = mainApi;

        var toolboxBaseUrl = $"{sandbox.ToolboxProxyUrl}/{sandbox.Id}";
        ToolboxBaseUrl = toolboxBaseUrl;
        var toolboxConfig = GeneratedClientSupport.CreateToolboxConfiguration(toolboxBaseUrl, _apiKey);
        _infoApi = new InfoApi(toolboxConfig);

        Apply(sandbox);

        Process = new SandboxProcess(this, new Daytona.ToolboxApiClient.Api.ProcessApi(toolboxConfig));
        Fs = new SandboxFileSystem(new Daytona.ToolboxApiClient.Api.FileSystemApi(toolboxConfig));
        Git = new SandboxGit(new Daytona.ToolboxApiClient.Api.GitApi(toolboxConfig));
    }

    public async Task StartAsync(int timeoutSecs = 60, CancellationToken ct = default)
    {
        var started = await GeneratedClientSupport.ExecuteMainAsync(
            () => _mainApi.StartSandboxAsync(Id, cancellationToken: ct),
            ct
        );
        Apply(started);
        await WaitUntilStartedAsync(timeoutSecs, ct);
    }

    public async Task StopAsync(int timeoutSecs = 60, CancellationToken ct = default)
    {
        await GeneratedClientSupport.ExecuteMainAsync(
            () => _mainApi.StopSandboxAsync(Id, cancellationToken: ct),
            ct
        );
        await WaitUntilStoppedAsync(timeoutSecs, ct);
    }

    public Task DeleteAsync(CancellationToken ct = default)
        => GeneratedClientSupport.ExecuteMainAsync(
            () => _mainApi.DeleteSandboxAsync(Id, cancellationToken: ct),
            ct
        );

    public async Task<Dictionary<string, string>> SetLabelsAsync(Dictionary<string, string> labels, CancellationToken ct = default)
    {
        var response = await GeneratedClientSupport.ExecuteMainAsync(
            () => _mainApi.ReplaceLabelsAsync(Id, new SandboxLabelsModel(labels), cancellationToken: ct),
            ct
        );

        Labels = response?.Labels ?? new Dictionary<string, string>();
        return Labels;
    }

    public async Task<string> GetUserHomeDirAsync(CancellationToken ct = default)
    {
        var response = await GeneratedClientSupport.ExecuteToolboxAsync(
            () => _infoApi.GetUserHomeDirAsync(ct),
            ct
        );
        return response?.Dir ?? string.Empty;
    }

    public async Task<string> GetWorkDirAsync(CancellationToken ct = default)
    {
        var response = await GeneratedClientSupport.ExecuteToolboxAsync(
            () => _infoApi.GetWorkDirAsync(ct),
            ct
        );
        return response?.Dir ?? string.Empty;
    }

    public async Task RefreshDataAsync(CancellationToken ct = default)
    {
        var latest = await GeneratedClientSupport.ExecuteMainAsync(
            () => _mainApi.GetSandboxAsync(Id, cancellationToken: ct),
            ct
        );
        Apply(latest);
    }

    public async Task WaitUntilStartedAsync(int timeoutSecs, CancellationToken ct = default)
    {
        var start = DateTimeOffset.UtcNow;
        while (!string.Equals(State, "started", StringComparison.OrdinalIgnoreCase))
        {
            await RefreshDataAsync(ct);
            if (string.Equals(State, "started", StringComparison.OrdinalIgnoreCase))
            {
                return;
            }

            if (string.Equals(State, "error", StringComparison.OrdinalIgnoreCase)
                || string.Equals(State, "build_failed", StringComparison.OrdinalIgnoreCase))
            {
                throw new DaytonaException($"Sandbox {Id} entered terminal state '{State}'.");
            }

            if (timeoutSecs > 0 && DateTimeOffset.UtcNow - start > TimeSpan.FromSeconds(timeoutSecs))
            {
                throw new DaytonaTimeoutException("Sandbox failed to reach started state within timeout.");
            }

            await Task.Delay(250, ct);
        }
    }

    public async Task WaitUntilStoppedAsync(int timeoutSecs, CancellationToken ct = default)
    {
        var start = DateTimeOffset.UtcNow;
        while (!string.Equals(State, "stopped", StringComparison.OrdinalIgnoreCase))
        {
            await RefreshDataAsync(ct);
            if (string.Equals(State, "stopped", StringComparison.OrdinalIgnoreCase))
            {
                return;
            }

            if (string.Equals(State, "error", StringComparison.OrdinalIgnoreCase)
                || string.Equals(State, "build_failed", StringComparison.OrdinalIgnoreCase))
            {
                throw new DaytonaException($"Sandbox {Id} entered terminal state '{State}'.");
            }

            if (timeoutSecs > 0 && DateTimeOffset.UtcNow - start > TimeSpan.FromSeconds(timeoutSecs))
            {
                throw new DaytonaTimeoutException("Sandbox failed to reach stopped state within timeout.");
            }

            await Task.Delay(250, ct);
        }
    }

    private void Apply(SandboxModel info)
    {
        var mapped = GeneratedClientSupport.ToSdkSandboxInfo(info);

        Id = mapped.Id;
        Name = mapped.Name;
        State = mapped.State;
        User = mapped.User;
        Target = mapped.Target;
        Cpu = mapped.Cpu;
        Gpu = mapped.Gpu;
        Memory = mapped.Memory;
        Disk = mapped.Disk;
        Env = mapped.Env;
        Labels = mapped.Labels;
        Public = mapped.Public;
        ToolboxProxyUrl = mapped.ToolboxProxyUrl;
        AutoStopInterval = mapped.AutoStopInterval;
        AutoArchiveInterval = mapped.AutoArchiveInterval;
        AutoDeleteInterval = mapped.AutoDeleteInterval;
        NetworkBlockAll = mapped.NetworkBlockAll;
        NetworkAllowList = mapped.NetworkAllowList;
    }
}