// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.model;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import com.fasterxml.jackson.annotation.JsonProperty;

@JsonIgnoreProperties(ignoreUnknown = true)
/**
 * Request payload for executing a command in an existing session.
 */
public class SessionExecuteRequest {
    @JsonProperty("command")
    private String command;

    @JsonProperty("runAsync")
    private Boolean runAsync;

    /**
     * Creates an empty session execution request.
     */
    public SessionExecuteRequest() {}

    /**
     * Creates a session execution request with explicit values.
     *
     * @param command command to execute
     * @param runAsync whether to execute asynchronously
     */
    public SessionExecuteRequest(String command, Boolean runAsync) {
        this.command = command;
        this.runAsync = runAsync;
    }

    /**
     * Returns command text.
     *
     * @return command to execute
     */
    public String getCommand() { return command; }

    /**
     * Sets command text.
     *
     * @param command command to execute
     */
    public void setCommand(String command) { this.command = command; }

    /**
     * Returns asynchronous execution flag.
     *
     * @return {@code true} to execute asynchronously
     */
    public Boolean getRunAsync() { return runAsync; }

    /**
     * Sets asynchronous execution flag.
     *
     * @param runAsync {@code true} to execute asynchronously
     */
    public void setRunAsync(Boolean runAsync) { this.runAsync = runAsync; }
}
