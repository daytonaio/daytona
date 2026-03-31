// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

using System.Net.WebSockets;
using System.Text;
using System.Text.Json;

namespace Daytona.Sdk;

public class PtyHandle : IAsyncDisposable
{
    private readonly ClientWebSocket _ws;
    private readonly Func<int, int, Task> _resizeCallback;
    private readonly Func<Task> _killCallback;
    private readonly Func<byte[], Task>? _onData;
    private readonly CancellationTokenSource _receiveCts = new();
    private readonly TaskCompletionSource<bool> _connectionTcs = new(TaskCreationOptions.RunContinuationsAsynchronously);
    private readonly TaskCompletionSource<PtyResult> _exitTcs = new(TaskCreationOptions.RunContinuationsAsynchronously);
    private volatile bool _connected;

    public string SessionId { get; }
    public int? ExitCode { get; private set; }
    public string? Error { get; private set; }
    public bool IsConnected => _connected && _ws.State == WebSocketState.Open;

    internal PtyHandle(
        ClientWebSocket ws,
        string sessionId,
        Func<int, int, Task> resizeCallback,
        Func<Task> killCallback,
        Func<byte[], Task>? onData)
    {
        _ws = ws;
        SessionId = sessionId;
        _resizeCallback = resizeCallback;
        _killCallback = killCallback;
        _onData = onData;
        _connected = true;

        _ = Task.Run(ReceiveLoopAsync);
    }

    public async Task WaitForConnectionAsync(int timeoutSeconds = 10, CancellationToken ct = default)
    {
        using var cts = CancellationTokenSource.CreateLinkedTokenSource(ct);
        cts.CancelAfter(TimeSpan.FromSeconds(timeoutSeconds));
        try
        {
            var connected = await _connectionTcs.Task.WaitAsync(cts.Token);
            if (!connected)
                throw new Exceptions.DaytonaException(Error ?? "PTY connection failed");
        }
        catch (OperationCanceledException)
        {
            throw new Exceptions.DaytonaTimeoutException("PTY connection timed out");
        }
    }

    public async Task SendInputAsync(string data, CancellationToken ct = default)
    {
        if (!IsConnected) throw new Exceptions.DaytonaException("PTY is not connected");
        var bytes = Encoding.UTF8.GetBytes(data);
        await _ws.SendAsync(bytes, WebSocketMessageType.Text, true, ct);
    }

    public async Task SendInputAsync(byte[] data, CancellationToken ct = default)
    {
        if (!IsConnected) throw new Exceptions.DaytonaException("PTY is not connected");
        await _ws.SendAsync(data, WebSocketMessageType.Binary, true, ct);
    }

    public async Task<PtyResult> WaitAsync(int timeoutSeconds = 30, CancellationToken ct = default)
    {
        using var cts = CancellationTokenSource.CreateLinkedTokenSource(ct);
        cts.CancelAfter(TimeSpan.FromSeconds(timeoutSeconds));
        try
        {
            return await _exitTcs.Task.WaitAsync(cts.Token);
        }
        catch (OperationCanceledException)
        {
            return new PtyResult { ExitCode = -1, Error = "Timed out waiting for PTY exit" };
        }
    }

    public async Task ResizeAsync(int cols, int rows, CancellationToken ct = default)
    {
        await _resizeCallback(cols, rows);
    }

    public async Task KillAsync(CancellationToken ct = default)
    {
        await _killCallback();
    }

    public async ValueTask DisposeAsync()
    {
        _receiveCts.Cancel();
        try
        {
            if (_ws.State == WebSocketState.Open)
                await _ws.CloseAsync(WebSocketCloseStatus.NormalClosure, "", CancellationToken.None);
        }
        catch { }
        _ws.Dispose();
        _receiveCts.Dispose();
    }

    private async Task ReceiveLoopAsync()
    {
        var buffer = new byte[8192];
        try
        {
            while (_ws.State == WebSocketState.Open && !_receiveCts.IsCancellationRequested)
            {
                WebSocketReceiveResult result;
                using var ms = new MemoryStream();
                do
                {
                    result = await _ws.ReceiveAsync(buffer, _receiveCts.Token);
                    ms.Write(buffer, 0, result.Count);
                } while (!result.EndOfMessage);

                var data = ms.ToArray();

                if (result.MessageType == WebSocketMessageType.Text)
                {
                    var text = Encoding.UTF8.GetString(data);
                    if (TryParseControl(text)) continue;
                    if (_onData != null) await _onData(data);
                }
                else if (result.MessageType == WebSocketMessageType.Binary)
                {
                    if (_onData != null) await _onData(data);
                }
                else if (result.MessageType == WebSocketMessageType.Close)
                {
                    HandleClose(result.CloseStatusDescription);
                    break;
                }
            }
        }
        catch (OperationCanceledException) { }
        catch (WebSocketException) { }
        finally
        {
            _connected = false;
            if (!_exitTcs.Task.IsCompleted)
                _exitTcs.TrySetResult(new PtyResult { ExitCode = ExitCode ?? -1, Error = Error });
            if (!_connectionTcs.Task.IsCompleted)
                _connectionTcs.TrySetResult(false);
        }
    }

    private bool TryParseControl(string text)
    {
        try
        {
            using var doc = JsonDocument.Parse(text);
            var root = doc.RootElement;
            if (root.TryGetProperty("type", out var typeEl) && typeEl.GetString() == "control")
            {
                if (root.TryGetProperty("status", out var statusEl))
                {
                    var status = statusEl.GetString();
                    if (status == "connected")
                    {
                        _connectionTcs.TrySetResult(true);
                        return true;
                    }
                    if (status == "error")
                    {
                        Error = root.TryGetProperty("error", out var errEl) ? errEl.GetString() : "Unknown error";
                        _connectionTcs.TrySetResult(false);
                        return true;
                    }
                }
                return true;
            }
        }
        catch (JsonException) { }
        return false;
    }

    private void HandleClose(string? reason)
    {
        _connected = false;
        if (!string.IsNullOrEmpty(reason))
        {
            try
            {
                using var doc = JsonDocument.Parse(reason);
                var root = doc.RootElement;
                if (root.TryGetProperty("exitCode", out var codeEl))
                    ExitCode = codeEl.GetInt32();
                if (root.TryGetProperty("exitReason", out var reasonEl))
                    Error = reasonEl.GetString();
                if (root.TryGetProperty("error", out var errEl))
                    Error = errEl.GetString();
            }
            catch (JsonException)
            {
                ExitCode = 0;
            }
        }
        else
        {
            ExitCode ??= 0;
        }

        _exitTcs.TrySetResult(new PtyResult { ExitCode = ExitCode ?? 0, Error = Error });
    }
}
