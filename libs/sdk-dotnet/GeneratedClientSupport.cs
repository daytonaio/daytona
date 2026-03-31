// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

using System.Reflection;
using System.Runtime.Serialization;
using System.Text.Json;
using Daytona.Sdk.Exceptions;
using Daytona.Sdk.Models;

namespace Daytona.Sdk;

internal static class GeneratedClientSupport
{
    private const string SourceHeader = "sdk-dotnet";
    private const string UserAgentHeader = "sdk-dotnet/0.1.0";

    internal static Daytona.ApiClient.Client.Configuration CreateMainConfiguration(string basePath, string apiKey)
    {
        var cfg = new Daytona.ApiClient.Client.Configuration
        {
            BasePath = NormalizeBasePath(basePath),
            AccessToken = apiKey
        };
        cfg.DefaultHeaders["X-Daytona-Source"] = SourceHeader;
        cfg.DefaultHeaders["User-Agent"] = UserAgentHeader;
        cfg.UserAgent = UserAgentHeader;
        return cfg;
    }

    internal static Daytona.ToolboxApiClient.Client.Configuration CreateToolboxConfiguration(string basePath, string apiKey)
    {
        var cfg = new Daytona.ToolboxApiClient.Client.Configuration
        {
            BasePath = NormalizeBasePath(basePath),
            AccessToken = apiKey
        };
        cfg.DefaultHeaders["Authorization"] = $"Bearer {apiKey}";
        cfg.DefaultHeaders["X-Daytona-Source"] = SourceHeader;
        cfg.DefaultHeaders["User-Agent"] = UserAgentHeader;
        cfg.UserAgent = UserAgentHeader;
        return cfg;
    }

    internal static string NormalizeBasePath(string basePath)
    {
        if (string.IsNullOrWhiteSpace(basePath))
        {
            throw new DaytonaException("API URL cannot be empty.");
        }

        return basePath.TrimEnd('/');
    }

    internal static async Task<T> ExecuteMainAsync<T>(Func<Task<T>> action, CancellationToken ct)
    {
        try
        {
            return await action();
        }
        catch (Daytona.ApiClient.Client.ApiException ex)
        {
            throw MapApiException(ex.ErrorCode, ex.ErrorContent, ex.Message, ex);
        }
        catch (OperationCanceledException ex) when (!ct.IsCancellationRequested)
        {
            throw new DaytonaTimeoutException("Request to Daytona API timed out.", ex);
        }
    }

    internal static async Task ExecuteMainAsync(Func<Task> action, CancellationToken ct)
    {
        try
        {
            await action();
        }
        catch (Daytona.ApiClient.Client.ApiException ex)
        {
            throw MapApiException(ex.ErrorCode, ex.ErrorContent, ex.Message, ex);
        }
        catch (OperationCanceledException ex) when (!ct.IsCancellationRequested)
        {
            throw new DaytonaTimeoutException("Request to Daytona API timed out.", ex);
        }
    }

    internal static async Task<T> ExecuteToolboxAsync<T>(Func<Task<T>> action, CancellationToken ct)
    {
        try
        {
            return await action();
        }
        catch (Daytona.ToolboxApiClient.Client.ApiException ex)
        {
            throw MapApiException(ex.ErrorCode, ex.ErrorContent, ex.Message, ex);
        }
        catch (OperationCanceledException ex) when (!ct.IsCancellationRequested)
        {
            throw new DaytonaTimeoutException("Request to Daytona API timed out.", ex);
        }
    }

    internal static async Task ExecuteToolboxAsync(Func<Task> action, CancellationToken ct)
    {
        try
        {
            await action();
        }
        catch (Daytona.ToolboxApiClient.Client.ApiException ex)
        {
            throw MapApiException(ex.ErrorCode, ex.ErrorContent, ex.Message, ex);
        }
        catch (OperationCanceledException ex) when (!ct.IsCancellationRequested)
        {
            throw new DaytonaTimeoutException("Request to Daytona API timed out.", ex);
        }
    }

    internal static SandboxInfo ToSdkSandboxInfo(Daytona.ApiClient.Model.Sandbox sandbox) => new()
    {
        Id = sandbox.Id ?? string.Empty,
        Name = sandbox.Name ?? string.Empty,
        State = EnumToWireValue(sandbox.State) ?? "unknown",
        User = sandbox.User,
        Target = sandbox.Target,
        Cpu = (int)sandbox.Cpu,
        Gpu = (int)sandbox.Gpu,
        Memory = (int)sandbox.Memory,
        Disk = (int)sandbox.Disk,
        Env = sandbox.Env ?? new Dictionary<string, string>(),
        Labels = sandbox.Labels ?? new Dictionary<string, string>(),
        Public = sandbox.Public,
        ToolboxProxyUrl = sandbox.ToolboxProxyUrl ?? string.Empty,
        AutoStopInterval = (int)sandbox.AutoStopInterval,
        AutoArchiveInterval = (int)sandbox.AutoArchiveInterval,
        AutoDeleteInterval = (int)sandbox.AutoDeleteInterval,
        NetworkBlockAll = sandbox.NetworkBlockAll,
        NetworkAllowList = sandbox.NetworkAllowList
    };

    internal static Snapshot ToSdkSnapshot(Daytona.ApiClient.Model.SnapshotDto snapshot) => new()
    {
        Id = snapshot.Id ?? string.Empty,
        Name = snapshot.Name ?? string.Empty,
        State = EnumToWireValue(snapshot.State),
        ImageName = snapshot.ImageName,
        Cpu = (int)snapshot.Cpu,
        Gpu = (int)snapshot.Gpu,
        Memory = (int)snapshot.Mem,
        Mem = (int)snapshot.Mem,
        Disk = (int)snapshot.Disk,
        CreatedAt = snapshot.CreatedAt,
        UpdatedAt = snapshot.UpdatedAt
    };

    internal static Volume ToSdkVolume(Daytona.ApiClient.Model.VolumeDto volume) => new()
    {
        Id = volume.Id ?? string.Empty,
        Name = volume.Name ?? string.Empty,
        State = EnumToWireValue(volume.State),
        CreatedAt = ParseDateTimeOffset(volume.CreatedAt),
        UpdatedAt = ParseDateTimeOffset(volume.UpdatedAt)
    };

    internal static ExecuteResponse ToSdkExecuteResponse(Daytona.ToolboxApiClient.Model.ExecuteResponse response) => new()
    {
        ExitCode = response.ExitCode,
        Result = response.Result ?? string.Empty
    };

    internal static Session ToSdkSession(Daytona.ToolboxApiClient.Model.Session session) => new()
    {
        SessionId = session.SessionId ?? string.Empty,
        Commands = session.Commands?.Select(ToSdkCommand).ToList() ?? new List<Command>()
    };

    internal static Command ToSdkCommand(Daytona.ToolboxApiClient.Model.Command command) => new()
    {
        Id = command.Id ?? string.Empty,
        CommandText = command.VarCommand ?? string.Empty,
        ExitCode = command.ExitCode
    };

    internal static SessionExecuteResponse ToSdkSessionExecuteResponse(Daytona.ToolboxApiClient.Model.SessionExecuteResponse response) => new()
    {
        CmdId = response.CmdId,
        Output = response.Output,
        ExitCode = response.ExitCode,
        Stdout = response.Stdout,
        Stderr = response.Stderr
    };

    internal static FileInfoModel ToSdkFileInfo(Daytona.ToolboxApiClient.Model.FileInfo fileInfo) => new()
    {
        Name = fileInfo.Name ?? string.Empty,
        Size = fileInfo.Size,
        Mode = fileInfo.Mode,
        ModTime = fileInfo.ModTime,
        ModifiedTime = fileInfo.ModTime,
        IsDirectory = fileInfo.IsDir,
        IsDir = fileInfo.IsDir
    };

    internal static MatchResult ToSdkMatch(Daytona.ToolboxApiClient.Model.Match match) => new()
    {
        File = match.File ?? string.Empty,
        Line = match.Line,
        Content = match.Content
    };

    internal static SearchResult ToSdkSearchResult(Daytona.ToolboxApiClient.Model.SearchFilesResponse response)
    {
        var matches = (response.Files ?? []).Select(path => new MatchResult
        {
            File = path,
            Line = 0,
            Content = null
        }).ToList();

        return new SearchResult
        {
            Matches = matches,
            Count = matches.Count
        };
    }

    internal static GitCommitResponse ToSdkGitCommitResponse(Daytona.ToolboxApiClient.Model.GitCommitResponse response) => new()
    {
        Hash = response.Hash ?? string.Empty
    };

    internal static BranchesResponse ToSdkBranchesResponse(Daytona.ToolboxApiClient.Model.ListBranchResponse response) => new()
    {
        Branches = response.Branches ?? new List<string>()
    };

    internal static GitStatus ToSdkGitStatus(Daytona.ToolboxApiClient.Model.GitStatus response) => new()
    {
        CurrentBranch = response.CurrentBranch,
        Ahead = response.Ahead,
        Behind = response.Behind,
        BranchPublished = response.BranchPublished,
        FileStatus = response.FileStatus?.Select(file => new GitFileStatus
        {
            Path = file.Name ?? string.Empty,
            Status = $"{EnumToWireValue(file.Staging)}/{EnumToWireValue(file.Worktree)}"
        }).ToList() ?? new List<GitFileStatus>()
    };

    private static DaytonaException MapApiException(int statusCode, object? errorContent, string fallbackMessage, Exception inner)
    {
        var message = ExtractErrorMessage(errorContent) ?? fallbackMessage;
        return statusCode switch
        {
            404 => new DaytonaNotFoundException(message),
            429 => new DaytonaRateLimitException(message),
            _ => new DaytonaException(message, statusCode, inner)
        };
    }

    private static string? ExtractErrorMessage(object? errorContent)
    {
        if (errorContent is null)
        {
            return null;
        }

        var text = errorContent.ToString();
        if (string.IsNullOrWhiteSpace(text))
        {
            return null;
        }

        try
        {
            using var doc = JsonDocument.Parse(text);
            if (doc.RootElement.ValueKind == JsonValueKind.Object)
            {
                if (doc.RootElement.TryGetProperty("message", out var messageProp))
                {
                    var message = messageProp.GetString();
                    if (!string.IsNullOrWhiteSpace(message))
                    {
                        return message;
                    }
                }

                if (doc.RootElement.TryGetProperty("error", out var errorProp))
                {
                    var error = errorProp.GetString();
                    if (!string.IsNullOrWhiteSpace(error))
                    {
                        return error;
                    }
                }
            }
        }
        catch
        {
        }

        return text;
    }

    private static DateTimeOffset? ParseDateTimeOffset(string? value)
    {
        if (string.IsNullOrWhiteSpace(value))
        {
            return null;
        }

        return DateTimeOffset.TryParse(value, out var parsed) ? parsed : null;
    }

    private static string? EnumToWireValue<TEnum>(TEnum enumValue) where TEnum : struct, Enum
    {
        var enumType = typeof(TEnum);
        var memberName = enumValue.ToString();
        var member = enumType.GetMember(memberName).FirstOrDefault();
        var enumMemberAttribute = member?.GetCustomAttribute<EnumMemberAttribute>();
        return enumMemberAttribute?.Value ?? memberName;
    }

    private static string? EnumToWireValue<TEnum>(TEnum? enumValue) where TEnum : struct, Enum
        => enumValue is null ? null : EnumToWireValue(enumValue.Value);
}