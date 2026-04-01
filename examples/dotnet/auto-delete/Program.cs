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
    Console.WriteLine($"autoDeleteInterval: {sandbox.AutoDeleteInterval}");

    await SetAutoDeleteIntervalAsync(sandbox, 60);
    Console.WriteLine($"autoDeleteInterval: {sandbox.AutoDeleteInterval}");

    await SetAutoDeleteIntervalAsync(sandbox, 0);
    Console.WriteLine($"autoDeleteInterval: {sandbox.AutoDeleteInterval}");

    await SetAutoDeleteIntervalAsync(sandbox, -1);
    Console.WriteLine($"autoDeleteInterval: {sandbox.AutoDeleteInterval}");
}
finally
{
    Console.WriteLine("Deleting sandbox");
    await sandbox.DeleteAsync();
}

static async Task SetAutoDeleteIntervalAsync(Sandbox sandbox, int interval)
{
    var api = CreateSandboxApi();
    await api.SetAutoDeleteIntervalAsync(sandbox.Id, interval);
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
