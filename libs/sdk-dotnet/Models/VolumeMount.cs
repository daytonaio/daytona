// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

using System.Text.Json.Serialization;

namespace Daytona.Sdk.Models;

public record VolumeMount(
    [property: JsonPropertyName("volumeId")] string VolumeId,
    [property: JsonPropertyName("mountPath")] string MountPath,
    [property: JsonPropertyName("subPath")] string? SubPath = null
);