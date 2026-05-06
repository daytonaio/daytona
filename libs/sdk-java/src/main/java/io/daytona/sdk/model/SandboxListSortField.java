// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.model;

/**
 * Field used to order results from {@link io.daytona.sdk.Daytona#list}.
 *
 * <p>This is the SDK-facing mirror of the api-client enum. It exists so users
 * never need to import from the api-client package.
 */
public enum SandboxListSortField {
    NAME("name"),
    CPU("cpu"),
    MEMORY_GIB("memoryGib"),
    DISK_GIB("diskGib"),
    LAST_ACTIVITY_AT("lastActivityAt"),
    CREATED_AT("createdAt");

    private final String value;

    SandboxListSortField(String value) {
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
    public io.daytona.api.client.model.SandboxListSortField toApiClient() {
        return io.daytona.api.client.model.SandboxListSortField.fromValue(value);
    }
}
