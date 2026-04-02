// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.model;

import com.fasterxml.jackson.annotation.JsonAlias;
import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import com.fasterxml.jackson.annotation.JsonProperty;

@JsonIgnoreProperties(ignoreUnknown = true)
/**
 * Result of command or code execution in a Sandbox.
 */
public class ExecuteResponse {
    @JsonProperty("exitCode")
    @JsonAlias("code")
    private Integer exitCode;

    @JsonProperty("result")
    private String result;

    /**
     * Returns process exit code.
     *
     * @return exit code, or {@code 0} when not present
     */
    public int getExitCode() { return exitCode == null ? 0 : exitCode; }

    /**
     * Sets process exit code.
     *
     * @param exitCode exit code
     */
    public void setExitCode(Integer exitCode) { this.exitCode = exitCode; }

    /**
     * Returns command standard output.
     *
     * @return execution output
     */
    public String getResult() { return result == null ? "" : result; }

    /**
     * Sets command standard output.
     *
     * @param result execution output
     */
    public void setResult(String result) { this.result = result; }
}
