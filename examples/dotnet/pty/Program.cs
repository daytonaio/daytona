// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

using System.Text;
using Daytona.Sdk;

using var daytona = new DaytonaClient();
var sandbox = await daytona.CreateAsync();
try
{
    Console.WriteLine("=== First PTY Session: Interactive Command with Exit ===");
    await using var handle = await sandbox.Process.CreatePtyAsync(new PtyCreateOptions
    {
        Id = "interactive-pty",
        Cols = 120,
        Rows = 30,
        OnData = data =>
        {
            Console.Write(Encoding.UTF8.GetString(data));
            return Task.CompletedTask;
        }
    });

    await handle.WaitForConnectionAsync(10);
    Console.WriteLine("\nSending interactive command...");
    await handle.SendInputAsync("echo 'Hello from PTY!'\n");
    await Task.Delay(1000);
    await handle.SendInputAsync("exit\n");

    var result = await handle.WaitAsync(10);
    Console.WriteLine($"\nPTY session exited with code: {result.ExitCode}");

    Console.WriteLine("\n=== Second PTY Session: Kill PTY Session ===");
    await using var handle2 = await sandbox.Process.CreatePtyAsync(new PtyCreateOptions
    {
        Id = "kill-pty",
        Cols = 120,
        Rows = 30,
        OnData = _ => Task.CompletedTask
    });
    await handle2.WaitForConnectionAsync(10);
    Console.WriteLine("Sending long-running command (infinite loop)...");
    await handle2.SendInputAsync("while true; do echo running; sleep 1; done\n");
    await Task.Delay(2000);
    await handle2.KillAsync();
    var result2 = await handle2.WaitAsync(10);
    Console.WriteLine($"\nPTY session terminated. Exit code: {result2.ExitCode}");
}
finally
{
    Console.WriteLine($"\nDeleting sandbox: {sandbox.Id}");
    await sandbox.DeleteAsync();
}

return 0;
