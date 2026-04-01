// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

using Daytona.ApiClient.Api;
using Daytona.ApiClient.Client;
using Daytona.Sdk;
using Daytona.Sdk.Models;

using var daytona = new DaytonaClient();

var sandbox = await daytona.CreateAsync(new CreateSandboxFromSnapshotParams());
try
{
    Console.WriteLine($"autoArchiveInterval: {sandbox.AutoArchiveInterval}");

    await SetAutoArchiveIntervalAsync(sandbox, 60);
    Console.WriteLine($"autoArchiveInterval: {sandbox.AutoArchiveInterval}");
}
finally
{
    Console.WriteLine("Deleting sandbox");
    await sandbox.DeleteAsync();
}

static async Task SetAutoArchiveIntervalAsync(Sandbox sandbox, int interval)
{
    var api = CreateSandboxApi();
    await api.SetAutoArchiveIntervalAsync(sandbox.Id, interval);
    await sandbox.RefreshDataAsync();
}

static SandboxApi CreateSandboxApi()
{
    var apiKey = Environment.GetEnvironmentVariable("DAYTONA_API_KEY") ?? string.Empty;
    var apiUrl = Environment.GetEnvironmentVariable("DAYTONA_API_URL") ?? "https://app.daytona.io/api";
    var config = new Configuration
    {
        BasePath = apiUrl.TrimEnd('/'),
        AccessToken = apiKey
    };

    return new SandboxApi(config);
}

return 0;
