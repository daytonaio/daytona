// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.model;

import io.daytona.toolbox.client.model.CodeRunResponse;
import io.daytona.toolbox.client.model.CodeRunArtifacts;

public class ExecuteResponse extends io.daytona.toolbox.client.model.ExecuteResponse {
    private CodeRunArtifacts artifacts;

    public ExecuteResponse() {}

    public ExecuteResponse(io.daytona.toolbox.client.model.ExecuteResponse source) {
        super();
        if (source != null) {
            setExitCode(source.getExitCode());
            setResult(source.getResult());
        }
    }

    public ExecuteResponse(CodeRunResponse source) {
        super();
        if (source != null) {
            setExitCode(source.getExitCode());
            setResult(source.getResult());
            setArtifacts(source.getArtifacts());
        }
    }

    public CodeRunArtifacts getArtifacts() {
        return artifacts;
    }

    public void setArtifacts(CodeRunArtifacts artifacts) {
        this.artifacts = artifacts;
    }
}
