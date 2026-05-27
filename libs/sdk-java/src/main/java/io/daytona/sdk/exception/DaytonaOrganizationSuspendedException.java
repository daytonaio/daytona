// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.exception;
/**
 * The organization is suspended; operation is not allowed.
 *
 * <p>Subclass of {@link DaytonaForbiddenException}.
 */
public class DaytonaOrganizationSuspendedException extends DaytonaForbiddenException {
    public DaytonaOrganizationSuspendedException(String message) {
        super(message);
    }

    public DaytonaOrganizationSuspendedException(String message, Throwable cause) {
        super(message, cause);
    }

    public DaytonaOrganizationSuspendedException(String message, String code, String source) {
        super(message, code, source);
    }

    public DaytonaOrganizationSuspendedException(String message, Throwable cause, String code, String source) {
        super(message, cause, code, source);
    }
}
