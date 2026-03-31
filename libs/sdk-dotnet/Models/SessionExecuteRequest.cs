// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

using System.Text.Json.Serialization;

namespace Daytona.Sdk.Models;

public record SessionExecuteRequest
{
    [JsonPropertyName("command")] public string Command { get; init; } = string.Empty;
    [JsonPropertyName("runAsync")] public bool? RunAsync { get; init; }
    [JsonPropertyName("suppressInputEcho")] public bool? SuppressInputEcho { get; init; }
}