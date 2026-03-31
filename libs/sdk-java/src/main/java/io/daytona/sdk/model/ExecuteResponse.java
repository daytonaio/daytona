// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.model;

import com.fasterxml.jackson.annotation.JsonAlias;
import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import com.fasterxml.jackson.annotation.JsonProperty;

@JsonIgnoreProperties(ignoreUnknown = true)
public class ExecuteResponse {
    @JsonProperty("exitCode")
    @JsonAlias("code")
    private Integer exitCode;

    @JsonProperty("result")
    private String result;

    public int getExitCode() { return exitCode == null ? 0 : exitCode; }
    public void setExitCode(Integer exitCode) { this.exitCode = exitCode; }
    public String getResult() { return result == null ? "" : result; }
    public void setResult(String result) { this.result = result; }
}