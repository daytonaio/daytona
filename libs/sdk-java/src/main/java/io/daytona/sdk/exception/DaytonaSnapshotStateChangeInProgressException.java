// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.exception;
/**
 * The snapshot is changing state; retry shortly.
 *
 * <p>Subclass of {@link DaytonaBadRequestException}.
 */
public class DaytonaSnapshotStateChangeInProgressException extends DaytonaBadRequestException {
    public DaytonaSnapshotStateChangeInProgressException(String message) {
        super(message);
    }

    public DaytonaSnapshotStateChangeInProgressException(String message, Throwable cause) {
        super(message, cause);
    }

    public DaytonaSnapshotStateChangeInProgressException(String message, String code, String source) {
        super(message, code, source);
    }

    public DaytonaSnapshotStateChangeInProgressException(String message, Throwable cause, String code, String source) {
        super(message, cause, code, source);
    }
}
