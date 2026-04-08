// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.model;

public class Command extends io.daytona.toolbox.client.model.Command {
    public Command() {}

    public Command(io.daytona.toolbox.client.model.Command source) {
        super();
        if (source != null) {
            setId(source.getId());
            setCommand(source.getCommand());
            setExitCode(source.getExitCode());
        }
    }
}
