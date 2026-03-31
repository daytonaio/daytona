// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

namespace Daytona.Sdk.Exceptions;

public class DaytonaTimeoutException : DaytonaException
{
    public DaytonaTimeoutException(string message, Exception? innerException = null)
        : base(message, null, innerException)
    {
    }
}