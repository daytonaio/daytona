// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.model;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import com.fasterxml.jackson.annotation.JsonProperty;

@JsonIgnoreProperties(ignoreUnknown = true)
/**
 * Response payload for session command execution.
 */
public class SessionExecuteResponse {
    @JsonProperty("cmdId")
    private String cmdId;
    @JsonProperty("output")
    private String output;
    @JsonProperty("exitCode")
    private Integer exitCode;

    /**
     * Returns executed command identifier.
     *
     * @return command ID
     */
    public String getCmdId() { return cmdId; }

    /**
     * Sets executed command identifier.
     *
     * @param cmdId command ID
     */
    public void setCmdId(String cmdId) { this.cmdId = cmdId; }

    /**
     * Returns command output.
     *
     * @return command output text
     */
    public String getOutput() { return output == null ? "" : output; }

    /**
     * Sets command output.
     *
     * @param output command output text
     */
    public void setOutput(String output) { this.output = output; }

    /**
     * Returns command exit code.
     *
     * @return exit code, or {@code 0} when not present
     */
    public int getExitCode() { return exitCode == null ? 0 : exitCode; }

    /**
     * Sets command exit code.
     *
     * @param exitCode command exit code
     */
    public void setExitCode(Integer exitCode) { this.exitCode = exitCode; }
}
