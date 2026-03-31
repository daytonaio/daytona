// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

using System.Text.Json.Serialization;

namespace Daytona.Sdk.Models;

public record Volume
{
    [JsonPropertyName("id")] public string Id { get; init; } = string.Empty;
    [JsonPropertyName("name")] public string Name { get; init; } = string.Empty;
    [JsonPropertyName("state")] public string? State { get; init; }
    [JsonPropertyName("createdAt")] public DateTimeOffset? CreatedAt { get; init; }
    [JsonPropertyName("updatedAt")] public DateTimeOffset? UpdatedAt { get; init; }
}