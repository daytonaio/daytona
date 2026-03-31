// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk;

public enum CodeLanguage {
    PYTHON("python"),
    TYPESCRIPT("typescript"),
    JAVASCRIPT("javascript");

    private final String value;

    CodeLanguage(String value) {
        this.value = value;
    }

    public String getValue() {
        return value;
    }
}