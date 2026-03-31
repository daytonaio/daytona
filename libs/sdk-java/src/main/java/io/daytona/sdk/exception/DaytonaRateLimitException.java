// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.exception;

public class DaytonaRateLimitException extends DaytonaException {
    public DaytonaRateLimitException(String message) {
        super(429, message);
    }
}