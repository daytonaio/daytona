// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.model;

public class ExecuteResponse extends io.daytona.toolbox.client.model.ExecuteResponse {
    public ExecuteResponse() {}

    public ExecuteResponse(io.daytona.toolbox.client.model.ExecuteResponse source) {
        super();
        if (source != null) {
            setExitCode(source.getExitCode());
            setResult(source.getResult());
        }
    }
}
