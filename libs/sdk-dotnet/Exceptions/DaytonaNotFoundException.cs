// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

namespace Daytona.Sdk.Exceptions;

public class DaytonaNotFoundException : DaytonaException
{
    public DaytonaNotFoundException(string message)
        : base(message, 404)
    {
    }
}