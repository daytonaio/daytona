// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

using System.Net.WebSockets;
using System.Text;
using Daytona.Sdk.Models;
using CreateSessionRequestModel = Daytona.ToolboxApiClient.Model.CreateSessionRequest;
using ExecuteRequestModel = Daytona.ToolboxApiClient.Model.ExecuteRequest;
using ProcessApi = Daytona.ToolboxApiClient.Api.ProcessApi;
using PtyCreateRequestModel = Daytona.ToolboxApiClient.Model.PtyCreateRequest;
using PtyResizeRequestModel = Daytona.ToolboxApiClient.Model.PtyResizeRequest;
using SessionExecuteRequestModel = Daytona.ToolboxApiClient.Model.SessionExecuteRequest;

namespace Daytona.Sdk;

public class SandboxProcess
{
    private static readonly byte[] StdoutPrefix = [0x01, 0x01, 0x01];
    private static readonly byte[] StderrPrefix = [0x02, 0x02, 0x02];

    private readonly Sandbox _sandbox;
    private readonly ProcessApi _api;

    internal SandboxProcess(Sandbox sandbox, ProcessApi api)
    {
        _sandbox = sandbox;
        _api = api;
    }

    public async Task<ExecuteResponse> ExecuteCommandAsync(
        string command,
        string? cwd = null,
        Dictionary<string, string>? env = null,
        int? timeout = null,
        CancellationToken ct = default)
    {
        var effectiveCommand = MergeEnvironmentWithCommand(command, env);
        var response = await GeneratedClientSupport.ExecuteToolboxAsync(
            () => _api.ExecuteCommandAsync(new ExecuteRequestModel(effectiveCommand, cwd, timeout ?? 10), ct),
            ct
        );

        return GeneratedClientSupport.ToSdkExecuteResponse(response);
    }

    public async Task<ExecuteResponse> CodeRunAsync(string code, CancellationToken ct = default)
    {
        var language = _sandbox.Labels.TryGetValue("code-toolbox-language", out var value)
            ? value
            : "python";

        var extension = language switch
        {
            "javascript" => "js",
            "typescript" => "ts",
            _ => "py"
        };

        var interpreter = language switch
        {
            "javascript" => "node",
            "typescript" => "npx ts-node",
            _ => "python3"
        };

        var token = $"DAYTONA_EOF_{Guid.NewGuid():N}";
        var tempFile = $"/tmp/daytona_code.{extension}";
        var command = $"cat <<'{token}' > {tempFile}\n{code}\n{token}\n{interpreter} {tempFile}";
        return await ExecuteCommandAsync(command, timeout: 60, ct: ct);
    }

    public Task CreateSessionAsync(string sessionId, CancellationToken ct = default)
        => GeneratedClientSupport.ExecuteToolboxAsync(
            () => _api.CreateSessionAsync(new CreateSessionRequestModel(sessionId), ct),
            ct
        );

    public async Task<Session> GetSessionAsync(string sessionId, CancellationToken ct = default)
    {
        var session = await GeneratedClientSupport.ExecuteToolboxAsync(
            () => _api.GetSessionAsync(sessionId, ct),
            ct
        );
        return GeneratedClientSupport.ToSdkSession(session);
    }

    public async Task<SessionExecuteResponse> ExecuteSessionCommandAsync(
        string sessionId,
        SessionExecuteRequest req,
        CancellationToken ct = default)
    {
        var response = await GeneratedClientSupport.ExecuteToolboxAsync(
            () => _api.SessionExecuteCommandAsync(
                sessionId,
                new SessionExecuteRequestModel(
                    async: false,
                    command: req.Command,
                    runAsync: req.RunAsync ?? false,
                    suppressInputEcho: req.SuppressInputEcho ?? false
                ),
                ct
            ),
            ct
        );

        var mapped = GeneratedClientSupport.ToSdkSessionExecuteResponse(response);

        if (string.IsNullOrEmpty(mapped.Output))
        {
            return mapped;
        }

        var demuxed = DemuxOutput(mapped.Output);
        return mapped with
        {
            Stdout = string.IsNullOrEmpty(mapped.Stdout) ? demuxed.Stdout : mapped.Stdout,
            Stderr = string.IsNullOrEmpty(mapped.Stderr) ? demuxed.Stderr : mapped.Stderr
        };
    }

    public async Task<Command> GetSessionCommandAsync(string sessionId, string commandId, CancellationToken ct = default)
    {
        var command = await GeneratedClientSupport.ExecuteToolboxAsync(
            () => _api.GetSessionCommandAsync(sessionId, commandId, ct),
            ct
        );
        return GeneratedClientSupport.ToSdkCommand(command);
    }

    public async Task<string> GetSessionCommandLogsAsync(string sessionId, string commandId, CancellationToken ct = default)
    {
        var output = await GeneratedClientSupport.ExecuteToolboxAsync(
            () => _api.GetSessionCommandLogsAsync(sessionId, commandId, cancellationToken: ct),
            ct
        );

        var demux = DemuxOutput(output);
        return string.IsNullOrEmpty(demux.Stdout) && string.IsNullOrEmpty(demux.Stderr)
            ? output
            : demux.Stdout + demux.Stderr;
    }

    public Task DeleteSessionAsync(string sessionId, CancellationToken ct = default)
        => GeneratedClientSupport.ExecuteToolboxAsync(
            () => _api.DeleteSessionAsync(sessionId, ct),
            ct
        );

    public async Task<PtyHandle> CreatePtyAsync(PtyCreateOptions options, CancellationToken ct = default)
    {
        await GeneratedClientSupport.ExecuteToolboxAsync(
            () => _api.CreatePtySessionAsync(new PtyCreateRequestModel(
                cols: options.Cols,
                rows: options.Rows,
                id: options.Id
            ), ct),
            ct
        );

        var baseUrl = _sandbox.ToolboxBaseUrl;
        var wsUrl = baseUrl
            .Replace("https://", "wss://")
            .Replace("http://", "ws://");
        wsUrl = $"{wsUrl}/process/pty/{Uri.EscapeDataString(options.Id)}/connect";

        var ws = new ClientWebSocket();
        ws.Options.SetRequestHeader("Authorization", $"Bearer {_sandbox.ApiKey}");
        await ws.ConnectAsync(new Uri(wsUrl), ct);

        return new PtyHandle(
            ws,
            options.Id,
            (cols, rows) => ResizePtySessionAsync(options.Id, cols, rows),
            () => KillPtySessionAsync(options.Id),
            options.OnData
        );
    }

    public Task ResizePtySessionAsync(string sessionId, int cols, int rows, CancellationToken ct = default)
        => GeneratedClientSupport.ExecuteToolboxAsync(
            () => _api.ResizePtySessionAsync(sessionId, new PtyResizeRequestModel(cols, rows), ct),
            ct
        );

    public Task KillPtySessionAsync(string sessionId, CancellationToken ct = default)
        => GeneratedClientSupport.ExecuteToolboxAsync(
            () => _api.DeletePtySessionAsync(sessionId, ct),
            ct
        );

    private static string MergeEnvironmentWithCommand(string command, Dictionary<string, string>? env)
    {
        if (env is null || env.Count == 0)
        {
            return command;
        }

        var envPrefix = string.Join(' ', env.Select(kv => $"{kv.Key}={EscapeShellArg(kv.Value)}"));
        return $"{envPrefix} {command}";
    }

    private static string EscapeShellArg(string value)
    {
        var escaped = value.Replace("'", "'\"'\"'");
        return $"'{escaped}'";
    }

    private static (string Stdout, string Stderr) DemuxOutput(string output)
    {
        var bytes = Encoding.UTF8.GetBytes(output);
        var stdout = new MemoryStream();
        var stderr = new MemoryStream();
        var target = stdout;

        var i = 0;
        while (i < bytes.Length)
        {
            if (HasPrefix(bytes, i, StdoutPrefix))
            {
                target = stdout;
                i += StdoutPrefix.Length;
                continue;
            }

            if (HasPrefix(bytes, i, StderrPrefix))
            {
                target = stderr;
                i += StderrPrefix.Length;
                continue;
            }

            target.WriteByte(bytes[i]);
            i++;
        }

        return (Encoding.UTF8.GetString(stdout.ToArray()), Encoding.UTF8.GetString(stderr.ToArray()));
    }

    private static bool HasPrefix(byte[] source, int offset, byte[] prefix)
    {
        if (offset + prefix.Length > source.Length)
        {
            return false;
        }

        for (var i = 0; i < prefix.Length; i++)
        {
            if (source[offset + i] != prefix[i])
            {
                return false;
            }
        }

        return true;
    }
}