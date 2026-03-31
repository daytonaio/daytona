// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.exception;

public class DaytonaNotFoundException extends DaytonaException {
    public DaytonaNotFoundException(String message) {
        super(404, message);
    }
}