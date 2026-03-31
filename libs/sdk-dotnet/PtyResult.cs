// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

namespace Daytona.Sdk;

public class PtyResult
{
    public int ExitCode { get; init; }
    public string? Error { get; init; }
}
