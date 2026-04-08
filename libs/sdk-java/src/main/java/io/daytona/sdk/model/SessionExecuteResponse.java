// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.model;

public class SessionExecuteResponse extends io.daytona.toolbox.client.model.SessionExecuteResponse {
    public SessionExecuteResponse() {}

    public SessionExecuteResponse(io.daytona.toolbox.client.model.SessionExecuteResponse source) {
        super();
        if (source != null) {
            setCmdId(source.getCmdId());
            setOutput(source.getOutput());
            setExitCode(source.getExitCode());
        }
    }
}
