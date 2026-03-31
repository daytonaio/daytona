// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.model;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import com.fasterxml.jackson.annotation.JsonProperty;

@JsonIgnoreProperties(ignoreUnknown = true)
public class GitCommitResponse {
    @JsonProperty("hash")
    private String hash;

    public String getHash() { return hash; }
    public void setHash(String hash) { this.hash = hash; }
}