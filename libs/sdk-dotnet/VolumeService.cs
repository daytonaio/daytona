// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

using Daytona.Sdk.Models;
using CreateVolumeModel = Daytona.ApiClient.Model.CreateVolume;
using VolumesApi = Daytona.ApiClient.Api.VolumesApi;

namespace Daytona.Sdk;

public class VolumeService
{
    private readonly VolumesApi _api;

    internal VolumeService(VolumesApi api)
    {
        _api = api;
    }

    public async Task<Volume> CreateAsync(string name, CancellationToken ct = default)
    {
        var volume = await GeneratedClientSupport.ExecuteMainAsync(
            () => _api.CreateVolumeAsync(new CreateVolumeModel(name), cancellationToken: ct),
            ct
        );

        return GeneratedClientSupport.ToSdkVolume(volume);
    }

    public async Task<List<Volume>> ListAsync(CancellationToken ct = default)
    {
        var volumes = await GeneratedClientSupport.ExecuteMainAsync(
            () => _api.ListVolumesAsync(cancellationToken: ct),
            ct
        );

        return volumes.Select(GeneratedClientSupport.ToSdkVolume).ToList();
    }

    public async Task<Volume> GetByNameAsync(string name, CancellationToken ct = default)
    {
        var volume = await GeneratedClientSupport.ExecuteMainAsync(
            () => _api.GetVolumeByNameAsync(name, cancellationToken: ct),
            ct
        );

        return GeneratedClientSupport.ToSdkVolume(volume);
    }

    public Task DeleteAsync(string id, CancellationToken ct = default)
        => GeneratedClientSupport.ExecuteMainAsync(
            () => _api.DeleteVolumeAsync(id, cancellationToken: ct),
            ct
        );
}