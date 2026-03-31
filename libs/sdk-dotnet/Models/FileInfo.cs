// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

using System.Text.Json.Serialization;

namespace Daytona.Sdk.Models;

public record FileInfoModel
{
    [JsonPropertyName("name")] public string Name { get; init; } = string.Empty;
    [JsonPropertyName("size")] public long Size { get; init; }
    [JsonPropertyName("mode")] public string? Mode { get; init; }
    [JsonPropertyName("modTime")] public string? ModTime { get; init; }
    [JsonPropertyName("modifiedTime")] public string? ModifiedTime { get; init; }
    [JsonPropertyName("isDirectory")] public bool? IsDirectory { get; init; }
    [JsonPropertyName("isDir")] public bool? IsDir { get; init; }
}

public record MatchResult
{
    [JsonPropertyName("file")] public string File { get; init; } = string.Empty;
    [JsonPropertyName("line")] public int Line { get; init; }
    [JsonPropertyName("content")] public string? Content { get; init; }
}

public record SearchResult
{
    [JsonPropertyName("matches")] public List<MatchResult> Matches { get; init; } = new();
    [JsonPropertyName("count")] public int? Count { get; init; }
}