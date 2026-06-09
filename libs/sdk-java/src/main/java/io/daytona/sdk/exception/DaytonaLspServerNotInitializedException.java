// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.exception;
/**
 * LSP server must be started via /lsp/start first.
 *
 * <p>Subclass of {@link DaytonaBadRequestException}.
 */
public class DaytonaLspServerNotInitializedException extends DaytonaBadRequestException {
    public DaytonaLspServerNotInitializedException(String message) {
        super(message);
    }

    public DaytonaLspServerNotInitializedException(String message, Throwable cause) {
        super(message, cause);
    }

    public DaytonaLspServerNotInitializedException(String message, String code, String source) {
        super(message, code, source);
    }

    public DaytonaLspServerNotInitializedException(String message, Throwable cause, String code, String source) {
        super(message, cause, code, source);
    }
}
