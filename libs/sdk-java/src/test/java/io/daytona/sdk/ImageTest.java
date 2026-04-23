// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk;

import org.junit.jupiter.api.Test;

import java.util.LinkedHashMap;
import java.util.Map;

import static org.assertj.core.api.Assertions.assertThat;

class ImageTest {

    @Test
    void baseCreatesFromInstruction() {
        assertThat(Image.base("python:3.12").getDockerfile())
                .isEqualTo("FROM python:3.12\n");
    }

    @Test
    void debianSlimUsesDefaultVersion() {
        assertThat(Image.debianSlim(null).getDockerfile())
                .isEqualTo("FROM python:3.11-slim\n");
    }

    @Test
    void debianSlimUsesProvidedVersion() {
        assertThat(Image.debianSlim("3.12").getDockerfile())
                .isEqualTo("FROM python:3.12-slim\n");
    }

    @Test
    void pipInstallIgnoresEmptyPackages() {
        Image image = Image.base("python:3.12").pipInstall();

        assertThat(image.getDockerfile()).isEqualTo("FROM python:3.12\n");
    }

    @Test
    void runCommandsAndEnvIgnoreNullValues() {
        Image image = Image.base("python:3.12")
                .runCommands((String[]) null)
                .env(null);

        assertThat(image.getDockerfile()).isEqualTo("FROM python:3.12\n");
    }

    @Test
    void envEscapesQuotes() {
        Map<String, String> env = new LinkedHashMap<String, String>();
        env.put("NAME", "va\"lue");

        assertThat(Image.base("python:3.12").env(env).getDockerfile())
                .contains("ENV NAME=\"va\\\"lue\"");
    }

    @Test
    void entrypointAndCmdEscapeQuotesAndAllowEmptyArrays() {
        String dockerfile = Image.base("python:3.12")
                .entrypoint("python", "say \"hi\"")
                .cmd((String[]) null)
                .getDockerfile();

        assertThat(dockerfile)
                .contains("ENTRYPOINT [\"python\",\"say \\\"hi\\\"\"]\n")
                .contains("CMD []\n");
    }

    @Test
    void workdirAppendsLiteralValue() {
        assertThat(Image.base("python:3.12").workdir("").getDockerfile())
                .isEqualTo("FROM python:3.12\nWORKDIR \n");
    }

    @Test
    void fluentMutationsAppendDockerfileLines() {
        String dockerfile = Image.base("python:3.12")
                .pipInstall("pytest", "requests")
                .runCommands("apt-get update", "apt-get install -y git")
                .workdir("/workspace")
                .entrypoint("python", "main.py")
                .cmd("--flag")
                .getDockerfile();

        assertThat(dockerfile)
                .contains("RUN pip install pytest requests\n")
                .contains("RUN apt-get update\n")
                .contains("RUN apt-get install -y git\n")
                .contains("WORKDIR /workspace\n")
                .contains("ENTRYPOINT [\"python\",\"main.py\"]\n")
                .contains("CMD [\"--flag\"]\n");
    }
}
