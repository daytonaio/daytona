// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

using System.Text.Json.Serialization;

namespace Daytona.Sdk.Models;

public record GitStatus
{
    [JsonPropertyName("currentBranch")] public string? CurrentBranch { get; init; }
    [JsonPropertyName("ahead")] public int? Ahead { get; init; }
    [JsonPropertyName("behind")] public int? Behind { get; init; }
    [JsonPropertyName("branchPublished")] public bool? BranchPublished { get; init; }
    [JsonPropertyName("fileStatus")] public List<GitFileStatus> FileStatus { get; init; } = new();
}

public record GitFileStatus
{
    [JsonPropertyName("path")] public string Path { get; init; } = string.Empty;
    [JsonPropertyName("status")] public string Status { get; init; } = string.Empty;
}

public record BranchesResponse
{
    [JsonPropertyName("branches")] public List<string> Branches { get; init; } = new();
}