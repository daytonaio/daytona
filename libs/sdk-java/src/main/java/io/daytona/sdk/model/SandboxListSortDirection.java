// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.model;

/**
 * Sort direction for {@link io.daytona.sdk.Daytona#list}.
 *
 * <p>This is the SDK-facing mirror of the api-client enum. It exists so users
 * never need to import from the api-client package.
 */
public enum SandboxListSortDirection {
    ASC("asc"),
    DESC("desc");

    private final String value;

    SandboxListSortDirection(String value) {
        this.value = value;
    }

    public String getValue() {
        return value;
    }

    @Override
    public String toString() {
        return value;
    }

    /** Convert to the api-client enum (used internally by the SDK). */
    public io.daytona.api.client.model.SandboxListSortDirection toApiClient() {
        return io.daytona.api.client.model.SandboxListSortDirection.fromValue(value);
    }
}
