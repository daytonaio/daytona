// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.model;

public class Session extends io.daytona.toolbox.client.model.Session {
    public Session() {}

    public Session(io.daytona.toolbox.client.model.Session source) {
        super();
        if (source != null) {
            setSessionId(source.getSessionId());
            setCommands(source.getCommands());
        }
    }
}
