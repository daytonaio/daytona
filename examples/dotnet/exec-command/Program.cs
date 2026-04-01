// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

using Daytona.Sdk;
using Daytona.Sdk.Models;

using var daytona = new DaytonaClient();

Console.WriteLine("Creating sandbox with Python language");
var sandbox = await daytona.CreateAsync(new CreateSandboxFromSnapshotParams
{
    Language = CodeLanguage.Python
});
Console.WriteLine($"Sandbox created: {sandbox.Id}");

try
{
    var cmd = await sandbox.Process.ExecuteCommandAsync("echo Hello World from CMD!");
    Console.WriteLine(cmd.Result);

    var code = await sandbox.Process.CodeRunAsync("print('Hello World from code!')");
    Console.WriteLine(code.Result);

    Console.WriteLine("Creating session");
    await sandbox.Process.CreateSessionAsync("test-session-1");
    await sandbox.Process.ExecuteSessionCommandAsync("test-session-1", new SessionExecuteRequest
    {
        Command = "export FOO=BAR"
    });

    var echo = await sandbox.Process.ExecuteSessionCommandAsync("test-session-1", new SessionExecuteRequest
    {
        Command = "echo $FOO"
    });

    var stdout = echo.Stdout ?? echo.Output ?? string.Empty;
    Console.WriteLine($"FOO={stdout.Trim()}");

    if (!string.IsNullOrWhiteSpace(echo.CmdId))
    {
        var logs = await sandbox.Process.GetSessionCommandLogsAsync("test-session-1", echo.CmdId);
        Console.WriteLine($"Session command logs: {logs}");
    }

    Console.WriteLine("Deleting session");
    await sandbox.Process.DeleteSessionAsync("test-session-1");
}
finally
{
    Console.WriteLine("Deleting sandbox");
    await sandbox.DeleteAsync();
}

return 0;
