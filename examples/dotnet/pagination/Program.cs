// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

using Daytona.Sdk;

using var daytona = new DaytonaClient();

var sandboxes = await daytona.ListAsync(null, 1, 5);
Console.WriteLine($"Found {sandboxes.Total} sandboxes");
foreach (var sb in sandboxes.Items)
{
    Console.WriteLine($"  {sb.Id}: {sb.State}");
}

var snapshots = await daytona.Snapshot.ListAsync(1, 5);
Console.WriteLine($"Found {snapshots.Total} snapshots");
foreach (var snap in snapshots.Items)
{
    Console.WriteLine($"  {snap.Name} ({snap.ImageName})");
}

return 0;
