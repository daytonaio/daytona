// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

using System.Text.Json.Serialization;

namespace Daytona.Sdk.Models;

public record Command
{
    [JsonPropertyName("id")] public string Id { get; init; } = string.Empty;
    [JsonPropertyName("command")] public string CommandText { get; init; } = string.Empty;
    [JsonPropertyName("exitCode")] public int? ExitCode { get; init; }
    [JsonPropertyName("output")] public string? Output { get; init; }
}