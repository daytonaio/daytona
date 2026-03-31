// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.model;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import com.fasterxml.jackson.annotation.JsonProperty;

@JsonIgnoreProperties(ignoreUnknown = true)
public class SessionExecuteResponse {
    @JsonProperty("cmdId")
    private String cmdId;
    @JsonProperty("output")
    private String output;
    @JsonProperty("exitCode")
    private Integer exitCode;

    public String getCmdId() { return cmdId; }
    public void setCmdId(String cmdId) { this.cmdId = cmdId; }
    public String getOutput() { return output == null ? "" : output; }
    public void setOutput(String output) { this.output = output; }
    public int getExitCode() { return exitCode == null ? 0 : exitCode; }
    public void setExitCode(Integer exitCode) { this.exitCode = exitCode; }
}