// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.model;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import com.fasterxml.jackson.annotation.JsonProperty;

@JsonIgnoreProperties(ignoreUnknown = true)
public class SessionExecuteRequest {
    @JsonProperty("command")
    private String command;

    @JsonProperty("runAsync")
    private Boolean runAsync;

    public SessionExecuteRequest() {}

    public SessionExecuteRequest(String command, Boolean runAsync) {
        this.command = command;
        this.runAsync = runAsync;
    }

    public String getCommand() { return command; }
    public void setCommand(String command) { this.command = command; }
    public Boolean getRunAsync() { return runAsync; }
    public void setRunAsync(Boolean runAsync) { this.runAsync = runAsync; }
}