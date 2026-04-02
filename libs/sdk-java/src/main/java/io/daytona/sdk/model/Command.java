// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.model;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import com.fasterxml.jackson.annotation.JsonProperty;

@JsonIgnoreProperties(ignoreUnknown = true)
/**
 * Metadata for a command executed in a session.
 */
public class Command {
    @JsonProperty("id")
    private String id;
    @JsonProperty("command")
    private String command;
    @JsonProperty("exitCode")
    private Integer exitCode;

    /**
     * Returns command identifier.
     *
     * @return command ID
     */
    public String getId() { return id; }

    /**
     * Sets command identifier.
     *
     * @param id command ID
     */
    public void setId(String id) { this.id = id; }

    /**
     * Returns command text.
     *
     * @return command string
     */
    public String getCommand() { return command; }

    /**
     * Sets command text.
     *
     * @param command command string
     */
    public void setCommand(String command) { this.command = command; }

    /**
     * Returns command exit code.
     *
     * @return exit code, or {@code 0} when not present
     */
    public int getExitCode() { return exitCode == null ? 0 : exitCode; }

    /**
     * Sets command exit code.
     *
     * @param exitCode exit code
     */
    public void setExitCode(Integer exitCode) { this.exitCode = exitCode; }
}
