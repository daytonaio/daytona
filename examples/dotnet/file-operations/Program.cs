// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

using System.Text;
using Daytona.Sdk;
using Daytona.Sdk.Models;

using var daytona = new DaytonaClient();

Console.WriteLine("Creating sandbox");
var sandbox = await daytona.CreateAsync(new CreateSandboxFromSnapshotParams());
Console.WriteLine($"Created sandbox with ID: {sandbox.Id}");

try
{
    Console.WriteLine("Creating folder /tmp/test-dir");
    await sandbox.Fs.CreateFolderAsync("/tmp/test-dir", "755");

    Console.WriteLine("Uploading hello.txt");
    var helloBytes = Encoding.UTF8.GetBytes("Hello, Daytona!");
    await sandbox.Fs.UploadFileAsync(helloBytes, "/tmp/test-dir/hello.txt");

    var files = await sandbox.Fs.ListFilesAsync("/tmp/test-dir");
    Console.WriteLine($"Files in /tmp/test-dir: {string.Join(", ", files.Select(f => f.Name))}");

    Console.WriteLine("Downloading hello.txt");
    var downloaded = await sandbox.Fs.DownloadFileAsync("/tmp/test-dir/hello.txt");
    Console.WriteLine($"Content: {Encoding.UTF8.GetString(downloaded)}");

    Console.WriteLine("Uploading config.env and replacing 'dev' with 'prod'");
    await sandbox.Fs.UploadFileAsync(Encoding.UTF8.GetBytes("env=dev"), "/tmp/test-dir/config.env");
    await sandbox.Fs.ReplaceInFilesAsync(["/tmp/test-dir/config.env"], "dev", "prod");
    var updatedConfig = await sandbox.Fs.DownloadFileAsync("/tmp/test-dir/config.env");
    Console.WriteLine($"Updated config: {Encoding.UTF8.GetString(updatedConfig)}");
}
finally
{
    Console.WriteLine("Deleting sandbox");
    await sandbox.DeleteAsync();
}

return 0;
