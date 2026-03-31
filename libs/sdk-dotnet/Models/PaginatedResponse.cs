// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

using System.Text.Json.Serialization;

namespace Daytona.Sdk.Models;

public record PaginatedResponse<T>
{
    [JsonPropertyName("items")] public List<T> Items { get; init; } = new();
    [JsonPropertyName("total")] public int Total { get; init; }
    [JsonPropertyName("page")] public int Page { get; init; }
    [JsonPropertyName("totalPages")] public int TotalPages { get; init; }
}