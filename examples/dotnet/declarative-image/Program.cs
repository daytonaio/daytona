// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

using Daytona.Sdk;
using Daytona.Sdk.Models;

using var daytona = new DaytonaClient();

var image = Image
    .DebianSlim("3.12")
    .PipInstall("numpy", "pandas")
    .RunCommands("mkdir -p /home/daytona/workspace")
    .Workdir("/home/daytona/workspace")
    .Env(new Dictionary<string, string> { ["MY_ENV_VAR"] = "test-value" });

var sandbox = await daytona.CreateAsync(new CreateSandboxFromImageParams
{
    Image = image
});

try
{
    Console.WriteLine("Verifying sandbox from declarative image:");

    Console.WriteLine("Python environment:");
    var numpyCheck = await sandbox.Process.ExecuteCommandAsync("python3 -c \"import numpy; print(numpy.__version__)\"");
    Console.WriteLine(numpyCheck.Result.Trim());

    var envCheck = await sandbox.Process.ExecuteCommandAsync("echo $MY_ENV_VAR");
    Console.WriteLine($"MY_ENV_VAR={envCheck.Result.Trim()}");
}
finally
{
    Console.WriteLine("Deleting sandbox");
    await sandbox.DeleteAsync();
}

return 0;
