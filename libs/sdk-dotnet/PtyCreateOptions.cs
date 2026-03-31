// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

namespace Daytona.Sdk;

public class PtyCreateOptions
{
    public string Id { get; set; } = Guid.NewGuid().ToString();
    public int Cols { get; set; } = 120;
    public int Rows { get; set; } = 30;
    public Func<byte[], Task>? OnData { get; set; }
}
