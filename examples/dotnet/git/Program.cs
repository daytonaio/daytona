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
    var repoPath = "/home/daytona/learn-typescript";
    Console.WriteLine($"Cloning repository to {repoPath}");
    await sandbox.Git.CloneAsync("https://github.com/panaverse/learn-typescript", repoPath);

    var branches = await sandbox.Git.BranchesAsync(repoPath);
    Console.WriteLine($"Branches: {string.Join(", ", branches.Branches)}");

    var status = await sandbox.Git.StatusAsync(repoPath);
    Console.WriteLine($"Current branch: {status.CurrentBranch}");
}
finally
{
    Console.WriteLine("Deleting sandbox");
    await sandbox.DeleteAsync();
}

return 0;
