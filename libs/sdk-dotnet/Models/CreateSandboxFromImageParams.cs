// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

using System.Text.Json.Serialization;

namespace Daytona.Sdk.Models;

public record CreateSandboxFromImageParams
{
    [JsonPropertyName("name")] public string? Name { get; init; }
    [JsonPropertyName("user")] public string? User { get; init; }
    [JsonPropertyName("language")] public CodeLanguage? Language { get; init; }
    [JsonPropertyName("image")] public object Image { get; init; } = string.Empty;
    [JsonPropertyName("resources")] public Resources? Resources { get; init; }
    [JsonPropertyName("env")] public Dictionary<string, string>? Env { get; init; }
    [JsonPropertyName("labels")] public Dictionary<string, string>? Labels { get; init; }
    [JsonPropertyName("public")] public bool? Public { get; init; }
    [JsonPropertyName("autoStopInterval")] public int? AutoStopInterval { get; init; }
    [JsonPropertyName("autoArchiveInterval")] public int? AutoArchiveInterval { get; init; }
    [JsonPropertyName("autoDeleteInterval")] public int? AutoDeleteInterval { get; init; }
    [JsonPropertyName("volumes")] public List<VolumeMount>? Volumes { get; init; }
    [JsonPropertyName("networkBlockAll")] public bool? NetworkBlockAll { get; init; }
    [JsonPropertyName("networkAllowList")] public string? NetworkAllowList { get; init; }
}