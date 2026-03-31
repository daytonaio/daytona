// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk;

import java.util.Map;
import java.util.StringJoiner;

public class Image {
    private final StringBuilder dockerfile = new StringBuilder();

    private Image() {}

    public static Image base(String baseImage) {
        Image image = new Image();
        image.dockerfile.append("FROM ").append(baseImage).append("\n");
        return image;
    }

    public static Image debianSlim(String pythonVersion) {
        String version = pythonVersion == null || pythonVersion.isEmpty() ? "3.11" : pythonVersion;
        return base("python:" + version + "-slim");
    }

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

    public Image runCommands(String... commands) {
        if (commands == null || commands.length == 0) {
            return this;
        }
        for (String cmd : commands) {
            dockerfile.append("RUN ").append(cmd).append("\n");
        }
        return this;
    }

    public Image env(Map<String, String> envVars) {
        if (envVars == null) {
            return this;
        }
        for (Map.Entry<String, String> e : envVars.entrySet()) {
            dockerfile.append("ENV ").append(e.getKey()).append("=").append(e.getValue()).append("\n");
        }
        return this;
    }

    public Image workdir(String path) {
        dockerfile.append("WORKDIR ").append(path).append("\n");
        return this;
    }

    public Image entrypoint(String... commands) {
        dockerfile.append("ENTRYPOINT ").append(jsonArray(commands)).append("\n");
        return this;
    }

    public Image cmd(String... commands) {
        dockerfile.append("CMD ").append(jsonArray(commands)).append("\n");
        return this;
    }

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