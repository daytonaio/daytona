// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.examples;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import io.daytona.sdk.CodeInterpreter;
import io.daytona.sdk.Daytona;
import io.daytona.sdk.Image;
import io.daytona.sdk.RunCodeOptions;
import io.daytona.sdk.Sandbox;
import io.daytona.sdk.model.CreateSandboxFromImageParams;
import io.daytona.sdk.model.ExecuteResponse;

import java.io.FileOutputStream;
import java.io.IOException;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.Base64;

public class Charts {
    public static void main(String[] args) {
        try (Daytona daytona = new Daytona()) {
            CreateSandboxFromImageParams params = new CreateSandboxFromImageParams();
            params.setImage(Image.debianSlim("3.13").pipInstall("matplotlib", "numpy"));
            params.setLanguage("python");

            System.out.println("Creating Python sandbox with matplotlib");
            Sandbox sandbox = daytona.create(params, 300);

            try {
                System.out.println("\n=== CodeInterpreter.runCode (with streaming callbacks) ===");
                CodeInterpreter.ExecutionResult result = sandbox.codeInterpreter.runCode(CODE,
                        new RunCodeOptions()
                                .setOnStdout(chunk -> System.out.print("[stdout] " + chunk))
                                .setOnStderr(chunk -> System.out.print("[stderr] " + chunk))
                );
                if (result.getError() != null) {
                    System.out.println("Error: " + result.getError().getValue());
                }

                System.out.println("\n\n=== Process.codeRun ===");
                ExecuteResponse processResult = sandbox.process.codeRun(CODE);
                System.out.println("Exit code: " + processResult.getExitCode());

                Path outputDir = Paths.get(System.getProperty("user.dir"));
                ObjectMapper mapper = new ObjectMapper();
                int chartIndex = 0;
                for (String line : processResult.getResult().split("\n")) {
                    if (line.startsWith("dtn_artifact_k39fd2:")) {
                        String json = line.substring("dtn_artifact_k39fd2:".length());
                        JsonNode artifact = mapper.readTree(json);
                        JsonNode value = artifact.get("value");
                        String title = value.has("title") ? value.get("title").asText("chart") : "chart";
                        String png = value.has("png") ? value.get("png").asText() : null;
                        if (png != null && !png.isEmpty()) {
                            String filename = title.replaceAll("[^a-zA-Z0-9_-]", "_") + ".png";
                            Path dest = outputDir.resolve(filename);
                            try (FileOutputStream fos = new FileOutputStream(dest.toFile())) {
                                fos.write(Base64.getDecoder().decode(png));
                            }
                            System.out.println("Saved chart: " + dest);
                            chartIndex++;
                        }
                    }
                }
                System.out.println("Total charts saved: " + chartIndex);
            } finally {
                System.out.println("Deleting sandbox");
                sandbox.delete();
            }
        } catch (Exception e) {
            System.err.println("Error: " + e.getMessage());
        }
    }

    private static final String CODE = String.join("\n",
            "import matplotlib.pyplot as plt",
            "import numpy as np",
            "",
            "x = np.linspace(0, 10, 30)",
            "y = np.sin(x)",
            "",
            "plt.figure(figsize=(8, 5))",
            "plt.plot(x, y, 'b-', linewidth=2)",
            "plt.title('Line Chart')",
            "plt.xlabel('X-axis')",
            "plt.ylabel('Y-axis')",
            "plt.grid(True)",
            "plt.show()",
            "",
            "plt.figure(figsize=(8, 5))",
            "plt.scatter(x, y, c=y, cmap='viridis', s=100*np.abs(y))",
            "plt.colorbar(label='Value')",
            "plt.title('Scatter Plot')",
            "plt.xlabel('X-axis')",
            "plt.ylabel('Y-axis')",
            "plt.show()",
            "",
            "categories = ['A', 'B', 'C', 'D', 'E']",
            "values = [40, 63, 15, 25, 8]",
            "plt.figure(figsize=(10, 6))",
            "plt.bar(categories, values, color='skyblue', edgecolor='navy')",
            "plt.title('Bar Chart')",
            "plt.xlabel('Categories')",
            "plt.ylabel('Values')",
            "plt.show()"
    );
}
