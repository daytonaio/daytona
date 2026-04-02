// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk;

import java.util.Map;
import java.util.StringJoiner;

/**
 * Declarative image builder used to define Sandbox runtime environments.
 *
 * <p>Use factory methods such as {@link #base(String)} or {@link #debianSlim(String)} and chain
 * mutating methods to append Dockerfile instructions.
 */
public class Image {
    private final StringBuilder dockerfile = new StringBuilder();

    private Image() {}

    /**
     * Creates an image definition from an existing base image.
     *
     * @param baseImage base image reference (for example {@code python:3.12-slim-bookworm})
     * @return new {@link Image} initialized with a {@code FROM} instruction
     */
    public static Image base(String baseImage) {
        Image image = new Image();
        image.dockerfile.append("FROM ").append(baseImage).append("\n");
        return image;
    }

    /**
     * Creates a Python Debian slim image.
     *
     * @param pythonVersion Python version to use; defaults to {@code 3.11} when {@code null} or empty
     * @return new {@link Image} using a Python slim base image
     */
    public static Image debianSlim(String pythonVersion) {
        String version = pythonVersion == null || pythonVersion.isEmpty() ? "3.11" : pythonVersion;
        return base("python:" + version + "-slim");
    }

    /**
     * Adds a {@code pip install} instruction for one or more packages.
     *
     * @param packages package names to install
     * @return this {@link Image} for method chaining
     */
    public Image pipInstall(String... packages) {
        if (packages == null || packages.length == 0) {
            return this;
        }
        StringJoiner joiner = new StringJoiner(" ");
        for (String pkg : packages) {
            joiner.add(pkg);
        }
        dockerfile.append("RUN pip install ").append(joiner.toString()).append("\n");
        return this;
    }

    /**
     * Adds one or more {@code RUN} instructions.
     *
     * @param commands shell commands to execute during image build
     * @return this {@link Image} for method chaining
     */
    public Image runCommands(String... commands) {
        if (commands == null || commands.length == 0) {
            return this;
        }
        for (String cmd : commands) {
            dockerfile.append("RUN ").append(cmd).append("\n");
        }
        return this;
    }

    /**
     * Adds environment variables using {@code ENV} instructions.
     *
     * @param envVars environment variables to set in the image
     * @return this {@link Image} for method chaining
     */
    public Image env(Map<String, String> envVars) {
        if (envVars == null) {
            return this;
        }
        for (Map.Entry<String, String> e : envVars.entrySet()) {
            dockerfile.append("ENV ").append(e.getKey()).append("=\"").append(e.getValue().replace("\"", "\\\"")).append("\"\n");
        }
        return this;
    }

    /**
     * Sets the default working directory using a {@code WORKDIR} instruction.
     *
     * @param path working directory path
     * @return this {@link Image} for method chaining
     */
    public Image workdir(String path) {
        dockerfile.append("WORKDIR ").append(path).append("\n");
        return this;
    }

    /**
     * Sets the container entrypoint.
     *
     * @param commands entrypoint command and arguments
     * @return this {@link Image} for method chaining
     */
    public Image entrypoint(String... commands) {
        dockerfile.append("ENTRYPOINT ").append(jsonArray(commands)).append("\n");
        return this;
    }

    /**
     * Sets the default container command.
     *
     * @param commands default command and arguments
     * @return this {@link Image} for method chaining
     */
    public Image cmd(String... commands) {
        dockerfile.append("CMD ").append(jsonArray(commands)).append("\n");
        return this;
    }

    /**
     * Returns generated Dockerfile content.
     *
     * @return Dockerfile text assembled by this builder
     */
    public String getDockerfile() {
        return dockerfile.toString();
    }

    private String jsonArray(String... values) {
        StringJoiner joiner = new StringJoiner(",", "[", "]");
        if (values != null) {
            for (String v : values) {
                joiner.add("\"" + v.replace("\"", "\\\"") + "\"");
            }
        }
        return joiner.toString();
    }
}
