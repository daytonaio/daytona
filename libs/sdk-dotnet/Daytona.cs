// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

using System.Text.Json;
using Daytona.Sdk.Exceptions;
using Daytona.Sdk.Models;
using MainSandboxApi = Daytona.ApiClient.Api.SandboxApi;
using MainSnapshotsApi = Daytona.ApiClient.Api.SnapshotsApi;
using MainVolumesApi = Daytona.ApiClient.Api.VolumesApi;
using CreateSandboxModel = Daytona.ApiClient.Model.CreateSandbox;
using CreateBuildInfoModel = Daytona.ApiClient.Model.CreateBuildInfo;
using SandboxVolumeModel = Daytona.ApiClient.Model.SandboxVolume;
using PaginatedSandboxesModel = Daytona.ApiClient.Model.PaginatedSandboxes;

namespace Daytona.Sdk;

public class DaytonaClient : IDisposable
{
    private readonly DaytonaConfig _config;
    private readonly MainSandboxApi _sandboxApi;
    private readonly MainSnapshotsApi _snapshotsApi;
    private readonly MainVolumesApi _volumesApi;
    private readonly string _apiKey;

    public SnapshotService Snapshot { get; }
    public VolumeService Volume { get; }

    public DaytonaClient()
        : this(new DaytonaConfig
        {
            ApiKey = Environment.GetEnvironmentVariable("DAYTONA_API_KEY"),
            ApiUrl = Environment.GetEnvironmentVariable("DAYTONA_API_URL") ?? "https://app.daytona.io/api",
            Target = Environment.GetEnvironmentVariable("DAYTONA_TARGET")
        })
    {
    }

    public DaytonaClient(DaytonaConfig config)
    {
        _config = config;
        _apiKey = config.ApiKey ?? Environment.GetEnvironmentVariable("DAYTONA_API_KEY") ?? string.Empty;
        if (string.IsNullOrWhiteSpace(_apiKey))
        {
            throw new DaytonaException("DAYTONA_API_KEY is required.");
        }

        _config.ApiUrl = string.IsNullOrWhiteSpace(_config.ApiUrl) ? "https://app.daytona.io/api" : _config.ApiUrl;
        _config.Target ??= Environment.GetEnvironmentVariable("DAYTONA_TARGET");

        var mainConfig = GeneratedClientSupport.CreateMainConfiguration(_config.ApiUrl, _apiKey);
        _sandboxApi = new MainSandboxApi(mainConfig);
        _snapshotsApi = new MainSnapshotsApi(mainConfig);
        _volumesApi = new MainVolumesApi(mainConfig);

        Snapshot = new SnapshotService(_snapshotsApi);
        Volume = new VolumeService(_volumesApi);
    }

    public Task<Sandbox> CreateAsync(CreateSandboxFromSnapshotParams? p = null, int timeoutSecs = 60, CancellationToken ct = default)
    {
        var labels = p?.Labels is null ? new Dictionary<string, string>() : new Dictionary<string, string>(p.Labels);
        if (p?.Language is not null)
        {
            labels["code-toolbox-language"] = LanguageToString(p.Language.Value);
        }

        var createSandbox = new CreateSandboxModel
        {
            Name = p?.Name,
            Snapshot = p?.Snapshot,
            User = p?.User,
            Env = p?.Env,
            Labels = labels,
            Target = _config.Target,
            Public = p?.Public ?? default,
            AutoStopInterval = p?.AutoStopInterval ?? default,
            AutoArchiveInterval = p?.AutoArchiveInterval ?? default,
            AutoDeleteInterval = p?.AutoDeleteInterval ?? default,
            Volumes = p?.Volumes?.Select(volume => new SandboxVolumeModel(volume.VolumeId, volume.MountPath, volume.SubPath)).ToList(),
            NetworkBlockAll = p?.NetworkBlockAll ?? default,
            NetworkAllowList = p?.NetworkAllowList
        };

        return CreateInternalAsync(createSandbox, timeoutSecs, ct);
    }

    public Task<Sandbox> CreateAsync(CreateSandboxFromImageParams p, int timeoutSecs = 60, CancellationToken ct = default)
    {
        var labels = p.Labels is null ? new Dictionary<string, string>() : new Dictionary<string, string>(p.Labels);
        if (p.Language is not null)
        {
            labels["code-toolbox-language"] = LanguageToString(p.Language.Value);
        }

        var createSandbox = new CreateSandboxModel
        {
            Name = p.Name,
            User = p.User,
            Env = p.Env,
            Labels = labels,
            Target = _config.Target,
            Public = p.Public ?? default,
            AutoStopInterval = p.AutoStopInterval ?? default,
            AutoArchiveInterval = p.AutoArchiveInterval ?? default,
            AutoDeleteInterval = p.AutoDeleteInterval ?? default,
            Volumes = p.Volumes?.Select(volume => new SandboxVolumeModel(volume.VolumeId, volume.MountPath, volume.SubPath)).ToList(),
            NetworkBlockAll = p.NetworkBlockAll ?? default,
            NetworkAllowList = p.NetworkAllowList
        };

        if (p.Resources is not null)
        {
            createSandbox.Cpu = p.Resources.Cpu ?? default;
            createSandbox.Gpu = p.Resources.Gpu ?? default;
            createSandbox.Memory = p.Resources.Memory ?? default;
            createSandbox.Disk = p.Resources.Disk ?? default;
        }

        if (p.Image is Image image)
        {
            createSandbox.BuildInfo = new CreateBuildInfoModel(image.Dockerfile);
        }
        else
        {
            createSandbox.Snapshot = p.Image.ToString();
        }

        return CreateInternalAsync(createSandbox, timeoutSecs, ct);
    }

    public async Task<Sandbox> GetAsync(string sandboxIdOrName, CancellationToken ct = default)
    {
        var dto = await GeneratedClientSupport.ExecuteMainAsync(
            () => _sandboxApi.GetSandboxAsync(sandboxIdOrName, cancellationToken: ct),
            ct
        );
        return BuildSandbox(dto);
    }

    public async Task<PaginatedResponse<SandboxInfo>> ListAsync(Dictionary<string, string>? labels = null, int page = 1, int limit = 10, CancellationToken ct = default)
    {
        var labelsJson = JsonSerializer.Serialize(labels ?? new Dictionary<string, string>());
        var response = await GeneratedClientSupport.ExecuteMainAsync(
            () => _sandboxApi.ListSandboxesPaginatedAsync(page: page, limit: limit, labels: labelsJson, cancellationToken: ct),
            ct
        );

        return ToSdkPaginatedResponse(response);
    }

    private async Task<Sandbox> CreateInternalAsync(CreateSandboxModel body, int timeoutSecs, CancellationToken ct)
    {
        var created = await GeneratedClientSupport.ExecuteMainAsync(
            () => _sandboxApi.CreateSandboxAsync(body, cancellationToken: ct),
            ct
        );

        var sandbox = BuildSandbox(created);
        await sandbox.WaitUntilStartedAsync(timeoutSecs, ct);
        return sandbox;
    }

    private Sandbox BuildSandbox(Daytona.ApiClient.Model.Sandbox info) => new(_apiKey, _sandboxApi, info);

    private static PaginatedResponse<SandboxInfo> ToSdkPaginatedResponse(PaginatedSandboxesModel response) => new()
    {
        Items = response.Items?.Select(GeneratedClientSupport.ToSdkSandboxInfo).ToList() ?? new List<SandboxInfo>(),
        Total = (int)response.Total,
        Page = (int)response.Page,
        TotalPages = (int)response.TotalPages
    };

    private static string LanguageToString(CodeLanguage language) => language switch
    {
        CodeLanguage.Python => "python",
        CodeLanguage.JavaScript => "javascript",
        CodeLanguage.TypeScript => "typescript",
        _ => "python"
    };

    public void Dispose()
    {
        _sandboxApi.Dispose();
        _snapshotsApi.Dispose();
        _volumesApi.Dispose();
    }
}