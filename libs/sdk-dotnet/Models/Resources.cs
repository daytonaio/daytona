// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

using System.Text.Json.Serialization;

namespace Daytona.Sdk.Models;

public record Resources(
    [property: JsonPropertyName("cpu")] int? Cpu = null,
    [property: JsonPropertyName("gpu")] int? Gpu = null,
    [property: JsonPropertyName("memory")] int? Memory = null,
    [property: JsonPropertyName("disk")] int? Disk = null
);

public record SandboxInfo
{
    [JsonPropertyName("id")] public string Id { get; init; } = string.Empty;
    [JsonPropertyName("name")] public string Name { get; init; } = string.Empty;
    [JsonPropertyName("state")] public string State { get; init; } = string.Empty;
    [JsonPropertyName("user")] public string? User { get; init; }
    [JsonPropertyName("target")] public string? Target { get; init; }
    [JsonPropertyName("cpu")] public int? Cpu { get; init; }
    [JsonPropertyName("gpu")] public int? Gpu { get; init; }
    [JsonPropertyName("memory")] public int? Memory { get; init; }
    [JsonPropertyName("disk")] public int? Disk { get; init; }
    [JsonPropertyName("env")] public Dictionary<string, string> Env { get; init; } = new();
    [JsonPropertyName("labels")] public Dictionary<string, string> Labels { get; init; } = new();
    [JsonPropertyName("public")] public bool Public { get; init; }
    [JsonPropertyName("toolboxProxyUrl")] public string ToolboxProxyUrl { get; init; } = string.Empty;
    [JsonPropertyName("autoStopInterval")] public int? AutoStopInterval { get; init; }
    [JsonPropertyName("autoArchiveInterval")] public int? AutoArchiveInterval { get; init; }
    [JsonPropertyName("autoDeleteInterval")] public int? AutoDeleteInterval { get; init; }
    [JsonPropertyName("networkBlockAll")] public bool? NetworkBlockAll { get; init; }
    [JsonPropertyName("networkAllowList")] public string? NetworkAllowList { get; init; }
}