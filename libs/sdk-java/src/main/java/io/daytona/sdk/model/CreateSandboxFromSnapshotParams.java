// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.model;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;

@JsonIgnoreProperties(ignoreUnknown = true)
/**
 * Parameters for creating a Sandbox from an existing snapshot.
 */
public class CreateSandboxFromSnapshotParams extends CreateSandboxParams {
    private String snapshot;

    /**
     * Returns snapshot name used for Sandbox creation.
     *
     * @return snapshot name
     */
    public String getSnapshot() { return snapshot; }

    /**
     * Sets snapshot name used for Sandbox creation.
     *
     * @param snapshot snapshot name
     */
    public void setSnapshot(String snapshot) { this.snapshot = snapshot; }
}
