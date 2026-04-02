// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk;

/**
 * Supported programming languages for direct code execution in a Sandbox.
 *
 * <p>Python is used as the default language when no explicit language label is set on the
 * Sandbox.
 */
public enum CodeLanguage {
    PYTHON("python"),
    TYPESCRIPT("typescript"),
    JAVASCRIPT("javascript");

    private final String value;

    CodeLanguage(String value) {
        this.value = value;
    }

    /**
     * Returns the wire value used in Daytona labels and API payloads.
     *
     * @return lowercase language identifier
     */
    public String getValue() {
        return value;
    }
}
