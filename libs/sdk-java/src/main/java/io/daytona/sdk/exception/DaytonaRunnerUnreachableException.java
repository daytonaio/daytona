// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.exception;
/**
 * The runner backend is unreachable.
 *
 * <p>Subclass of {@link DaytonaBadGatewayException}.
 */
public class DaytonaRunnerUnreachableException extends DaytonaBadGatewayException {
    public DaytonaRunnerUnreachableException(String message) {
        super(message);
    }

    public DaytonaRunnerUnreachableException(String message, Throwable cause) {
        super(message, cause);
    }

    public DaytonaRunnerUnreachableException(String message, String code, String source) {
        super(message, code, source);
    }

    public DaytonaRunnerUnreachableException(String message, Throwable cause, String code, String source) {
        super(message, cause, code, source);
    }
}
