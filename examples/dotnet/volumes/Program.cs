// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

using Daytona.Sdk;

using var daytona = new DaytonaClient();

var volumeName = $"test-vol-{DateTimeOffset.UtcNow.ToUnixTimeMilliseconds()}";
var created = await daytona.Volume.CreateAsync(volumeName);

try
{
    var fetched = await daytona.Volume.GetByNameAsync(volumeName);
    Console.WriteLine($"id: {fetched.Id}");
    Console.WriteLine($"name: {fetched.Name}");
    Console.WriteLine($"state: {fetched.State}");
}
finally
{
    Console.WriteLine("Deleting volume");
    await WaitUntilDeletableAsync(daytona, volumeName);
    await daytona.Volume.DeleteAsync(created.Id);
}

static async Task WaitUntilDeletableAsync(DaytonaClient daytona, string volumeName)
{
    var start = DateTimeOffset.UtcNow;
    while (DateTimeOffset.UtcNow - start < TimeSpan.FromSeconds(60))
    {
        var volume = await daytona.Volume.GetByNameAsync(volumeName);
        if (string.Equals(volume.State, "ready", StringComparison.OrdinalIgnoreCase)
            || string.Equals(volume.State, "error", StringComparison.OrdinalIgnoreCase))
        {
            return;
        }

        await Task.Delay(1000);
    }

    throw new InvalidOperationException("Timed out waiting for volume to become deletable.");
}

return 0;
