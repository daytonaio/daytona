// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

namespace Daytona.Sdk.Exceptions;

public class DaytonaException : Exception
{
    public int? StatusCode { get; }

    public DaytonaException(string message, int? statusCode = null, Exception? innerException = null)
        : base(message, innerException)
    {
        StatusCode = statusCode;
    }
}