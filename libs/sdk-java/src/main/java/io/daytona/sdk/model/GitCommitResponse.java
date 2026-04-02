// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.model;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import com.fasterxml.jackson.annotation.JsonProperty;

@JsonIgnoreProperties(ignoreUnknown = true)
/**
 * Result of a Git commit operation.
 */
public class GitCommitResponse {
    @JsonProperty("hash")
    private String hash;

    /**
     * Returns commit hash.
     *
     * @return commit SHA
     */
    public String getHash() { return hash; }

    /**
     * Sets commit hash.
     *
     * @param hash commit SHA
     */
    public void setHash(String hash) { this.hash = hash; }
}
