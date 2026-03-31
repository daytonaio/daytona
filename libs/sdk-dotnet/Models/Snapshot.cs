// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

using System.Text.Json.Serialization;

namespace Daytona.Sdk.Models;

public record Snapshot
{
    [JsonPropertyName("id")] public string Id { get; init; } = string.Empty;
    [JsonPropertyName("name")] public string Name { get; init; } = string.Empty;
    [JsonPropertyName("state")] public string? State { get; init; }
    [JsonPropertyName("imageName")] public string? ImageName { get; init; }
    [JsonPropertyName("cpu")] public int? Cpu { get; init; }
    [JsonPropertyName("gpu")] public int? Gpu { get; init; }
    [JsonPropertyName("memory")] public int? Memory { get; init; }
    [JsonPropertyName("mem")] public int? Mem { get; init; }
    [JsonPropertyName("disk")] public int? Disk { get; init; }
    [JsonPropertyName("createdAt")] public DateTimeOffset? CreatedAt { get; init; }
    [JsonPropertyName("updatedAt")] public DateTimeOffset? UpdatedAt { get; init; }
}