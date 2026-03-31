// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

using System.Text.Json.Serialization;

namespace Daytona.Sdk.Models;

public record GitCommitResponse
{
    [JsonPropertyName("hash")] public string Hash { get; init; } = string.Empty;
}