// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

using System.Text.Json.Serialization;

namespace Daytona.Sdk.Models;

public record Session
{
    [JsonPropertyName("sessionId")] public string SessionId { get; init; } = string.Empty;
    [JsonPropertyName("commands")] public List<Command> Commands { get; init; } = new();
}