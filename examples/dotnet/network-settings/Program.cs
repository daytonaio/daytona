// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

using Daytona.Sdk;
using Daytona.Sdk.Models;

using var daytona = new DaytonaClient();

Console.WriteLine("Creating sandbox");
var sandbox = await daytona.CreateAsync(new CreateSandboxFromSnapshotParams());
Console.WriteLine($"Sandbox created: {sandbox.Id}");

try
{
    var fetched = await daytona.GetAsync(sandbox.Id);
    Console.WriteLine($"id: {fetched.Id}");
    Console.WriteLine($"state: {fetched.State}");
}
finally
{
    Console.WriteLine("Deleting sandbox");
    await sandbox.DeleteAsync();
}

return 0;
