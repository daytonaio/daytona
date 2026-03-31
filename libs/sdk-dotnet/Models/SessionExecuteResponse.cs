// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

using System.Text.Json.Serialization;

namespace Daytona.Sdk.Models;

public record SessionExecuteResponse
{
    [JsonPropertyName("cmdId")] public string? CmdId { get; init; }
    [JsonPropertyName("output")] public string? Output { get; init; }
    [JsonPropertyName("exitCode")] public int? ExitCode { get; init; }

    public string? Stdout { get; init; }
    public string? Stderr { get; init; }
}