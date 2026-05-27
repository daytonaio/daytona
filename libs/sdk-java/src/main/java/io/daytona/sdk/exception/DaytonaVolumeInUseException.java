// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.exception;
/**
 * The volume is currently attached and cannot be deleted.
 *
 * <p>Subclass of {@link DaytonaConflictException}.
 */
public class DaytonaVolumeInUseException extends DaytonaConflictException {
    public DaytonaVolumeInUseException(String message) {
        super(message);
    }

    public DaytonaVolumeInUseException(String message, Throwable cause) {
        super(message, cause);
    }

    public DaytonaVolumeInUseException(String message, String code, String source) {
        super(message, code, source);
    }

    public DaytonaVolumeInUseException(String message, Throwable cause, String code, String source) {
        super(message, cause, code, source);
    }
}
