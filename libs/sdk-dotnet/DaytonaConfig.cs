// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

namespace Daytona.Sdk;

public class DaytonaConfig
{
    public string? ApiKey { get; set; }
    public string ApiUrl { get; set; } = "https://app.daytona.io/api";
    public string? Target { get; set; }
}