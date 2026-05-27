// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.exception;
/**
 * The organization's quota was exceeded.
 *
 * <p>Subclass of {@link DaytonaBadRequestException}.
 */
public class DaytonaOrganizationQuotaExceededException extends DaytonaBadRequestException {
    public DaytonaOrganizationQuotaExceededException(String message) {
        super(message);
    }

    public DaytonaOrganizationQuotaExceededException(String message, Throwable cause) {
        super(message, cause);
    }

    public DaytonaOrganizationQuotaExceededException(String message, String code, String source) {
        super(message, code, source);
    }

    public DaytonaOrganizationQuotaExceededException(String message, Throwable cause, String code, String source) {
        super(message, cause, code, source);
    }
}
