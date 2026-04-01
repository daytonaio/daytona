// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

using Daytona.Sdk.Models;
using CreateSnapshotModel = Daytona.ApiClient.Model.CreateSnapshot;
using SnapshotsApi = Daytona.ApiClient.Api.SnapshotsApi;

namespace Daytona.Sdk;

public class SnapshotService
{
    private readonly SnapshotsApi _api;

    internal SnapshotService(SnapshotsApi api)
    {
        _api = api;
    }

    public async Task<Snapshot> CreateAsync(string name, string? image = null, Resources? resources = null, CancellationToken ct = default)
    {
        var createSnapshot = new CreateSnapshotModel(
            name,
            imageName: image!,
            cpu: resources?.Cpu ?? default,
            gpu: resources?.Gpu ?? default,
            memory: resources?.Memory ?? default,
            disk: resources?.Disk ?? default
        );

        var snapshot = await GeneratedClientSupport.ExecuteMainAsync(
            () => _api.CreateSnapshotAsync(createSnapshot, cancellationToken: ct),
            ct
        );

        return GeneratedClientSupport.ToSdkSnapshot(snapshot);
    }

    public async Task<PaginatedResponse<Snapshot>> ListAsync(int page = 1, int limit = 10, CancellationToken ct = default)
    {
        var response = await GeneratedClientSupport.ExecuteMainAsync(
            () => _api.GetAllSnapshotsAsync(page: page, limit: limit, cancellationToken: ct),
            ct
        );

        return new PaginatedResponse<Snapshot>
        {
            Items = response.Items?.Select(GeneratedClientSupport.ToSdkSnapshot).ToList() ?? new List<Snapshot>(),
            Total = (int)response.Total,
            Page = (int)response.Page,
            TotalPages = (int)response.TotalPages
        };
    }

    public async Task<Snapshot> GetAsync(string nameOrId, CancellationToken ct = default)
    {
        var snapshot = await GeneratedClientSupport.ExecuteMainAsync(
            () => _api.GetSnapshotAsync(nameOrId, cancellationToken: ct),
            ct
        );

        return GeneratedClientSupport.ToSdkSnapshot(snapshot);
    }

    public Task DeleteAsync(string id, CancellationToken ct = default)
        => GeneratedClientSupport.ExecuteMainAsync(
            () => _api.RemoveSnapshotAsync(id, cancellationToken: ct),
            ct
        );
}