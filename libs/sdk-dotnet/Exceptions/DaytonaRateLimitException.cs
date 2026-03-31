// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

namespace Daytona.Sdk.Exceptions;

public class DaytonaRateLimitException : DaytonaException
{
    public DaytonaRateLimitException(string message)
        : base(message, 429)
    {
    }
}