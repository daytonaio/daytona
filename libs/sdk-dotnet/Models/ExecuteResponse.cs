// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

using System.Text.Json.Serialization;

namespace Daytona.Sdk.Models;

public record ExecuteResponse
{
    [JsonPropertyName("exitCode")] public int ExitCode { get; init; }
    [JsonPropertyName("result")] public string Result { get; init; } = string.Empty;
}