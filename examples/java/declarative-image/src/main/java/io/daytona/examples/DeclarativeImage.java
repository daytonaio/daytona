// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.examples;

import io.daytona.sdk.Daytona;
import io.daytona.sdk.Image;
import io.daytona.sdk.Sandbox;
import io.daytona.sdk.model.CreateSandboxFromImageParams;
import io.daytona.sdk.model.ExecuteResponse;

import java.util.HashMap;
import java.util.Map;

public class DeclarativeImage {
    public static void main(String[] args) {
        try (Daytona daytona = new Daytona()) {
            Map<String, String> envVars = new HashMap<>();
            envVars.put("MY_ENV_VAR", "test-value");

            Image image = Image.debianSlim("3.12")
                    .pipInstall("numpy", "pandas")
                    .runCommands("mkdir -p /home/daytona/workspace")
                    .workdir("/home/daytona/workspace")
                    .env(envVars);

            CreateSandboxFromImageParams params = new CreateSandboxFromImageParams();
            params.setImage(image);

            Sandbox sandbox = daytona.create(params, 300);
            try {
                System.out.println("Verifying sandbox from declarative image:");

                System.out.println("Python environment:");
                ExecuteResponse result = sandbox.process.executeCommand(
                        "python3 -c \"import numpy; print(numpy.__version__)\"");
                System.out.println(result.getResult().trim());

                ExecuteResponse envResult = sandbox.process.executeCommand("echo $MY_ENV_VAR");
                System.out.println("MY_ENV_VAR=" + envResult.getResult().trim());
            } finally {
                System.out.println("Deleting sandbox");
                sandbox.delete();
            }
        }
    }
}
