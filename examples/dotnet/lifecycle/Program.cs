// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

using Daytona.Sdk;
using Daytona.Sdk.Models;

using var daytona = new DaytonaClient();

Console.WriteLine("Creating sandbox");
var sandbox = await daytona.CreateAsync(new CreateSandboxFromSnapshotParams());
Console.WriteLine($"Sandbox created: {sandbox.Id} (state: {sandbox.State})");

var labels = await sandbox.SetLabelsAsync(new Dictionary<string, string> { ["test"] = "lifecycle" });
Console.WriteLine($"Labels set: {string.Join(", ", labels.Select(kv => $"{kv.Key}={kv.Value}"))}");

Console.WriteLine("Stopping sandbox");
await sandbox.StopAsync();
await sandbox.WaitUntilStoppedAsync(60);
Console.WriteLine("Sandbox stopped");

Console.WriteLine("Starting sandbox");
await sandbox.StartAsync();
await sandbox.WaitUntilStartedAsync(60);
Console.WriteLine("Sandbox started");

Console.WriteLine("Getting existing sandbox");
var fetched = await daytona.GetAsync(sandbox.Id);
Console.WriteLine($"Got sandbox: {fetched.Id} (state: {fetched.State})");

var list = await daytona.ListAsync();
Console.WriteLine($"Total sandboxes: {list.Total}");

Console.WriteLine("Deleting sandbox");
await sandbox.DeleteAsync();
Console.WriteLine("Sandbox deleted");

return 0;
