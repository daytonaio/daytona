// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

using Daytona.Sdk;
using Daytona.Sdk.Models;

using var daytona = new DaytonaClient(new DaytonaConfig { Target = "us" });

Console.WriteLine("Creating sandbox with target: us");
var sandbox = await daytona.CreateAsync(new CreateSandboxFromSnapshotParams());
Console.WriteLine($"Sandbox created: {sandbox.Id}");
Console.WriteLine($"target: {sandbox.Target}");

try
{
}
finally
{
    Console.WriteLine("Deleting sandbox");
    await sandbox.DeleteAsync();
}

return 0;
