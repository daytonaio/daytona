// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.model;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import com.fasterxml.jackson.annotation.JsonProperty;

import java.util.ArrayList;
import java.util.List;

@JsonIgnoreProperties(ignoreUnknown = true)
public class Session {
    @JsonProperty("sessionId")
    private String sessionId;
    @JsonProperty("commands")
    private List<Command> commands;

    public String getSessionId() { return sessionId; }
    public void setSessionId(String sessionId) { this.sessionId = sessionId; }
    public List<Command> getCommands() { return commands == null ? new ArrayList<Command>() : commands; }
    public void setCommands(List<Command> commands) { this.commands = commands; }
}