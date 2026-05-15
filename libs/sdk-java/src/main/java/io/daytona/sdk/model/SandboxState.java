// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.model;

/**
 * Lifecycle state of a Sandbox.
 *
 * <p>This is the SDK-facing mirror of the api-client enum. It exists so users
 * never need to import from the api-client package.
 */
public enum SandboxState {
    CREATING("creating"),
    RESTORING("restoring"),
    DESTROYED("destroyed"),
    DESTROYING("destroying"),
    STARTED("started"),
    STOPPED("stopped"),
    STARTING("starting"),
    STOPPING("stopping"),
    ERROR("error"),
    BUILD_FAILED("build_failed"),
    PENDING_BUILD("pending_build"),
    BUILDING_SNAPSHOT("building_snapshot"),
    UNKNOWN("unknown"),
    PULLING_SNAPSHOT("pulling_snapshot"),
    ARCHIVED("archived"),
    ARCHIVING("archiving"),
    RESIZING("resizing"),
    SNAPSHOTTING("snapshotting"),
    FORKING("forking");

    private final String value;

    SandboxState(String value) {
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
    public io.daytona.api.client.model.SandboxState toApiClient() {
        return io.daytona.api.client.model.SandboxState.fromValue(value);
    }
}
