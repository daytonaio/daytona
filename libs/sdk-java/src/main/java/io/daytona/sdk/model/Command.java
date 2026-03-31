// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.model;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import com.fasterxml.jackson.annotation.JsonProperty;

@JsonIgnoreProperties(ignoreUnknown = true)
public class Command {
    @JsonProperty("id")
    private String id;
    @JsonProperty("command")
    private String command;
    @JsonProperty("exitCode")
    private Integer exitCode;

    public String getId() { return id; }
    public void setId(String id) { this.id = id; }
    public String getCommand() { return command; }
    public void setCommand(String command) { this.command = command; }
    public int getExitCode() { return exitCode == null ? 0 : exitCode; }
    public void setExitCode(Integer exitCode) { this.exitCode = exitCode; }
}